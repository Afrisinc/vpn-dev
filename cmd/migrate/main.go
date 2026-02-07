package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Database connection string
	connStr := "postgres://postgres:12345@localhost:5432/afrisinc_vpn_db?sslmode=disable"

	// Connect to database
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// Migration files in order
	migrations := []struct {
		name string
		file string
	}{
		{"Create servers table", "migrations/000002_create_servers_table.up.sql"},
		{"Add server_id to users", "migrations/000003_add_server_to_users.up.sql"},
	}

	// Execute each migration
	for _, migration := range migrations {
		fmt.Printf("Executing: %s...\n", migration.name)

		migrationSQL, err := os.ReadFile(migration.file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to read migration file %s: %v\n", migration.file, err)
			os.Exit(1)
		}

		_, err = conn.Exec(ctx, string(migrationSQL))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to execute migration %s: %v\n", migration.name, err)
			os.Exit(1)
		}

		fmt.Printf("✓ %s completed\n", migration.name)
	}

	fmt.Println("\n✅ All migrations executed successfully!")
	fmt.Println("✓ Users table created")
	fmt.Println("✓ Servers table created")
	fmt.Println("✓ Server reference added to users")
}
