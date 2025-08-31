package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"

	vaultapi "github.com/hashicorp/vault/api"
	"github.com/retailnext/vault-init/pkgs/aws"
	"github.com/retailnext/vault-init/pkgs/files"
	"github.com/retailnext/vault-init/pkgs/gcp"
	"github.com/retailnext/vault-init/pkgs/objects"
	"github.com/retailnext/vault-init/pkgs/vault"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

type MainClients struct {
	Clients   *objects.Clients
	PostTasks []objects.Task
}

type RawPostTask struct {
	TaskType    string                 `yaml:"type"`
	TaskContent map[string]interface{} `yaml:"task"`
}

func SetupClients(cCtx *cli.Context) (mainClients *MainClients, err error) {
	// get clients
	var vaultClient *vault.Vault
	var ssmClient objects.SSMClient
	var initOutSecret string

	vaultClient, err = vault.NewClientWithContext(cCtx.Context, cCtx.String("vault-addr"), []byte(cCtx.String("cacert")))
	if err != nil {
		return
	}
	switch initOutType := initOutPathType(cCtx.String("initout")); initOutType {
	case "gcp":
		ssmClient, err = gcp.NewSecretClient(cCtx.Context)
		if err != nil {
			return mainClients, err
		}
		initOutSecret = cCtx.String("initout")
	case "aws":
		var name, region string
		name, _, region, err = aws.ParseSecretArn(cCtx.String("initout"))
		if err != nil {
			return mainClients, err
		}
		ssmClient, err = aws.NewSecretClient(cCtx.Context, region)
		if err != nil {
			return mainClients, err
		}
		initOutSecret = name
	case "file":
		ssmClient, err = files.NewLocalFileClient(cCtx.Context)
		if err != nil {
			return mainClients, err
		}
		initOutSecret, err = filepath.Abs(cCtx.String("initout"))
		if err != nil {
			return mainClients, err
		}
	default:
		return mainClients, fmt.Errorf("%s is %s type, which is invalid", cCtx.String("initout"), initOutType)
	}

	objectClients := &objects.Clients{
		VaultClient:   vaultClient,
		SSMClient:     ssmClient,
		InitOutSecret: initOutSecret,
	}
	mainClients = &MainClients{
		Clients:   objectClients,
		PostTasks: []objects.Task{},
	}

	return mainClients, err
}

func (c *MainClients) SetupPostTasks(taskContent []byte) (err error) {
	if len(taskContent) == 0 {
		return err
	}

	// Set post tasks
	var rawTasks []RawPostTask
	err = yaml.Unmarshal(taskContent, &rawTasks)
	if err != nil {
		return err
	}
	for _, rawTask := range rawTasks {
		subTaskContent, err := yaml.Marshal(rawTask.TaskContent)
		if err != nil {
			return err
		}

		switch rawTask.TaskType {
		case "policy":
			postTask := &objects.PolicyTask{}
			err = postTask.Set(c.Clients, subTaskContent)
			if err != nil {
				return err
			}
			c.PostTasks = append(c.PostTasks, postTask)
		case "oidc_auth":
			postTask := &objects.OIDCAuthTask{}
			err = postTask.Set(c.Clients, subTaskContent)
			if err != nil {
				return err
			}
			c.PostTasks = append(c.PostTasks, postTask)
		default:
			return fmt.Errorf("%s is not valid task type", rawTask.TaskType)
		}
	}

	return err
}

func (c *MainClients) InitVault() error {
	// get clients
	vaultClient := c.Clients.VaultClient
	if vaultClient == nil {
		return fmt.Errorf("vault client is not initialized")
	}
	ssmClient := c.Clients.SSMClient
	if ssmClient == nil {
		return fmt.Errorf("secret client is not initialized")
	}
	initOutSecret := c.Clients.InitOutSecret
	if initOutSecret == "" {
		return fmt.Errorf("the path for secret to be retrieved is not known")
	}

	// initialize vault
	isInitialized, err := vaultClient.GetInitStatus()
	if err != nil {
		slog.Error("fail_to_get_status", "error", err)
		return err
	}
	if isInitialized {
		// it is already initialized
		slog.Info("vault is already initialized")
		return c.SetVaultTokenFromSecret()
	}

	slog.Info("vault_init")
	initOutput, err := vaultClient.Initialize()
	if err != nil {
		slog.Error("fail_to_initialize", "error", err)
		return err
	}

	// save the output to secret
	slog.Info("save_vault_init_out")
	_, err = ssmClient.AddVersion(initOutSecret, initOutput)
	if err != nil {
		slog.Error("fail_to_add_to_secret", "error", err)
		slog.Info(string(initOutput)) // so that we can add it manually
	}
	return nil
}

func (c *MainClients) SetVaultTokenFromSecret() error {
	// get clients
	ssmClient := c.Clients.SSMClient
	initOutSecret := c.Clients.InitOutSecret
	vaultClient := c.Clients.VaultClient

	slog.Info("set_vault_token")
	lastInitout, err := ssmClient.GetValue(initOutSecret)
	if err != nil {
		return err
	}
	initResp := &vaultapi.InitResponse{}
	err = json.Unmarshal(lastInitout, initResp)
	if err != nil {
		return err
	}
	if initResp.RootToken == "" {
		return fmt.Errorf("no root token is found from the secret")
	}
	vaultClient.SetToken(initResp.RootToken)
	return nil
}

func (c *MainClients) CheckHealthVault() error {
	// get clients
	vaultClient := c.Clients.VaultClient
	if vaultClient == nil {
		return fmt.Errorf("vault client is not initialized")
	}
	slog.Info("check_health_status")
	healthStatus, err := vaultClient.GetHealthStatus()
	if err != nil {
		return err
	}
	if !healthStatus.Initialized || healthStatus.Sealed {
		return fmt.Errorf("vault_not_ready: initialized = %v, sealed = %v", healthStatus.Initialized, healthStatus.Sealed)
	}
	return nil
}

func (c *MainClients) ExecutePostTasks() error {
	for _, postTask := range c.PostTasks {
		if err := postTask.Do(); err != nil {
			return err
		}
	}
	return nil
}
