package model

import "time"

// Server represents a VPN server node in the infrastructure
type Server struct {
	// Identity
	ID          string `db:"id" json:"id" validate:"required"`
	Name        string `db:"name" json:"name" validate:"required"`
	Location    string `db:"location" json:"location" validate:"required"`
	RegionCode  string `db:"region_code" json:"regionCode" validate:"required"`
	CountryCode string `db:"country_code" json:"countryCode" validate:"required"`

	// Connection info
	PublicIP    string `db:"public_ip" json:"publicIp" validate:"required,ipv4"`
	AgentURL    string `db:"agent_url" json:"agentUrl" validate:"required,url"`
	AgentAPIKey string `db:"agent_api_key" json:"agentApiKey,omitempty"` // Don't expose in JSON
	WGPort      int    `db:"wireguard_port" json:"wireguardPort"`

	// WireGuard config
	ServerPublicKey  string `db:"server_public_key" json:"serverPublicKey" validate:"required"`
	NetworkCIDR      string `db:"network_cidr" json:"networkCidr" validate:"required"`
	NextAvailableIP  int    `db:"next_available_ip" json:"nextAvailableIp"`

	// Capacity & limits
	MaxClients        int `db:"max_clients" json:"maxClients"`
	CurrentClients    int `db:"current_clients" json:"currentClients"`
	BandwidthLimitMbp *int `db:"bandwidth_limit_mbps" json:"bandwidthLimitMbps,omitempty"`

	// Status
	Status       string `db:"status" json:"status" validate:"oneof=active maintenance offline full"`
	HealthStatus string `db:"health_status" json:"healthStatus" validate:"oneof=healthy degraded down unknown"`
	LastHealthCheck *time.Time `db:"last_health_check" json:"lastHealthCheck,omitempty"`

	// Geo info
	Latitude  *float64 `db:"latitude" json:"latitude,omitempty"`
	Longitude *float64 `db:"longitude" json:"longitude,omitempty"`

	// Metadata
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

// IsHealthy checks if server is in a usable state
func (s *Server) IsHealthy() bool {
	return s.Status == "active" && s.HealthStatus == "healthy"
}

// CanAcceptClients checks if server can accept new clients
func (s *Server) CanAcceptClients() bool {
	return s.Status == "active" && s.CurrentClients < s.MaxClients
}

// GetAvailableSlots returns remaining client slots
func (s *Server) GetAvailableSlots() int {
	return s.MaxClients - s.CurrentClients
}
