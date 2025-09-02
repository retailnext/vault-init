package vault

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	testvault "github.com/testcontainers/testcontainers-go/modules/vault"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	VAULT_TOKEN = "root-token"
)

type VaultServer struct {
	*testvault.VaultContainer
}

func (v *VaultServer) Execute(ctx context.Context, command string) ([]byte, error) {
	_, out, err := v.Exec(ctx, strings.Split(command, " "))
	if err != nil {
		return []byte{}, err
	}
	output, err := io.ReadAll(out)
	return output, err

}

func StartTestDevVaultInTest(t *testing.T, ctx context.Context) (vserver *VaultServer, vclient *Vault, err error) {
	vaultDevContainer, err := testvault.Run(ctx,
		"hashicorp/vault:1.13.0",
		testvault.WithToken(VAULT_TOKEN),
		testcontainers.WithWaitStrategy(
			wait.ForLog(fmt.Sprintf("Root Token: %s", VAULT_TOKEN)).WithOccurrence(1).WithStartupTimeout(10*time.Second),
		),
	)
	if err != nil {
		return
	}

	vserver = &VaultServer{vaultDevContainer}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(vaultDevContainer); err != nil {
			t.Fatalf("failed to terminate the vault container: %s", err)
		}
	})
	vaultAddr, err := vserver.HttpHostAddress(ctx)
	if err != nil {
		return
	}
	vclient, err = NewClient(vaultAddr, []byte{})
	if err != nil {
		return
	}
	vclient.SetToken(VAULT_TOKEN)
	return vserver, vclient, nil
}
