package vault

import (
	"context"

	vaultapi "github.com/hashicorp/vault/api"
)

type Vault struct {
	client *vaultapi.Client
	ctx    context.Context
}

func NewClient(vaultaddr string, caCertBytes []byte) (vaultClient *Vault, err error) {
	return NewClientWithContext(context.Background(), vaultaddr, caCertBytes)
}

func NewClientWithContext(ctx context.Context, vaultaddr string, caCertBytes []byte) (vaultClient *Vault, err error) {
	config := vaultapi.DefaultConfig()
	config.Address = vaultaddr
	if len(caCertBytes) > 0 {
		err = config.ConfigureTLS(&vaultapi.TLSConfig{
			CACertBytes: caCertBytes,
		})
		if err != nil {
			return
		}
	}
	client, err := vaultapi.NewClient(config)
	vaultClient = &Vault{
		client: client,
		ctx:    ctx,
	}
	return
}

func (v *Vault) SetToken(token string) {
	v.client.SetToken(token)
}

func (v *Vault) GetToken() string {
	return v.client.Token()
}
