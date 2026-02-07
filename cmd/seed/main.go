package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ElissaDesign/vpn-dev/internal/model"
	"github.com/ElissaDesign/vpn-dev/internal/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Database connection string
	connStr := "postgres://postgres:12345@localhost:5432/afrisinc_vpn_db?sslmode=disable"

	// Connect to database
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Create server repository
	serverRepo := repository.NewServerRepository(pool)

	// Define servers to seed
	servers := []*model.Server{
		{
			ID:              "us-newyork",
			Name:            "New York",
			Location:        "New York, USA",
			RegionCode:      "us-east",
			CountryCode:     "US",
			PublicIP:        "203.0.113.1",
			AgentURL:        "https://wg-agent-nyc.example.com",
			AgentAPIKey:     "secret-key-nyc",
			WGPort:          51820,
			ServerPublicKey: "QGF4I+0L7klYTozF8pFDZUKC4fGt3tEuCakfWYghN0Y=",
			NetworkCIDR:     "192.168.88.0/24",
			MaxClients:      250,
			CurrentClients:  0,
			Status:          "active",
			HealthStatus:    "healthy",
			Latitude:        ptr(40.7128),
			Longitude:       ptr(-74.0060),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "eu-paris",
			Name:            "Paris",
			Location:        "Paris, France",
			RegionCode:      "eu-west",
			CountryCode:     "FR",
			PublicIP:        "203.0.113.2",
			AgentURL:        "https://wg-agent-paris.example.com",
			AgentAPIKey:     "secret-key-paris",
			WGPort:          51820,
			ServerPublicKey: "cWC4I+0L7klYTozF8pFDZUKC4fGt3tEuCakfWYghN0Y=",
			NetworkCIDR:     "192.168.89.0/24",
			MaxClients:      250,
			CurrentClients:  0,
			Status:          "active",
			HealthStatus:    "healthy",
			Latitude:        ptr(48.8566),
			Longitude:       ptr(2.3522),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "sg-singapore",
			Name:            "Singapore",
			Location:        "Singapore",
			RegionCode:      "ap-southeast",
			CountryCode:     "SG",
			PublicIP:        "203.0.113.3",
			AgentURL:        "https://wg-agent-sg.example.com",
			AgentAPIKey:     "secret-key-sg",
			WGPort:          51820,
			ServerPublicKey: "EFG567I+0L7klYTozF8pFDZUKC4fGt3tEuCakfWYghN0Y=",
			NetworkCIDR:     "192.168.90.0/24",
			MaxClients:      250,
			CurrentClients:  0,
			Status:          "active",
			HealthStatus:    "healthy",
			Latitude:        ptr(1.3521),
			Longitude:       ptr(103.8198),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	// Seed servers
	fmt.Println("🌱 Seeding VPN servers...")
	fmt.Println("─────────────────────────────────────────────────────────")

	successCount := 0
	for _, server := range servers {
		err := serverRepo.Create(ctx, server)
		if err != nil {
			fmt.Printf("✗ Failed to create server %s: %v\n", server.ID, err)
			continue
		}
		fmt.Printf("✓ Created server: %s (%s)\n", server.Name, server.ID)
		successCount++
	}

	fmt.Println("─────────────────────────────────────────────────────────")
	fmt.Printf("✅ Seeding complete: %d/%d servers created\n", successCount, len(servers))

	if successCount > 0 {
		fmt.Println("\n📊 Seeded Servers:")
		for _, server := range servers {
			fmt.Printf("  • %s (%s): %s - %s\n", server.Name, server.ID, server.Location, server.Status)
		}
		fmt.Println("\n🔗 You can now manage these servers via:")
		fmt.Println("  GET  /admin/servers           - List all servers")
		fmt.Println("  GET  /admin/servers/healthy   - List healthy servers")
		fmt.Println("  POST /admin/servers           - Create new server")
		fmt.Println("  PUT  /admin/servers/{id}/status - Update server status")
	}
}

// Helper function to create pointers
func ptr(v float64) *float64 {
	return &v
}
