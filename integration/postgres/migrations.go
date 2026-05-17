package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(ctx context.Context, dbURL, sourceURL string) error {
	if sourceURL == "" {
		return errors.New("migrations source URL is empty")
	}
	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	_ = ctx
	return nil
}
