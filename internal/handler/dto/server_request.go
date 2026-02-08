package dto

// CreateServerRequest is the request body for creating a new server
type CreateServerRequest struct {
	ID              string   `json:"id" validate:"required" example:"us-newyork"`
	Name            string   `json:"name" validate:"required" example:"New York"`
	Location        string   `json:"location" validate:"required" example:"New York, USA"`
	RegionCode      string   `json:"regionCode" validate:"required" example:"us-east"`
	CountryCode     string   `json:"countryCode" validate:"required,len=2" example:"US"`
	PublicIP        string   `json:"publicIp" validate:"required,ipv4" example:"192.0.2.1"`
	AgentURL        string   `json:"agentUrl" validate:"required,url" example:"https://wg-agent-nyc.example.com"`
	AgentAPIKey     string   `json:"agentApiKey" validate:"required" example:"secret-key-nyc"`
	WireGuardPort   int      `json:"wireguardPort" validate:"required,min=1,max=65535" example:"51820"`
	ServerPublicKey string   `json:"serverPublicKey" validate:"required" example:"ABC123..."`
	NetworkCIDR     string   `json:"networkCidr" validate:"required" example:"192.168.88.0/24"`
	MaxClients      int      `json:"maxClients" validate:"required,min=1" example:"250"`
	BandwidthLimit  *int     `json:"bandwidthLimitMbps,omitempty" example:"1000"`
	Latitude        *float64 `json:"latitude,omitempty" example:"40.7128"`
	Longitude       *float64 `json:"longitude,omitempty" example:"-74.0060"`
}

// UpdateServerStatusRequest is the request body for updating server status
type UpdateServerStatusRequest struct {
	Status       string `json:"status" validate:"required,oneof=active maintenance offline full" example:"active"`
	HealthStatus string `json:"healthStatus" validate:"required,oneof=healthy degraded down unknown" example:"healthy"`
}

// ServerResponse is the response body for server operations
type ServerResponse struct {
	ID              string   `json:"id" example:"us-newyork"`
	Name            string   `json:"name" example:"New York"`
	Location        string   `json:"location" example:"New York, USA"`
	RegionCode      string   `json:"regionCode" example:"us-east"`
	CountryCode     string   `json:"countryCode" example:"US"`
	PublicIP        string   `json:"publicIp" example:"192.0.2.1"`
	AgentURL        string   `json:"agentUrl" example:"https://wg-agent-nyc.example.com"`
	WireGuardPort   int      `json:"wireguardPort" example:"51820"`
	ServerPublicKey string   `json:"serverPublicKey" example:"ABC123..."`
	NetworkCIDR     string   `json:"networkCidr" example:"192.168.88.0/24"`
	MaxClients      int      `json:"maxClients" example:"250"`
	CurrentClients  int      `json:"currentClients" example:"42"`
	BandwidthLimit  *int     `json:"bandwidthLimitMbps,omitempty" example:"1000"`
	Status          string   `json:"status" example:"active"`
	HealthStatus    string   `json:"healthStatus" example:"healthy"`
	Latitude        *float64 `json:"latitude,omitempty" example:"40.7128"`
	Longitude       *float64 `json:"longitude,omitempty" example:"-74.0060"`
	CreatedAt       string   `json:"createdAt" example:"2025-02-07T10:30:00Z"`
	UpdatedAt       string   `json:"updatedAt" example:"2025-02-07T10:35:00Z"`
}
