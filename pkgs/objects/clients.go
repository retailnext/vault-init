package objects

import "github.com/retailnext/vault-init/pkgs/vault"

// Secret client interface
type SSMClient interface {
	GetValue(string) ([]byte, error)
	AddVersion(string, []byte) (string, error)
}

// Client setup
type Clients struct {
	VaultClient   *vault.Vault
	SSMClient     SSMClient
	InitOutSecret string
}
