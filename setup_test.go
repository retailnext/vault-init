package main

import (
	"context"
	"os"
	"testing"

	"github.com/retailnext/vault-init/pkgs/objects"
	"github.com/retailnext/vault-init/pkgs/vault"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func TestInitVault(t *testing.T) {
	initoutFile := "./testinitout.txt"
	initOut := `
		{"keys":[],"keys_base64":[],"recovery_keys":["35a24","1ac7","ac0f5"],"recovery_keys_base64":["AOk","dK3rH","rPA"],"root_token":"root-token"}
	`
	if err := os.WriteFile(initoutFile, []byte(initOut), 0666); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := os.Remove(initoutFile); err != nil {
			t.Fatal(err)
		}
	}()

	ctx := context.Background()
	vserver, _, err := vault.StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}

	// setup clients
	vaultAddr, err := vserver.HttpHostAddress(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var mainClients *MainClients
	action := func(ctx context.Context, cmd *cli.Command) error {
		var err error
		mainClients, err = SetupClients(context.Background(), cmd)
		return err
	}

	cmd := getVaultInitCliCmd(nil, action)
	if err := cmd.Run(context.Background(), []string{"vault-init", "--vault-addr", vaultAddr, "--initout", initoutFile}); err != nil {
		t.Fatal(err)
	}

	// initVault
	err = mainClients.InitVault()
	if err != nil {
		t.Fatal(err)
	}

	vaultToken := mainClients.Clients.VaultClient.GetToken()
	assert.Equal(t, "root-token", vaultToken)

}

func TestExecutePostTasks(t *testing.T) {
	postTasks := `
- type: policy
  task:
    name: admin
    policy_content: |
      path "sys/leases/*"
      {
        capabilities = ["create", "read", "update", "delete", "list", "sudo"]
      }
`
	expectedPolicy := `path "sys/leases/*"
{
  capabilities = ["create", "read", "update", "delete", "list", "sudo"]
}
`
	ctx := context.Background()
	_, vclient, err := vault.StartTestDevVaultInTest(t, ctx)
	if err != nil {
		t.Fatal(err)
	}
	mainClients := &MainClients{}
	mainClients.Clients = &objects.Clients{
		VaultClient: vclient,
	}
	err = mainClients.SetupPostTasks([]byte(postTasks))
	if err != nil {
		t.Fatal(err)
	}
	err = mainClients.ExecutePostTasks()
	if err != nil {
		t.Fatal(err)
	}

	actualPolicy, err := vclient.GetPolicy("admin")
	if err != nil {
		t.Fatal()
	}
	assert.Equal(t, expectedPolicy, actualPolicy)

}
