package objects

import (
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

type Task interface {
	Do() error
	Set(c *Clients, task []byte) error
}

// === PolicyTask =====

type PolicyTask struct {
	Name          string `yaml:"name"`
	PolicyContent string `yaml:"policy_content"`
	Client        *Clients
}

func (p *PolicyTask) Set(c *Clients, task []byte) (err error) {
	err = yaml.Unmarshal(task, p)
	if err != nil {
		return err
	}
	p.Client = c
	if p.Client.VaultClient == nil {
		return fmt.Errorf("vault client is not initialized")
	}
	return nil
}

func (p *PolicyTask) Do() (err error) {
	slog.Info("add_policy", "name", p.Name)
	return p.Client.VaultClient.SetPolicy(p.Name, p.PolicyContent)
}

// == OIDC jwt auth setup task ===
type OIDCAuthTask struct {
	AuthPath         string      `yaml:"auth_path"`
	OIDCDiscoveryURL string      `yaml:"oidc_discovery_url"`
	BoundIssuer      string      `yaml:"bound_issuer"`
	Role             JWTAuthRole `yaml:"role"`
	Client           *Clients
}

type JWTAuthRole struct {
	Name           string   `yaml:"name"`
	PolicyNames    []string `yaml:"policy_names"`
	BoundAudiences []string `yaml:"bound_audiences"`
	BoundClaimSub  string   `yaml:"bound_claim_sub"`
	UserClaim      string   `yaml:"user_claim"`
	TTL            string   `yaml:"ttl"`
}

func (a *OIDCAuthTask) Set(c *Clients, task []byte) (err error) {
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

func (a *OIDCAuthTask) Do() (err error) {
	slog.Info("enable jwt auth", "path", a.AuthPath)
	authType, err := a.Client.VaultClient.GetAuthType(a.AuthPath)
	if err != nil {
		// try to enable it
		err = a.Client.VaultClient.EnableAuthByPath(a.AuthPath, "jwt")
		if err != nil {
			return err
		}
	} else {
		if authType == "jwt" {
			slog.Info("jwt auth path is already set")
		} else {
			return fmt.Errorf("auth path is already set by something else: auth_type = %s", authType)
		}
	}

	err = a.Client.VaultClient.AddAuthJWTConfig(a.AuthPath, a.OIDCDiscoveryURL, a.BoundIssuer)
	if err != nil {
		return err
	}
	return a.Client.VaultClient.AddAuthRoleJWTType(
		a.AuthPath,
		a.Role.Name,
		a.Role.PolicyNames,
		a.Role.BoundAudiences,
		a.Role.BoundClaimSub,
		a.Role.UserClaim,
		a.Role.TTL,
	)
}

type SecretSyncTask struct {
	MountPath string               `yaml:"mount_path"`
	KVSecret  []SecretSyncKVSecret `yaml:"kv_secret"`
	Client    *Clients
}

type SecretSyncKVSecret struct {
	Name       string `yaml:"name"`
	SecretPath string `yaml:"secret_path"`
}

func (s *SecretSyncTask) Set(c *Clients, task []byte) (err error) {
	err = yaml.Unmarshal(task, s)
	if err != nil {
		return err
	}
	s.Client = c
	if s.Client.VaultClient == nil {
		return fmt.Errorf("vault client is not initialized")
	}
	if s.Client.SSMClient == nil {
		return fmt.Errorf("ssm client is not initialized")
	}
	if strings.TrimSpace(s.MountPath) == "" {
		return fmt.Errorf("mount_path is required")
	}
	if len(s.KVSecret) == 0 {
		return fmt.Errorf("kv_secret must have at least one entry")
	}
	for i, secret := range s.KVSecret {
		if strings.TrimSpace(secret.Name) == "" {
			return fmt.Errorf("kv_secret[%d]: name is required", i)
		}
		if strings.TrimSpace(secret.SecretPath) == "" {
			return fmt.Errorf("kv_secret[%d]: secret_path is required", i)
		}
	}
	return nil
}

func (s *SecretSyncTask) Do() (err error) {
	slog.Info("create kv", "mount", s.MountPath)
	err = s.Client.VaultClient.CreateKVMountIfNotExist(s.MountPath)
	if err != nil {
		return err
	}
	for _, secret := range s.KVSecret {
		slog.Info("write kv secret", "mount", s.MountPath, "secret", secret.Name, "secret_path", secret.SecretPath)
		secretValue, err := s.Client.SSMClient.GetValue(secret.SecretPath)
		if err != nil {
			return err
		}
		if !utf8.Valid(secretValue) {
			return fmt.Errorf("secret %q at path %q contains non-UTF-8 bytes and cannot be written to Vault as a string value", secret.Name, secret.SecretPath)
		}
		err = s.Client.VaultClient.WriteSecret(s.MountPath, secret.Name, string(secretValue))
		if err != nil {
			return err
		}
	}
	return nil
}
