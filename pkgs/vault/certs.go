package vault

import (
	"errors"
	"fmt"
)

var (
	ErrCertIssueFailed = errors.New("failed to issue new certificate from vault")
)

type VaultPKI struct {
	Path    string
	Role    string
	CN      string
	CertTTL string
}

type Certs struct {
	CA          string
	Certificate string
	Key         string
}

func (v *Vault) IssueNewCertificate(vaultpki VaultPKI) (certs Certs, err error) {
	data := map[string]interface{}{
		"common_name": vaultpki.CN,
		"ttl":         vaultpki.CertTTL,
	}
	issuePath := fmt.Sprintf("%s/issue/%s", vaultpki.Path, vaultpki.Role)
	resp, err := v.client.Logical().Write(issuePath, data)
	if err != nil {
		return certs, err
	}
	if resp == nil {
		return certs, ErrCertIssueFailed
	}

	certs = Certs{
		CA:          resp.Data["issuing_ca"].(string),
		Certificate: resp.Data["certificate"].(string),
		Key:         resp.Data["private_key"].(string),
	}
	return
}
