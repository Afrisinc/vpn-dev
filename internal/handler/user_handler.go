package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/ElissaDesign/vpn-dev/internal/config"
	"github.com/ElissaDesign/vpn-dev/internal/model"
	"github.com/ElissaDesign/vpn-dev/internal/service"
	"github.com/ElissaDesign/vpn-dev/internal/utils/response"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var (
	// client          *mongo.Client
	// db              *mongo.Database
	interfaceName    = "wg0"                                          // Interface name for WireGuard
	serverPublicKey  = "Nhc8iTX1IzuZaAF/iGsl5mUuTqiXsjhCHZFehsPyCVM=" // Public key of the server
	serverPrivateKey = "YOUR_SERVER_PRIVATE_KEY_HERE"                 // Private key of the server
	serverEndpoint   = "192.168.1.70:51820"                           // Endpoint of the server
)

type UserHandler struct {
	userService *service.UserService
	config      *config.WireGuardConfig
}

func NewUserHandler(s *service.UserService, cfg *config.WireGuardConfig) *UserHandler {
	return &UserHandler{
		userService: s,
		config:      cfg,
	}
}

func validateEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".") // Validate email format
}

// ensureWireGuardInterface checks if the WireGuard interface exists and creates it if it doesn't
func (h *UserHandler) ensureWireGuardInterface() error {
	// Check if interface exists using Windows commands
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Get-NetAdapter | Where-Object {$_.Name -eq '%s'}", h.config.InterfaceName))
	output, err := cmd.CombinedOutput()
	if err != nil || len(output) == 0 {
		// Interface doesn't exist, create it using WireGuard's Windows commands
		createCmd := exec.Command("wireguard", "/installtunnelservice", h.config.InterfaceName)
		if err := createCmd.Run(); err != nil {
			return fmt.Errorf("failed to create WireGuard interface: %v", err)
		}

		// Wait a moment for the interface to be created
		time.Sleep(2 * time.Second)

		// Set the server's private key
		keyCmd := exec.Command("wireguard", "/setprivatekey", h.config.InterfaceName, h.config.ServerPrivateKey)
		if err := keyCmd.Run(); err != nil {
			return fmt.Errorf("failed to set server private key: %v", err)
		}

		// Set the listen port
		listenCmd := exec.Command("wireguard", "/setlistenport", h.config.InterfaceName, h.config.ListenPort)
		if err := listenCmd.Run(); err != nil {
			return fmt.Errorf("failed to set listen port: %v", err)
		}
	}
	return nil
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetUsers(r.Context()) // Fetch all users from the service
	if err != nil {
		log.Printf("Error fetching users: %v", err) // Log error if fetching fails
		response.Error(w, http.StatusInternalServerError, "Error fetching users", err.Error())
		return
	}
	response.Success(w, "Users fetched successfully", users) // Return success response with users
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Hit POST /users/register")
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Invalid request body: %v", err)
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if !validateEmail(user.Email) || user.IP == "" || user.SubType == "" {
		log.Printf("Missing or invalid fields for user: %v", user)
		http.Error(w, "Missing or invalid fields", http.StatusBadRequest)
		return
	}

	// Ensure WireGuard interface exists
	if err := h.ensureWireGuardInterface(); err != nil {
		log.Printf("Failed to ensure WireGuard interface: %v", err)
		response.Error(w, http.StatusInternalServerError, "Failed to setup WireGuard interface", err.Error())
		return
	}

	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		log.Printf("Key generation failed: %v", err)
		http.Error(w, "Key generation failed", http.StatusInternalServerError)
		return
	}
	privateKey := key.String()
	publicKey := key.PublicKey().String()

	userObj := &model.User{
		Email: user.Email,
		WireGuard: model.WireGuard{
			PrivateKey: privateKey,
			PublicKey:  publicKey,
			IP:         user.IP,
		},
		Subscription: &model.Subscription{
			Type:      user.SubType,
			Status:    "active",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		},
		CreatedAt: time.Now(),
	}

	if err := h.userService.CreateUser(r.Context(), userObj); err != nil {
		log.Printf("Error creating user: %v", err)
		response.Error(w, http.StatusInternalServerError, "Error creating user", err.Error())
		return
	}

	cmd := exec.Command("wg", "set", h.config.InterfaceName, "peer", publicKey, "allowed-ips", user.IP+"/32")
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("WireGuard add peer failed: %v, Output: %s", err, output)
		http.Error(w, "WireGuard add peer failed", http.StatusInternalServerError)
		return
	}

	config := fmt.Sprintf(`[Interface]
     PrivateKey = %s
     Address = %s/24
     
     [Peer]
     PublicKey = %s
     Endpoint = %s
     AllowedIPs = 0.0.0.0/0
     PersistentKeepalive = 25
     `, privateKey, user.IP, h.config.ServerPublicKey, h.config.ServerEndpoint)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"config": config})

	response.Success(w, "User created successfully", userObj)
}

func (h *UserHandler) ToggleUserHandler(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		PublicKey string `json:"publicKey"` // Public key of the user
		Action    string `json:"action"`    // Action to perform: "connect" or "disconnect"
	}
	var body Request
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("Invalid JSON: %v", err) // Log error if JSON is invalid
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if body.PublicKey == "" || (body.Action != "connect" && body.Action != "disconnect") {
		log.Printf("Invalid fields in request: %v", body) // Log error if fields are invalid
		http.Error(w, "Invalid fields", http.StatusBadRequest)
		return
	}

	if body.Action == "disconnect" {
		cmd := exec.Command("wg", "set", h.config.InterfaceName, "peer", body.PublicKey, "remove") // Remove peer
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to disconnect peer: %v", err) // Log error if disconnection fails
			http.Error(w, "Failed to disconnect peer", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Disconnected peer.")) // Return success message
		return
	} else if body.Action == "connect" {
		user, err := h.userService.GetUserByPublicKey(r.Context(), body.PublicKey) // Get user by public key
		if err != nil {
			log.Printf("User not found: %v", err) // Log error if user not found
			response.Error(w, http.StatusNotFound, "User not found", err.Error())
			return
		}

		cmd := exec.Command("wg", "set", h.config.InterfaceName, "peer", user.WireGuard.PublicKey, "allowed-ips", user.WireGuard.IP+"/32") // Connect peer
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to connect peer: %v", err) // Log error if connection fails
			http.Error(w, "Failed to connect peer", http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Peer connected.")) // Return success message
	}
}

func (h *UserHandler) GetUserByPublicKey(w http.ResponseWriter, r *http.Request) {
	publicKey := r.URL.Query().Get("publicKey") // Get public key from query parameter
	if publicKey == "" {
		log.Printf("Missing publicKey query parameter") // Log error if public key is missing
		response.Error(w, http.StatusBadRequest, "Missing publicKey query parameter", "")
		return
	}

	user, err := h.userService.GetUserByPublicKey(r.Context(), publicKey) // Fetch user by public key
	if err != nil {
		log.Printf("User not found: %v", err) // Log error if user not found
		response.Error(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	response.Success(w, "User fetched successfully", user) // Return success response with user
}
