package main

import (
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
	App := &cli.App{
		Name:  "vault-init",
		Usage: "Initialize vault",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "vault-addr", Usage: "Vault address", EnvVars: []string{"VAULT_ADDR"}, Required: true},
			&cli.StringFlag{Name: "cacert", Usage: "CA cert for vault server", EnvVars: []string{"VAULT_CA"}, Required: false},
			&cli.StringFlag{
				Name:     "initout",
				Usage:    "Output destination for the output of vault init",
				EnvVars:  []string{"VAULT_OUT"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "post-init",
				Usage:    "Instruction on what to do after vault init",
				EnvVars:  []string{"POST_INIT"},
				Required: false,
			},
			&cli.BoolFlag{
				Name:     "dry-run",
				Usage:    "Dry run to check the input",
				Required: false,
			},
		},
		Before: validateFlag,
		Action: retryInitVault,
	}
	if err := App.Run(os.Args); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func validateFlag(cCtx *cli.Context) (err error) {
	// Action in cli.StringFlag occurs AFTER beforeAPP so it is not useful

	// Check initout
	if initOutPathType(cCtx.String("initout")) == "" {
		return fmt.Errorf("unrecognizable initout path: %s does not match of the known types", cCtx.String("initout"))
	}

	return err
}

func retryInitVault(cCtx *cli.Context) error {
	clients, err := SetupClients(cCtx)
	if err != nil {
		return err
	}
	if cCtx.String("post-init") != "" {
		err = clients.SetupPostTasks([]byte(cCtx.String("post-init")))
		if err != nil {
			return err
		}
	}
	if cCtx.Bool("dry-run") {
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
