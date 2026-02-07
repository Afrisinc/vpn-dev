package dto

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
	IP    string `json:"ip" validate:"required,ipv4" example:"10.0.0.2"`
}

// RegisterResponse is the response body for successful user registration
type RegisterResponse struct {
	UserID              string `json:"userId" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email               string `json:"email" example:"user@example.com"`
	IP                  string `json:"ip" example:"10.0.0.2"`
	PublicKey           string `json:"publicKey" example:"QGF4I+0L7klYTozF8pFDZUKC4fGt3tEuCakfWYghN0Y="`
	WireGuardConfig     string `json:"wireGuardConfig"`
	CreatedAt           string `json:"createdAt" example:"2025-02-07T08:30:00Z"`
}
