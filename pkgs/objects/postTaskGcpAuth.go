package objects

import (
	"fmt"
	"log/slog"

	"gopkg.in/yaml.v3"
)

type GCPAuthTask struct {
	AuthPath  string             `yaml:"auth_path"`
	RoleBound []GCPAuthRoleBound `yaml:"role_bound"`
	Client    *Clients
}
type GCPAuthRoleBound struct {
	RoleName        string   `yaml:"role_name"`
	ServiceAccounts []string `yaml:"service_accounts"`
	PolicyNames     []string `yaml:"policy_names"`
}

func (a *GCPAuthTask) Set(c *Clients, task []byte) (err error) {
	err = yaml.Unmarshal(task, a)
	if err != nil {
		return err
	}
	a.Client = c
	if a.Client.VaultClient == nil {
		return fmt.Errorf("vault client is not initialized")
	}
	return nil
}

func (a *GCPAuthTask) Do() (err error) {
	slog.Info("enable gcp auth", "path", a.AuthPath)
	authType, err := a.Client.VaultClient.GetAuthType(a.AuthPath)
	if err != nil {
		// try to enable it
		slog.Info("gcp auth path is not set, enabling it")
		err = a.Client.VaultClient.EnableAuthByPath(a.AuthPath, "gcp")
		if err != nil {
			return err
		}
	} else {
		if authType == "gcp" {
			slog.Info("gcp auth path is already set")
		} else {
			return fmt.Errorf("auth path is already set by something else: auth_type = %s", authType)
		}
	}
	slog.Info("role count to be added", "count", len(a.RoleBound))
	for _, role := range a.RoleBound {
		slog.Info("add gcp auth role", "role_name", role.RoleName)
		err = a.Client.VaultClient.AddAuthRoleIAMType(
			a.AuthPath,
			role.RoleName,
			role.PolicyNames,
			role.ServiceAccounts,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
