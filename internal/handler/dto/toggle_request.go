package dto

// ToggleRequest is the request body for toggling VPN connection
type ToggleRequest struct {
	PublicKey string `json:"publicKey" validate:"required,wireguard_key" example:"QGF4I+0L7klYTozF8pFDZUKC4fGt3tEuCakfWYghN0Y="`
	Action    string `json:"action" validate:"required,oneof=connect disconnect" example:"connect"`
}

// ToggleResponse is the response body for toggle operations
type ToggleResponse struct {
	UserID        string  `json:"userId" example:"550e8400-e29b-41d4-a716-446655440000"`
	PublicKey     string  `json:"publicKey" example:"QGF4I+0L7klYTozF8pFDZUKC4fGt3tEuCakfWYghN0Y="`
	Status        string  `json:"status" example:"active"`
	Action        string  `json:"action" example:"connect"`
	UpdatedAt     string  `json:"updatedAt" example:"2025-02-07T08:35:00Z"`
	LastConnected *string `json:"lastConnected,omitempty" example:"2025-02-07T08:35:00Z"`
}
