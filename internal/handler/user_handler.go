package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ElissaDesign/vpn-dev/internal/model"
	"github.com/ElissaDesign/vpn-dev/internal/service"
	"github.com/ElissaDesign/vpn-dev/internal/utils/response"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type UserHandler struct {
	userService *service.UserService
}

func generateKeyPair() (privateKey string, publicKey string, err error) {
	key, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		return "", "", err
	}
	pub := key.PublicKey()
	return key.String(), pub.String(), nil
}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{userService: s}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetUsers(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Error fetching users", err.Error())
		return
	}
	response.Success(w, "User fetched successfully", users)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	priv, pub, err := generateKeyPair()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Key generation error", err.Error())
		return
	}

	ip := "1.1.1.1"
	subType := "free"

	userObj := &model.User{
		Email: user.Email,
		WireGuard: model.WireGuard{
			PrivateKey: priv,
			PublicKey:  pub,
			IP:         ip,
		},
		Subscription: &model.Subscription{
			Type:      subType,
			Status:    "active",
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		},
		CreatedAt: time.Now(),
	}

	log.Println("Object user ============================================================>:\n", userObj)

	if err := h.userService.CreateUser(r.Context(), userObj); err != nil {
		response.Error(w, http.StatusInternalServerError, "Error creating user", err.Error())
		return
	}

	// Print WireGuard config format
	fmt.Println("\n--- WireGuard Config ---")
	fmt.Printf(`[Interface]
PrivateKey = %s
Address = %s/24

[Peer]
PublicKey = <server-public-key>
Endpoint = <server-ip>:51820
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
`, userObj.WireGuard.PrivateKey, userObj.WireGuard.IP)
	response.Success(w, "User created successfully", userObj)
}
