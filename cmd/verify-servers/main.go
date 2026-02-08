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

	// Get all servers
	rows, err := conn.Query(ctx, `
		SELECT id, name, location, region_code, country_code, public_ip::text, status, health_status,
		       current_clients, max_clients, created_at::text
		FROM servers
		ORDER BY created_at DESC
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query error: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	fmt.Println("📊 VPN Servers in Database")
	fmt.Println("═══════════════════════════════════════════════════════════════")

	count := 0
	for rows.Next() {
		var id, name, location, region, country, publicIP, status, healthStatus string
		var currentClients, maxClients int
		var createdAt string

		if err := rows.Scan(&id, &name, &location, &region, &country, &publicIP, &status,
			&healthStatus, &currentClients, &maxClients, &createdAt); err != nil {
			fmt.Fprintf(os.Stderr, "Scan error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\n🌐 %s (%s)\n", name, id)
		fmt.Printf("   📍 Location: %s (%s)\n", location, country)
		fmt.Printf("   🔗 Region: %s\n", region)
		fmt.Printf("   🌍 Public IP: %s\n", publicIP)
		fmt.Printf("   ✅ Status: %s | Health: %s\n", status, healthStatus)
		fmt.Printf("   👥 Clients: %d/%d\n", currentClients, maxClients)
		fmt.Printf("   📅 Created: %s\n", createdAt)

		count++
	}

	fmt.Println("\n═══════════════════════════════════════════════════════════════")
	fmt.Printf("✓ Total servers: %d\n", count)

	// Get server count
	var totalServers int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM servers").Scan(&totalServers)
	if err == nil {
		fmt.Printf("✓ Database count: %d\n", totalServers)
	}
}
