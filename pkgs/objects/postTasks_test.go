package objects

import (
	"context"
	"testing"

	"github.com/retailnext/vault-init/pkgs/vault"
	"github.com/stretchr/testify/assert"
)

func TestOIDCAuthTask(t *testing.T) {
	oidcTaskContent := `auth_path: jwt
oidc_discovery_url: "https://app.terraform.io"
bound_issuer: "https://app.terraform.io"
role:
  name: tfc-agent
  policy_names:
    - admin
  bound_audiences:
    - vault.workload.identity
  bound_claim_sub: "organization:my-org-name:project:my-project-name:workspace:my-workspace-name:run_phase:*"
  user_claim: terraform_full_workspace
  ttl: 20m
`
	expectedJwtAuthRole := map[string]interface{}{
		"allowed_redirect_uris": interface{}(nil),
		"bound_audiences":       []interface{}{"vault.workload.identity"},
		"bound_claims": map[string]interface{}{
			"sub": "organization:my-org-name:project:my-project-name:workspace:my-workspace-name:run_phase:*",
		},
		"bound_claims_type":       "glob",
		"bound_subject":           "",
		"claim_mappings":          interface{}(nil),
		"clock_skew_leeway":       float64(0),
		"expiration_leeway":       float64(0),
		"groups_claim":            "",
		"max_age":                 float64(0),
		"not_before_leeway":       float64(0),
		"oidc_scopes":             interface{}(nil),
		"policies":                []interface{}{"admin"},
		"role_type":               "jwt",
		"token_bound_cidrs":       []interface{}{},
		"token_explicit_max_ttl":  float64(0),
		"token_max_ttl":           float64(0),
		"token_no_default_policy": false,
		"token_num_uses":          float64(0),
		"token_period":            float64(0),
		"token_policies":          []interface{}{"admin"},
		"token_ttl":               float64(1200),
		"token_type":              "default",
		"user_claim":              "terraform_full_workspace",
		"user_claim_json_pointer": false,
		"verbose_oidc_logging":    false,
	}
	ctx := context.Background()
	_, vclient, err := vault.StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}
	oidcAuthTask := &OIDCAuthTask{}
	clients := &Clients{
		VaultClient: vclient,
	}

	if err = oidcAuthTask.Set(clients, []byte(oidcTaskContent)); err != nil {
		t.Fatal(err)
	}
	if err = oidcAuthTask.Do(); err != nil {
		t.Fatal(err)
	}

	actualJwtAuthRole, err := vclient.GetRawAuthRole("jwt", "tfc-agent")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expectedJwtAuthRole, actualJwtAuthRole)
}
