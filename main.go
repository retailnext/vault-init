package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/retailnext/vault-init/pkgs/aws"
	"github.com/retailnext/vault-init/pkgs/gcp"
	"github.com/retailnext/vault-init/pkgs/retry"
	"github.com/urfave/cli/v3"
)

const (
	ADMINPOLICYNAME = "admin"
	VAULTSTATUSFAIL = "fail_to_get_status"
	CLIENTCONTEXT   = "client"
)

func main() {
	App := getVaultInitCliCmd(validateFlag, retryInitVault)
	if err := App.Run(context.Background(), os.Args); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func validateFlag(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	// Action in cli.StringFlag occurs AFTER beforeAPP so it is not useful

	// Check initout
	if initOutPathType(cmd.StringArg("initout")) == "" {
		return ctx, fmt.Errorf("unrecognizable initout path: %s does not match of the known types", cmd.StringArg("initout"))
	}

	return ctx, nil
}

func retryInitVault(ctx context.Context, cmd *cli.Command) error {
	clients, err := SetupClients(ctx, cmd)
	if err != nil {
		return err
	}
	if cmd.StringArg("post-init") != "" {
		err = clients.SetupPostTasks([]byte(cmd.StringArg("post-init")))
		if err != nil {
			return err
		}
	}
	if cmd.Bool("dry-run") {
		fmt.Println("All the clients can be created")
		return nil
	}

	return retry.Do(
		func() error {
			err := clients.InitVault()
			if err != nil {
				return err
			}
			err = clients.CheckHealthVault()
			if err != nil {
				return err
			}

			return clients.ExecutePostTasks()
		},
		retry.Delay(5*time.Second),
		retry.Attempts(10),
	)
}

func initOutPathType(path string) string {
	if gcp.CheckSecretPath(path) {
		return "gcp"
	}
	if aws.CheckSecretPath(path) {
		return "aws"
	}
	if _, err := filepath.Abs(path); err == nil {
		return "file"
	}

	return ""
}

func getVaultInitCliCmd(beforeFunc cli.BeforeFunc, actionFunc cli.ActionFunc) *cli.Command {
	return &cli.Command{
		Name:  "vault-init",
		Usage: "Initialize vault",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "vault-addr", Usage: "Vault address", Sources: cli.EnvVars("VAULT_ADDR"), Required: true},
			&cli.StringFlag{Name: "cacert", Usage: "CA cert for vault server", Sources: cli.EnvVars("VAULT_CA"), Required: false},
			&cli.StringFlag{
				Name:     "initout",
				Usage:    "Output destination for the output of vault init",
				Sources:  cli.EnvVars("VAULT_OUT"),
				Required: true,
			},
			&cli.StringFlag{
				Name:     "post-init",
				Usage:    "Instruction on what to do after vault init",
				Sources:  cli.EnvVars("POST_INIT"),
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "dry-run",
				Usage:    "Dry run to check the input",
				Required: false,
			},
		},
		Before: beforeFunc,
		Action: actionFunc,
	}
}
