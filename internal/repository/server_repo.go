package repository

import (
	"context"
	"fmt"

	"github.com/ElissaDesign/vpn-dev/internal/errors"
	"github.com/ElissaDesign/vpn-dev/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ServerRepository handles database operations for servers
type ServerRepository struct {
	pool *pgxpool.Pool
}

// NewServerRepository creates a new server repository
func NewServerRepository(pool *pgxpool.Pool) *ServerRepository {
	return &ServerRepository{
		pool: pool,
	}
}

// GetAll retrieves all servers from the database
func (r *ServerRepository) GetAll(ctx context.Context) ([]model.Server, error) {
	query := `
		SELECT id, name, location, region_code, country_code, public_ip::text, agent_url, agent_api_key,
		       wireguard_port, server_public_key, network_cidr, next_available_ip, max_clients,
		       current_clients, bandwidth_limit_mbps, status, health_status, last_health_check,
		       latitude, longitude, created_at, updated_at
		FROM servers
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}
	defer rows.Close()

	var servers []model.Server
	for rows.Next() {
		var server model.Server
		if err := rows.Scan(
			&server.ID,
			&server.Name,
			&server.Location,
			&server.RegionCode,
			&server.CountryCode,
			&server.PublicIP,
			&server.AgentURL,
			&server.AgentAPIKey,
			&server.WGPort,
			&server.ServerPublicKey,
			&server.NetworkCIDR,
			&server.NextAvailableIP,
			&server.MaxClients,
			&server.CurrentClients,
			&server.BandwidthLimitMbp,
			&server.Status,
			&server.HealthStatus,
			&server.LastHealthCheck,
			&server.Latitude,
			&server.Longitude,
			&server.CreatedAt,
			&server.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
		}
		servers = append(servers, server)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return servers, nil
}

// GetByID retrieves a server by ID
func (r *ServerRepository) GetByID(ctx context.Context, id string) (*model.Server, error) {
	query := `
		SELECT id, name, location, region_code, country_code, public_ip::text, agent_url, agent_api_key,
		       wireguard_port, server_public_key, network_cidr, next_available_ip, max_clients,
		       current_clients, bandwidth_limit_mbps, status, health_status, last_health_check,
		       latitude, longitude, created_at, updated_at
		FROM servers
		WHERE id = $1
	`

	var server model.Server
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&server.ID,
		&server.Name,
		&server.Location,
		&server.RegionCode,
		&server.CountryCode,
		&server.PublicIP,
		&server.AgentURL,
		&server.AgentAPIKey,
		&server.WGPort,
		&server.ServerPublicKey,
		&server.NetworkCIDR,
		&server.NextAvailableIP,
		&server.MaxClients,
		&server.CurrentClients,
		&server.BandwidthLimitMbp,
		&server.Status,
		&server.HealthStatus,
		&server.LastHealthCheck,
		&server.Latitude,
		&server.Longitude,
		&server.CreatedAt,
		&server.UpdatedAt,
	)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return nil, errors.ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return &server, nil
}

// GetByRegion retrieves all servers in a specific region
func (r *ServerRepository) GetByRegion(ctx context.Context, regionCode string) ([]model.Server, error) {
	query := `
		SELECT id, name, location, region_code, country_code, public_ip::text, agent_url, agent_api_key,
		       wireguard_port, server_public_key, network_cidr, next_available_ip, max_clients,
		       current_clients, bandwidth_limit_mbps, status, health_status, last_health_check,
		       latitude, longitude, created_at, updated_at
		FROM servers
		WHERE region_code = $1 AND status = 'active'
		ORDER BY current_clients ASC
	`

	rows, err := r.pool.Query(ctx, query, regionCode)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}
	defer rows.Close()

	var servers []model.Server
	for rows.Next() {
		var server model.Server
		if err := rows.Scan(
			&server.ID,
			&server.Name,
			&server.Location,
			&server.RegionCode,
			&server.CountryCode,
			&server.PublicIP,
			&server.AgentURL,
			&server.AgentAPIKey,
			&server.WGPort,
			&server.ServerPublicKey,
			&server.NetworkCIDR,
			&server.NextAvailableIP,
			&server.MaxClients,
			&server.CurrentClients,
			&server.BandwidthLimitMbp,
			&server.Status,
			&server.HealthStatus,
			&server.LastHealthCheck,
			&server.Latitude,
			&server.Longitude,
			&server.CreatedAt,
			&server.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
		}
		servers = append(servers, server)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return servers, nil
}

// GetHealthyServers retrieves all healthy active servers with available capacity
func (r *ServerRepository) GetHealthyServers(ctx context.Context) ([]model.Server, error) {
	query := `
		SELECT id, name, location, region_code, country_code, public_ip::text, agent_url, agent_api_key,
		       wireguard_port, server_public_key, network_cidr, next_available_ip, max_clients,
		       current_clients, bandwidth_limit_mbps, status, health_status, last_health_check,
		       latitude, longitude, created_at, updated_at
		FROM servers
		WHERE status = 'active' AND health_status = 'healthy' AND current_clients < max_clients
		ORDER BY current_clients ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}
	defer rows.Close()

	var servers []model.Server
	for rows.Next() {
		var server model.Server
		if err := rows.Scan(
			&server.ID,
			&server.Name,
			&server.Location,
			&server.RegionCode,
			&server.CountryCode,
			&server.PublicIP,
			&server.AgentURL,
			&server.AgentAPIKey,
			&server.WGPort,
			&server.ServerPublicKey,
			&server.NetworkCIDR,
			&server.NextAvailableIP,
			&server.MaxClients,
			&server.CurrentClients,
			&server.BandwidthLimitMbp,
			&server.Status,
			&server.HealthStatus,
			&server.LastHealthCheck,
			&server.Latitude,
			&server.Longitude,
			&server.CreatedAt,
			&server.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
		}
		servers = append(servers, server)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return servers, nil
}

// Create inserts a new server into the database
func (r *ServerRepository) Create(ctx context.Context, server *model.Server) error {
	query := `
		INSERT INTO servers (
			id, name, location, region_code, country_code, public_ip, agent_url, agent_api_key,
			wireguard_port, server_public_key, network_cidr, next_available_ip, max_clients,
			current_clients, bandwidth_limit_mbps, status, health_status, latitude, longitude,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING created_at, updated_at
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		server.ID,
		server.Name,
		server.Location,
		server.RegionCode,
		server.CountryCode,
		server.PublicIP,
		server.AgentURL,
		server.AgentAPIKey,
		server.WGPort,
		server.ServerPublicKey,
		server.NetworkCIDR,
		server.NextAvailableIP,
		server.MaxClients,
		server.CurrentClients,
		server.BandwidthLimitMbp,
		server.Status,
		server.HealthStatus,
		server.Latitude,
		server.Longitude,
		server.CreatedAt,
		server.UpdatedAt,
	).Scan(&server.CreatedAt, &server.UpdatedAt)

	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	return nil
}

// UpdateStatus updates server status and health
func (r *ServerRepository) UpdateStatus(ctx context.Context, id string, status, healthStatus string) error {
	query := `
		UPDATE servers
		SET status = $1, health_status = $2, updated_at = NOW(), last_health_check = NOW()
		WHERE id = $3
	`

	result, err := r.pool.Exec(ctx, query, status, healthStatus, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}

	return nil
}

// IncrementClientCount increments the current client count for a server
func (r *ServerRepository) IncrementClientCount(ctx context.Context, id string) error {
	query := `
		UPDATE servers
		SET current_clients = current_clients + 1, updated_at = NOW()
		WHERE id = $1 AND current_clients < max_clients
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("server at capacity or not found")
	}

	return nil
}

// DecrementClientCount decrements the current client count for a server
func (r *ServerRepository) DecrementClientCount(ctx context.Context, id string) error {
	query := `
		UPDATE servers
		SET current_clients = GREATEST(0, current_clients - 1), updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}

	return nil
}

// UpdateNextAvailableIP updates the next available IP for a server
func (r *ServerRepository) UpdateNextAvailableIP(ctx context.Context, id string, nextIP int) error {
	query := `
		UPDATE servers
		SET next_available_ip = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.pool.Exec(ctx, query, nextIP, id)
	if err != nil {
		return fmt.Errorf("%w: %v", errors.ErrDatabase, err)
	}

	if result.RowsAffected() == 0 {
		return errors.ErrNotFound
	}

	return nil
}
