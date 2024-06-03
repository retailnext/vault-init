package vault

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"testing"

	certutil "github.com/retailnext/vault-init/pkgs/certs"
	"github.com/stretchr/testify/assert"
)

func TestIssueNewCertificate(t *testing.T) {
	ctx := context.Background()
	vserver, vclient, err := StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Setup rootCA
	_, err = vserver.Execute(ctx, "vault secrets enable pki")
	if err != nil {
		t.Fatal(err)
	}
	_, err = vserver.Execute(ctx, "vault secrets tune -max-lease-ttl=87600h pki")
	if err != nil {
		t.Fatal(err)
	}
	execOutput, err := vserver.Execute(ctx, "vault write -field=certificate pki/root/generate/internal common_name=example.com issuer_name=root-2023 ttl=87600h")
	if err != nil {
		t.Fatal(err)
	}
	slog.Info(string(execOutput))
	_, err = vserver.Execute(ctx, "vault write pki/roles/2023-servers allow_any_name=true")
	if err != nil {
		t.Fatal(err)
	}

	commonName := "test.example.com"
	certs, err := vclient.IssueNewCertificate(VaultPKI{
		Path:    "pki",
		Role:    "2023-servers",
		CN:      commonName,
		CertTTL: "24h",
	})

	// verify that all the certs are generated
	assert.NoError(t, err)
	v := reflect.ValueOf(certs)
	typeOfV := v.Type()
	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i).Interface().(string) // better be string
		if value == "" {
			assert.Fail(t, fmt.Sprintf("%s should not be empty", typeOfV.Field(i).Name))
		}
	}

	// verify that common_name is the same
	parsed, err := certutil.ParsePEMCertificate(certs.Certificate)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, commonName, parsed.Subject.CommonName)
	assert.Equal(t, commonName, parsed.DNSNames[0])
}
