package config

// WireGuardConfig holds the configuration for WireGuard
type WireGuardConfig struct {
	InterfaceName    string // Name of the WireGuard interface
	ServerPublicKey  string // Public key of the server
	ServerPrivateKey string // Private key of the server
	ServerEndpoint   string // Endpoint of the server (IP:Port)
	ListenPort       string // Port that WireGuard listens on
}

// NewWireGuardConfig creates a new WireGuard configuration
func NewWireGuardConfig() *WireGuardConfig {
	return &WireGuardConfig{
		InterfaceName:    "wg0",
		ServerPublicKey:  "cWC4I+0L7klYTozF8pFDZUKC4fGt3tEuCakfWYghN0Y=",
		ServerPrivateKey: "uHXTbxAQJ3kO2gXwGui3p+/F1XYhYccXe1jBW/3wEnI=",
		ServerEndpoint:   "192.168.1.71",
		ListenPort:       "51820",
	}
}
