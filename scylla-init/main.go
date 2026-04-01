package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/gocql/gocql"
	"github.com/retailnext/vault-init/pkgs/vault"
	"github.com/urfave/cli/v3"
)

//go:embed init.sql.tmpl
var initSQL string

func main() {
	app := &cli.Command{
		Name:   "scylla-init",
		Usage:  "Initialize ScyllaDB roles using credentials from Vault",
		Flags:  flags(),
		Action: run,
	}
	if err := app.Run(context.Background(), os.Args); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "scylla-host",
			Usage:    "ScyllaDB hostname",
			Sources:  cli.EnvVars("SCYLLA_HOST"),
			Required: true,
		},
		&cli.StringFlag{
			Name:     "vault-addr",
			Usage:    "Vault address",
			Sources:  cli.EnvVars("VAULT_ADDR"),
			Required: true,
		},
		&cli.StringFlag{
			Name:     "vault-token",
			Usage:    "Vault token for authentication",
			Sources:  cli.EnvVars("VAULT_TOKEN"),
			Required: true,
		},
		&cli.StringFlag{
			Name:  "admin-user",
			Usage: "ScyllaDB admin username",
			Value: "cassandra",
		},
		&cli.StringFlag{
			Name:    "cacert",
			Usage:   "CA cert file path for Vault TLS",
			Sources: cli.EnvVars("VAULT_CA"),
		},
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	var caCertBytes []byte
	if cacert := cmd.String("cacert"); cacert != "" {
		var err error
		caCertBytes, err = os.ReadFile(cacert)
		if err != nil {
			return fmt.Errorf("reading cacert %q: %w", cacert, err)
		}
	}

	vaultClient, err := vault.NewClient(cmd.String("vault-addr"), caCertBytes)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}
	vaultClient.SetToken(cmd.String("vault-token"))

	adminPass, err := vaultClient.ReadSecret("adminsecret", "scyllaadmin")
	if err != nil {
		return fmt.Errorf("reading scyllaadmin secret: %w", err)
	}

	vaultUserPass, err := vaultClient.ReadSecret("adminsecret", "vaultuserpass")
	if err != nil {
		return fmt.Errorf("reading vaultuserpass secret: %w", err)
	}

	cluster := gocql.NewCluster(cmd.String("scylla-host"))
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: cmd.String("admin-user"),
		Password: adminPass,
	}
	cluster.Consistency = gocql.Quorum

	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("connecting to scylladb: %w", err)
	}
	defer session.Close()

	cql := strings.ReplaceAll(initSQL, "{{ vaultpass }}", vaultUserPass)

	for _, stmt := range splitStatements(cql) {
		slog.Info("executing", "cql", stmt)
		if err := session.Query(stmt).WithContext(ctx).Exec(); err != nil {
			return fmt.Errorf("executing %q: %w", stmt, err)
		}
	}

	slog.Info("scylladb initialization complete")
	return nil
}

func splitStatements(cql string) []string {
	parts := strings.Split(cql, ";")
	stmts := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			stmts = append(stmts, s)
		}
	}
	return stmts
}
