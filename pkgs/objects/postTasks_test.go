package objects

import (
	"context"
	"fmt"
	"testing"

	"github.com/retailnext/vault-init/pkgs/vault"
	"github.com/stretchr/testify/assert"
)

// mockSSMClient implements SSMClient for testing.
type mockSSMClient struct {
	secrets map[string][]byte
}

func (m *mockSSMClient) GetValue(path string) ([]byte, error) {
	v, ok := m.secrets[path]
	if !ok {
		return nil, fmt.Errorf("secret not found: %s", path)
	}
	return v, nil
}

func (m *mockSSMClient) AddVersion(path string, value []byte) (string, error) {
	if m.secrets == nil {
		m.secrets = make(map[string][]byte)
	}
	m.secrets[path] = value
	return "1", nil
}

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

func TestSecretSyncTask_Set(t *testing.T) {
	taskContent := `mount_path: secret
kv_secret:
  - name: mykey
    secret_path: /ssm/mykey
`
	ctx := context.Background()
	_, vclient, err := vault.StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}
	clients := &Clients{
		VaultClient: vclient,
		SSMClient:   &mockSSMClient{},
	}

	task := &SecretSyncTask{}
	err = task.Set(clients, []byte(taskContent))
	assert.NoError(t, err)
	assert.Equal(t, "secret", task.MountPath)
	assert.Len(t, task.KVSecret, 1)
	assert.Equal(t, "mykey", task.KVSecret[0].Name)
	assert.Equal(t, "/ssm/mykey", task.KVSecret[0].SecretPath)
}

func TestSecretSyncTask_Set_NoVaultClient(t *testing.T) {
	task := &SecretSyncTask{}
	clients := &Clients{SSMClient: &mockSSMClient{}}
	err := task.Set(clients, []byte("mount_path: secret\n"))
	assert.ErrorContains(t, err, "vault client is not initialized")
}

func TestSecretSyncTask_Set_NoSSMClient(t *testing.T) {
	ctx := context.Background()
	_, vclient, err := vault.StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}
	task := &SecretSyncTask{}
	clients := &Clients{VaultClient: vclient}
	err = task.Set(clients, []byte("mount_path: secret\n"))
	assert.ErrorContains(t, err, "ssm client is not initialized")
}

func TestSecretSyncTask_Do(t *testing.T) {
	taskContent := `mount_path: secret
kv_secret:
  - name: mykey
    secret_path: /ssm/mykey
  - name: anotherkey
    secret_path: /ssm/anotherkey
`
	ctx := context.Background()
	_, vclient, err := vault.StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}
	clients := &Clients{
		VaultClient: vclient,
		SSMClient: &mockSSMClient{
			secrets: map[string][]byte{
				"/ssm/mykey":      []byte("myvalue"),
				"/ssm/anotherkey": []byte("anothervalue"),
			},
		},
	}

	task := &SecretSyncTask{}
	if err = task.Set(clients, []byte(taskContent)); err != nil {
		t.Fatal(err)
	}
	err = task.Do()
	assert.NoError(t, err)

	for _, tc := range []struct{ name, want string }{
		{"mykey", "myvalue"},
		{"anotherkey", "anothervalue"},
	} {
		got, err := vclient.ReadSecret("secret", tc.name)
		assert.NoError(t, err)
		assert.Equal(t, tc.want, got)
	}
}

func TestSecretSyncTask_Do_NonUTF8Secret(t *testing.T) {
	taskContent := `mount_path: secret
kv_secret:
  - name: binkey
    secret_path: /ssm/binkey
`
	ctx := context.Background()
	_, vclient, err := vault.StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}
	clients := &Clients{
		VaultClient: vclient,
		SSMClient: &mockSSMClient{
			secrets: map[string][]byte{
				"/ssm/binkey": {0xff, 0xfe, 0x00}, // invalid UTF-8
			},
		},
	}

	task := &SecretSyncTask{}
	if err = task.Set(clients, []byte(taskContent)); err != nil {
		t.Fatal(err)
	}
	err = task.Do()
	assert.ErrorContains(t, err, "non-UTF-8 bytes")
}

func TestSecretSyncTask_Do_SSMError(t *testing.T) {
	taskContent := `mount_path: secret
kv_secret:
  - name: mykey
    secret_path: /ssm/missing
`
	ctx := context.Background()
	_, vclient, err := vault.StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}
	clients := &Clients{
		VaultClient: vclient,
		SSMClient:   &mockSSMClient{secrets: map[string][]byte{}},
	}

	task := &SecretSyncTask{}
	if err = task.Set(clients, []byte(taskContent)); err != nil {
		t.Fatal(err)
	}
	err = task.Do()
	assert.ErrorContains(t, err, "secret not found: /ssm/missing")
}
