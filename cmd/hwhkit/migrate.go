package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

func migrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration management (wraps golang-migrate)",
	}
	cmd.AddCommand(migrateCreateCmd())
	cmd.AddCommand(migrateUpCmd())
	cmd.AddCommand(migrateDownCmd())
	cmd.AddCommand(migrateVersionCmd())
	cmd.AddCommand(migrateForceCmd())
	return cmd
}

func migrateCreateCmd() *cobra.Command {
	var dir string
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new up/down migration pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			ts := time.Now().UTC().Format("20060102150405")
			name := strings.ReplaceAll(args[0], " ", "_")
			up := filepath.Join(dir, ts+"_"+name+".up.sql")
			down := filepath.Join(dir, ts+"_"+name+".down.sql")
			for _, p := range []string{up, down} {
				if err := os.WriteFile(p, []byte("-- "+filepath.Base(p)+"\n"), 0o644); err != nil {
					return err
				}
			}
			fmt.Println(up)
			fmt.Println(down)
			return nil
		},
	}
	cmd.Flags().StringVar(&dir, "dir", "migrations", "migrations directory")
	return cmd
}

func openMigrate(url, dir string) (*migrate.Migrate, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	return migrate.New("file://"+abs, url)
}

func migrateUpCmd() *cobra.Command {
	var url, dir string
	var steps int
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Apply pending migrations (or N steps with --steps)",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := openMigrate(url, dir)
			if err != nil {
				return err
			}
			defer m.Close()
			if steps > 0 {
				return ignoreNoChange(m.Steps(steps))
			}
			return ignoreNoChange(m.Up())
		},
	}
	cmd.Flags().StringVar(&url, "url", os.Getenv("HWHKIT_PG_URL"), "postgres URL (or $HWHKIT_PG_URL)")
	cmd.Flags().StringVar(&dir, "dir", "migrations", "migrations directory")
	cmd.Flags().IntVar(&steps, "steps", 0, "apply N migrations (0 = all pending)")
	return cmd
}

func migrateDownCmd() *cobra.Command {
	var url, dir string
	var steps int
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Revert migrations (defaults to 1 step)",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := openMigrate(url, dir)
			if err != nil {
				return err
			}
			defer m.Close()
			if steps <= 0 {
				steps = 1
			}
			return ignoreNoChange(m.Steps(-steps))
		},
	}
	cmd.Flags().StringVar(&url, "url", os.Getenv("HWHKIT_PG_URL"), "postgres URL")
	cmd.Flags().StringVar(&dir, "dir", "migrations", "migrations directory")
	cmd.Flags().IntVar(&steps, "steps", 1, "revert N migrations")
	return cmd
}

func migrateVersionCmd() *cobra.Command {
	var url, dir string
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print current migration version",
		RunE: func(cmd *cobra.Command, args []string) error {
			m, err := openMigrate(url, dir)
			if err != nil {
				return err
			}
			defer m.Close()
			ver, dirty, err := m.Version()
			if errors.Is(err, migrate.ErrNilVersion) {
				fmt.Println("no migrations applied yet")
				return nil
			}
			if err != nil {
				return err
			}
			fmt.Printf("version=%d dirty=%v\n", ver, dirty)
			return nil
		},
	}
	cmd.Flags().StringVar(&url, "url", os.Getenv("HWHKIT_PG_URL"), "postgres URL")
	cmd.Flags().StringVar(&dir, "dir", "migrations", "migrations directory")
	return cmd
}

func migrateForceCmd() *cobra.Command {
	var url, dir string
	cmd := &cobra.Command{
		Use:   "force <version>",
		Short: "Force the migration version (clears dirty state). Use with care.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ver, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}
			m, err := openMigrate(url, dir)
			if err != nil {
				return err
			}
			defer m.Close()
			return m.Force(ver)
		},
	}
	cmd.Flags().StringVar(&url, "url", os.Getenv("HWHKIT_PG_URL"), "postgres URL")
	cmd.Flags().StringVar(&dir, "dir", "migrations", "migrations directory")
	return cmd
}

func ignoreNoChange(err error) error {
	if err == nil || errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("ok")
		return nil
	}
	return err
}
