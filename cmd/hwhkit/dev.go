package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hwhkit/hwhkit-go/config"
	"github.com/spf13/cobra"
)

func devCmd() *cobra.Command {
	var configDir, env string
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Bring up dependency containers based on detected integrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			loader := config.DefaultLoader()
			cfg, err := loader.Load(context.Background(), &config.BootstrapConfig{ConfigDir: configDir, Env: env})
			if err != nil {
				return err
			}
			svcs := detect(cfg.Config)
			if len(svcs) == 0 {
				fmt.Println("no integrations enabled; nothing to bring up")
				return nil
			}
			fmt.Printf("starting services: %s\n", strings.Join(svcs, ", "))
			return runCompose(append([]string{"compose", "up", "-d"}, svcs...))
		},
	}
	cmd.Flags().StringVar(&configDir, "config-dir", "config", "config directory")
	cmd.Flags().StringVar(&env, "env", "dev", "environment")
	return cmd
}

func detect(cfg *config.AppConfig) []string {
	var out []string
	if cfg.Integrations.SQL.Postgres.Enabled {
		out = append(out, "postgres")
	}
	if cfg.Integrations.Redis.Enabled {
		out = append(out, "redis")
	}
	if cfg.Integrations.MongoDB.Enabled {
		out = append(out, "mongodb")
	}
	if cfg.Integrations.Messaging.NATS.Enabled {
		out = append(out, "nats")
	}
	if cfg.Integrations.Vector.Qdrant.Enabled {
		out = append(out, "qdrant")
	}
	if cfg.Integrations.Neo4j.Enabled {
		out = append(out, "neo4j")
	}
	if cfg.Integrations.Storage.S3.Enabled {
		out = append(out, "minio")
	}
	if cfg.Observability.OTel.Enabled {
		out = append(out, "jaeger")
	}
	return out
}

func runCompose(args []string) error {
	c := exec.Command("docker", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}
