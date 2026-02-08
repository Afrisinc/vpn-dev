package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	connStr := "postgres://postgres:12345@localhost:5432/afrisinc_vpn_db?sslmode=disable"

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// Check table exists
	var exists bool
	err = conn.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_name = 'users'
		)
	`).Scan(&exists)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query error: %v\n", err)
		os.Exit(1)
	}

	if exists {
		fmt.Println("✓ Users table exists")
	} else {
		fmt.Println("✗ Users table NOT found")
		os.Exit(1)
	}

	// Get table structure
	rows, err := conn.Query(ctx, `
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_name = 'users'
		ORDER BY ordinal_position
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query error: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("\n📋 Table Structure:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	fmt.Printf("%-20s %-15s %-10s %-30s\n", "Column", "Type", "Nullable", "Default")
	fmt.Println("─────────────────────────────────────────────────────────────")

	for rows.Next() {
		var colName, dataType string
		var isNullable string
		var colDefault *string

		err := rows.Scan(&colName, &dataType, &isNullable, &colDefault)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
			os.Exit(1)
		}

		defaultVal := "—"
		if colDefault != nil {
			defaultVal = *colDefault
		}

		fmt.Printf("%-20s %-15s %-10s %-30s\n", colName, dataType, isNullable, defaultVal)
	}

	// Get indexes
	fmt.Println("\n🔑 Indexes:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	indexRows, err := conn.Query(ctx, `
		SELECT indexname, indexdef
		FROM pg_indexes
		WHERE tablename = 'users'
		ORDER BY indexname
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Index query error: %v\n", err)
		os.Exit(1)
	}
	defer indexRows.Close()

	for indexRows.Next() {
		var indexName, indexDef string
		if err := indexRows.Scan(&indexName, &indexDef); err != nil {
			fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  • %s\n", indexName)
	}

	// Get constraints
	fmt.Println("\n⚠️  Constraints:")
	fmt.Println("─────────────────────────────────────────────────────────────")
	constraintRows, err := conn.Query(ctx, `
		SELECT constraint_name, constraint_type
		FROM information_schema.table_constraints
		WHERE table_name = 'users'
		ORDER BY constraint_name
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Constraint query error: %v\n", err)
		os.Exit(1)
	}
	defer constraintRows.Close()

	for constraintRows.Next() {
		var constraintName, constraintType string
		if err := constraintRows.Scan(&constraintName, &constraintType); err != nil {
			fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  • %s (%s)\n", constraintName, constraintType)
	}

	// Count rows
	var rowCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&rowCount)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Count query error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Database setup complete!")
	fmt.Printf("✓ Users table is ready with %d rows\n", rowCount)
}
