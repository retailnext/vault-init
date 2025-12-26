package objects

import (
	"context"
	"testing"

	"github.com/retailnext/vault-init/pkgs/vault"
	"github.com/stretchr/testify/assert"
)

func TestGCPAuthTask(t *testing.T) {
	gcpAuthTaskContent := `auth_path: gcp
role_bound:
 - role_name: tf-apply
   service_accounts:
    - "tf-apply@sandbox-account.google.com"
   policy_names:
    - admin
`
	expectedGcpAuthRole := map[string]interface{}{
		"add_group_aliases":       false,
		"allow_gce_inference":     true,
		"bound_service_accounts":  []interface{}{"tf-apply@sandbox-account.google.com"},
		"max_jwt_exp":             float64(3600),
		"policies":                []interface{}{"admin"},
		"role_id":                 "role_id_set",
		"token_bound_cidrs":       []interface{}{},
		"token_explicit_max_ttl":  float64(0),
		"token_max_ttl":           float64(0),
		"token_no_default_policy": false,
		"token_num_uses":          float64(0),
		"token_period":            float64(0),
		"token_policies":          []interface{}{"admin"},
		"token_ttl":               float64(0),
		"token_type":              "default",
		"type":                    "iam",
	}
	ctx := context.Background()
	_, vclient, err := vault.StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}
	gcpAuthTask := &GCPAuthTask{}
	clients := &Clients{
		VaultClient: vclient,
	}

	if err = gcpAuthTask.Set(clients, []byte(gcpAuthTaskContent)); err != nil {
		t.Fatal(err)
	}
	if err = gcpAuthTask.Do(); err != nil {
		t.Fatal(err)
	}

	actualGcpAuthRole, err := vclient.GetRawAuthRole("gcp", "tf-apply")
	if err != nil {
		t.Fatal(err)
	}
	roleId, ok := actualGcpAuthRole["role_id"].(string)
	if ok || roleId != "" {
		actualGcpAuthRole["role_id"] = "role_id_set"
	}

	assert.Equal(t, expectedGcpAuthRole, actualGcpAuthRole)
}
