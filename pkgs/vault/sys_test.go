package vault

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetSAToAdmin(t *testing.T) {
	ctx := context.Background()
	testPolicyName := "admin"
	testPolicy := `
path "sys/leases/*"
{
	capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "auth/*"
{
	capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
`
	_, vclient, err := StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Create a policy
	err = vclient.SetPolicy(testPolicyName, testPolicy)
	assert.NoError(t, err)

	actualPolicy, err := vclient.GetPolicy(testPolicyName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, testPolicy, actualPolicy)

	// Enable gcp
	err = vclient.EnableAuth("gcp")
	if err != nil {
		t.Fatal(err)
	}

	// Add sa@test.com service account to admin role under gcp auth path
	err = vclient.AddAuthRoleIAMType("gcp", "admin", []string{"admin"}, []string{"sa@test.com"})
	if err != nil {
		t.Fatal(err)
	}

	// Get the role
	roleData, err := vclient.GetAuthRole("gcp", "admin")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, []string{"admin"}, roleData.Policies)
	assert.Equal(t, []string{"sa@test.com"}, roleData.BoundServiceAccounts)
	assert.Equal(t, 3600, roleData.MaxJWTExp)
}

func TestInitializeAlreadyInit(t *testing.T) {
	ctx := context.Background()
	_, vclient, err := StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := vclient.Initialize()
	assert.NoError(t, err)
	respString := string(resp)
	assert.Equal(t, "", respString)
}

func TestAddRoleJTWType(t *testing.T) {
	ctx := context.Background()
	testPolicyName := "admin"
	testPolicy := `
path "sys/leases/*"
{
	capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
path "auth/*"
{
	capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
`
	_, vclient, err := StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Create a policy
	err = vclient.SetPolicy(testPolicyName, testPolicy)
	assert.NoError(t, err)

	// Enable gcp
	err = vclient.EnableAuth("jwt")
	if err != nil {
		t.Fatal(err)
	}

	// Add sub to admin role under jwt auth path
	roleName := "tfc-role"
	sub := "organization:my-org-name:project:my-project-name:workspace:my-workspace-name:run_phase:*"
	boundAudience := "vault.workload.identity"
	userClaim := "terraform_full_workspace"
	err = vclient.AddAuthRoleJWTType("jwt", roleName, []string{testPolicy}, []string{boundAudience}, sub, userClaim, "20m")
	if err != nil {
		t.Fatal(err)
	}

	// Get the role
	roleData, err := vclient.GetRawAuthRole("jwt", roleName)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, boundAudience, roleData["bound_audiences"].([]interface{})[0].(string))
	assert.Equal(t, sub, roleData["bound_claims"].(map[string]interface{})["sub"].(string))
	assert.Equal(t, "glob", roleData["bound_claims_type"].(string))
	assert.Equal(t, strings.TrimSpace(testPolicy), roleData["policies"].([]interface{})[0].(string))
}

func TestGetAuthTypeNonExist(t *testing.T) {
	ctx := context.Background()
	_, vclient, err := StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = vclient.GetAuthType("jwt")
	if err == nil {
		t.Error("it expected an error but no error is returned")
		t.Fail()
	} else {
		assert.ErrorContains(t, err, "No auth engine at jwt")
	}
}

func TestGetAuthType(t *testing.T) {
	ctx := context.Background()
	_, vclient, err := StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}
	err = vclient.EnableAuth("jwt")
	if err != nil {
		t.Fatal(err)
	}
	auth, err := vclient.GetAuthType("jwt")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, "jwt", auth)
}
