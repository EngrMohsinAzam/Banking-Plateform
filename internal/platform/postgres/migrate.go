package postgres

import (
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/mohsinazam/banking/migrations"
)

// RunMigrations applies all pending SQL migrations.
func RunMigrations(dsn string) error {
	source, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("migration init: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up: %w", err)
	}
	return nil
}

// MustRunMigrations panics if migrations fail (tests only).
func MustRunMigrations(dsn string) {
	if err := RunMigrations(dsn); err != nil {
		panic(err)
	}
}

// ResetMigrations rolls back all migrations. Used in tests.
func ResetMigrations(dsn string) error {
	source, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("migration init: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration down: %w", err)
	}
	return nil
}

// ResetAllMigrations rolls back every applied migration. Used in tests.
func ResetAllMigrations(dsn string) error {
	source, err := iofs.New(migrations.Files, ".")
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dsn)
	if err != nil {
		return fmt.Errorf("migration init: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	for {
		if err := m.Down(); err != nil {
			if err == migrate.ErrNoChange {
				return nil
			}
			return fmt.Errorf("migration down: %w", err)
		}
	}
}

// ListMigrationFiles returns embedded migration filenames (tests/diagnostics).
func ListMigrationFiles() ([]string, error) {
	return fs.Glob(migrations.Files, "*.sql")
}
