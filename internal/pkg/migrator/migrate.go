package migrator

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/tern/v2/migrate"
)

const VERSION_TABLE = "public.schema_version"

func loadConfigAndDbConn(ctx context.Context, connectionURL string) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, connectionURL)
	if err != nil {
		return nil, fmt.Errorf("database error: %w ", err)
	}
	return conn, nil
}

func StartMigrate(ctx context.Context, connectionURL string, migrations embed.FS) error {
	conn, err := loadConfigAndDbConn(ctx, connectionURL)

	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer conn.Close(ctx)

	fmt.Printf("Starting migration\n")
	m, err := migrate.NewMigrator(ctx, conn, VERSION_TABLE)
	if err != nil {
		return fmt.Errorf("error creating migrator: %v", err)
	}

	migrationRoot, _ := fs.Sub(migrations, "migrations")

	err = m.LoadMigrations(migrationRoot)
	if err != nil {
		return fmt.Errorf("error loading migrations: %v", err)
	}

	fmt.Printf("Loaded %d migrations\n", len(m.Migrations))

	if len(m.Migrations) == 0 {
		return fmt.Errorf("no migrations found")
	}

	m.OnStart = func(sequence int32, name, direction, sql string) {
		fmt.Printf("Executing %s %s\n", direction, name)
	}

	err = m.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("error migrating: %v", err)
	}

	return nil
}
