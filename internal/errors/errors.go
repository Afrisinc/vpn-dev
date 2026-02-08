package errors

import "errors"

// Custom error types for the VPN service
var (
	ErrNotFound           = errors.New("resource not found")
	ErrDuplicateEmail     = errors.New("email already exists")
	ErrDuplicateIP        = errors.New("IP address already allocated")
	ErrDuplicatePublicKey = errors.New("public key already exists")
	ErrInvalidInput       = errors.New("invalid input")
	ErrDatabase           = errors.New("database error")
	ErrValidation         = errors.New("validation error")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrWireGuardFailed    = errors.New("wireguard operation failed")
)
