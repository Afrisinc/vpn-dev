package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ElissaDesign/vpn-dev/internal/config"
	apierrors "github.com/ElissaDesign/vpn-dev/internal/errors"
	"github.com/ElissaDesign/vpn-dev/internal/handler/dto"
	"github.com/ElissaDesign/vpn-dev/internal/middleware"
	"github.com/ElissaDesign/vpn-dev/internal/model"
	"github.com/ElissaDesign/vpn-dev/internal/repository"
	"github.com/ElissaDesign/vpn-dev/internal/service"
	"github.com/ElissaDesign/vpn-dev/internal/utils/response"
	"github.com/ElissaDesign/vpn-dev/internal/validator"
	"github.com/rs/zerolog"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	service    *service.UserService
	serverRepo *repository.ServerRepository
	wgConfig   *config.WireGuardConfig
	validator  *validator.Validator
	logger     *zerolog.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(svc *service.UserService, serverRepo *repository.ServerRepository, cfg *config.WireGuardConfig, val *validator.Validator, logger *zerolog.Logger) *UserHandler {
	return &UserHandler{
		service:    svc,
		serverRepo: serverRepo,
		wgConfig:   cfg,
		validator:  val,
		logger:     logger,
	}
}

// GetUsers retrieves all registered VPN users
// @Summary Get all VPN users
// @Description Retrieve a list of all registered VPN users with their WireGuard configuration
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} response.JSONResponse{data=[]model.User} "List of all users"
// @Failure 500 {object} response.JSONResponse "Internal server error"
// @Router /users [get]
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)

	users, err := h.service.GetUsers(ctx)
	if err != nil {
		h.logger.Error().
			Str("requestId", requestID).
			Err(err).
			Msg("Failed to fetch users")
		response.InternalError(w, "Failed to fetch users", "database_error")
		return
	}

	h.logger.Info().
		Str("requestId", requestID).
		Int("count", len(users)).
		Msg("Successfully fetched users")

	response.OK(w, "Users fetched successfully", users)
}

// CreateUser registers a new VPN user - Production Ready Endpoint
// @Summary Register a new VPN user
// @Description Register a new VPN user with email and IP address. Automatically generates WireGuard public/private key pair and returns ready-to-use configuration
// @Tags Users
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "User registration request with email and IP"
// @Success 201 {object} response.JSONResponse{data=dto.RegisterResponse} "User created successfully with WireGuard configuration"
// @Failure 400 {object} response.JSONResponse "Invalid input - invalid email format or IP address format"
// @Failure 409 {object} response.JSONResponse "Conflict - email or IP address already registered"
// @Failure 500 {object} response.JSONResponse "Internal server error - database or WireGuard operation failed"
// @Router /users/register [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)

	// Parse request
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn().
			Str("requestId", requestID).
			Err(err).
			Msg("Invalid request body")
		response.BadRequest(w, "Invalid request body", map[string]string{
			"body": "Request body is not valid JSON",
		})
		return
	}

	// Validate request
	validationErrs := h.validator.Validate(req)
	if len(validationErrs) > 0 {
		h.logger.Warn().
			Str("requestId", requestID).
			Interface("errors", validationErrs).
			Msg("Validation errors")

		fieldErrors := make(map[string]string)
		for _, ve := range validationErrs {
			fieldErrors[ve.Field] = ve.Message
		}
		response.BadRequest(w, "Validation failed", fieldErrors)
		return
	}

	// Check email uniqueness
	_, err := h.service.GetUserByEmail(ctx, req.Email)
	if err == nil {
		h.logger.Warn().
			Str("requestId", requestID).
			Str("email", req.Email).
			Msg("Email already exists")
		response.Conflict(w, "Email address is already registered")
		return
	}

	// Check IP uniqueness
	_, err = h.service.GetUserByIP(ctx, req.IP)
	if err == nil {
		h.logger.Warn().
			Str("requestId", requestID).
			Str("ip", req.IP).
			Msg("IP already allocated")
		response.Conflict(w, "IP address is already allocated")
		return
	}

	// Get healthy servers for load balancing
	healthyServers, err := h.serverRepo.GetHealthyServers(ctx)
	if err != nil || len(healthyServers) == 0 {
		h.logger.Error().
			Str("requestId", requestID).
			Err(err).
			Msg("No healthy servers available")
		response.InternalError(w, "No healthy VPN servers available at this time", "no_servers")
		return
	}

	// Select first server (lowest load - already sorted by GetHealthyServers)
	selectedServer := healthyServers[0]

	// Generate WireGuard keys
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		h.logger.Error().
			Str("requestId", requestID).
			Err(err).
			Msg("Failed to generate WireGuard keys")
		response.InternalError(w, "Failed to generate VPN keys", "key_generation_error")
		return
	}

	pubKey := privateKey.PublicKey()
	privKeyStr := privateKey.String()
	pubKeyStr := pubKey.String()

	// Create user object with server assignment
	user := &model.User{
		Email:      req.Email,
		IP:         req.IP,
		PrivateKey: privKeyStr,
		PublicKey:  pubKeyStr,
		Status:     "disconnected",
		ServerID:   &selectedServer.ID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Create user in database
	if err := h.service.CreateUser(ctx, user); err != nil {
		if err == apierrors.ErrDuplicateEmail {
			response.Conflict(w, "Email already exists")
			return
		}
		if err == apierrors.ErrDuplicateIP {
			response.Conflict(w, "IP already allocated")
			return
		}

		h.logger.Error().
			Str("requestId", requestID).
			Err(err).
			Msg("Failed to create user in database")
		response.InternalError(w, "Failed to create user", "database_error")
		return
	}

	// Add peer to WireGuard via agent
	h.logger.Info().
		Str("requestId", requestID).
		Str("userId", user.ID.String()).
		Str("serverId", selectedServer.ID).
		Msg("Adding peer to WireGuard server")

	wgAgent := service.NewWireGuardAgentClient(h.logger)
	err = wgAgent.AddPeer(ctx,
		selectedServer.AgentURL,
		selectedServer.AgentAPIKey,
		user.PublicKey,
		user.IP+"/32",
		"") // endpoint can be empty initially

	if err != nil {
		h.logger.Error().
			Str("requestId", requestID).
			Str("userId", user.ID.String()).
			Str("serverId", selectedServer.ID).
			Err(err).
			Msg("Failed to add peer to WireGuard - rolling back user creation")

		// Rollback: delete user from database
		if delErr := h.service.DeleteUser(ctx, user.ID); delErr != nil {
			h.logger.Error().
				Str("requestId", requestID).
				Str("userId", user.ID.String()).
				Err(delErr).
				Msg("Failed to rollback user creation")
		}

		response.InternalError(w, "Failed to configure WireGuard", "wg_error")
		return
	}

	// Generate WireGuard configuration for client
	wgConfig := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25`, privKeyStr, req.IP, h.wgConfig.ServerPublicKey, h.wgConfig.ServerEndpoint)

	h.logger.Info().
		Str("requestId", requestID).
		Str("email", req.Email).
		Str("ip", req.IP).
		Str("userId", user.ID.String()).
		Str("serverId", selectedServer.ID).
		Msg("User registered and peer added to WireGuard successfully")

	// Return response
	resp := dto.RegisterResponse{
		UserID:              user.ID.String(),
		Email:               user.Email,
		IP:                  user.IP,
		PublicKey:           user.PublicKey,
		WireGuardConfig:     wgConfig,
		CreatedAt:           user.CreatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Location", "/users/"+user.ID.String())
	response.Created(w, "User registered successfully", resp)
}

// ToggleUserHandler toggles a user's VPN connection - Production Ready Endpoint
// @Summary Toggle user VPN connection status
// @Description Connect or disconnect a VPN user by public key. Supports idempotent operations - calling with the same state returns success
// @Tags Users
// @Accept json
// @Produce json
// @Param request body dto.ToggleRequest true "Toggle request with public key and action (connect/disconnect)"
// @Success 200 {object} response.JSONResponse{data=dto.ToggleResponse} "Connection toggled successfully with updated user state"
// @Failure 400 {object} response.JSONResponse "Invalid input - invalid public key format or invalid action value"
// @Failure 404 {object} response.JSONResponse "Not found - user with given public key does not exist"
// @Failure 500 {object} response.JSONResponse "Internal server error - database or WireGuard operation failed"
// @Router /users/toggle [post]
func (h *UserHandler) ToggleUserHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)

	// Parse request
	var req dto.ToggleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn().
			Str("requestId", requestID).
			Err(err).
			Msg("Invalid toggle request body")
		response.BadRequest(w, "Invalid request body", map[string]string{
			"body": "Request body is not valid JSON",
		})
		return
	}

	// Validate request
	validationErrs := h.validator.Validate(req)
	if len(validationErrs) > 0 {
		h.logger.Warn().
			Str("requestId", requestID).
			Interface("errors", validationErrs).
			Msg("Toggle validation errors")

		fieldErrors := make(map[string]string)
		for _, ve := range validationErrs {
			fieldErrors[ve.Field] = ve.Message
		}
		response.BadRequest(w, "Validation failed", fieldErrors)
		return
	}

	// Find user by public key
	user, err := h.service.GetUserByPublicKey(ctx, req.PublicKey)
	if err == apierrors.ErrNotFound {
		h.logger.Warn().
			Str("requestId", requestID).
			Str("publicKey", req.PublicKey).
			Msg("User not found")
		response.NotFound(w, "User not found")
		return
	}
	if err != nil {
		h.logger.Error().
			Str("requestId", requestID).
			Err(err).
			Msg("Failed to lookup user")
		response.InternalError(w, "Failed to lookup user", "database_error")
		return
	}

	// Check if state change is needed (idempotency)
	newStatus := "active"
	if req.Action == "disconnect" {
		newStatus = "disconnected"
	}

	if user.Status == newStatus {
		// No-op, already in desired state (idempotent)
		h.logger.Info().
			Str("requestId", requestID).
			Str("userId", user.ID.String()).
			Str("action", req.Action).
			Str("status", newStatus).
			Msg("User already in desired state (idempotent)")

		returnToggleResponse(w, user, req.Action)
		return
	}

	// Get user's server assignment
	if user.ServerID == nil {
		h.logger.Error().
			Str("requestId", requestID).
			Str("userId", user.ID.String()).
			Msg("User has no server assignment")
		response.InternalError(w, "User has no server assignment", "server_error")
		return
	}

	server, err := h.serverRepo.GetByID(ctx, *user.ServerID)
	if err != nil {
		h.logger.Error().
			Str("requestId", requestID).
			Str("serverId", *user.ServerID).
			Err(err).
			Msg("Failed to get user's server")
		response.InternalError(w, "Server not found", "server_error")
		return
	}

	// Call WireGuard agent to update peer status
	h.logger.Info().
		Str("requestId", requestID).
		Str("userId", user.ID.String()).
		Str("serverId", server.ID).
		Str("action", req.Action).
		Msg("Updating peer on WireGuard server")

	wgAgent := service.NewWireGuardAgentClient(h.logger)

	if req.Action == "connect" {
		err = wgAgent.AddPeer(ctx,
			server.AgentURL,
			server.AgentAPIKey,
			user.PublicKey,
			user.IP+"/32",
			"")
	} else {
		err = wgAgent.RemovePeer(ctx,
			server.AgentURL,
			server.AgentAPIKey,
			user.PublicKey)
	}

	if err != nil {
		h.logger.Error().
			Str("requestId", requestID).
			Str("userId", user.ID.String()).
			Str("serverId", server.ID).
			Err(err).
			Str("action", req.Action).
			Msg("Failed to update peer on WireGuard")
		response.InternalError(w, "Failed to update WireGuard connection", "wg_error")
		return
	}

	// Update user status in database
	if err := h.service.UpdateUserStatus(ctx, user.ID, newStatus); err != nil {
		h.logger.Error().
			Str("requestId", requestID).
			Err(err).
			Str("userId", user.ID.String()).
			Msg("Failed to update user status in database")
		response.InternalError(w, "Failed to update connection status", "database_error")
		return
	}

	// Update last connected if action is connect
	if req.Action == "connect" {
		if err := h.service.UpdateLastConnected(ctx, user.ID); err != nil {
			h.logger.Warn().
				Str("requestId", requestID).
				Err(err).
				Str("userId", user.ID.String()).
				Msg("Failed to update last connected time")
			// Don't fail the request for this non-critical operation
		}
	}

	// Update user object for response
	user.Status = newStatus
	user.UpdatedAt = time.Now()

	h.logger.Info().
		Str("requestId", requestID).
		Str("userId", user.ID.String()).
		Str("serverId", server.ID).
		Str("action", req.Action).
		Str("status", newStatus).
		Msg("Peer updated on WireGuard successfully")

	returnToggleResponse(w, user, req.Action)
}

// GetUserByPublicKey retrieves a user by public key
// @Summary Get user by WireGuard public key
// @Description Retrieve a specific user's information using their WireGuard public key
// @Tags Users
// @Accept json
// @Produce json
// @Param publicKey query string true "WireGuard public key (44 characters base64)"
// @Success 200 {object} response.JSONResponse{data=model.User} "User found and returned with full details"
// @Failure 400 {object} response.JSONResponse "Bad request - missing publicKey query parameter"
// @Failure 404 {object} response.JSONResponse "Not found - no user exists with given public key"
// @Failure 500 {object} response.JSONResponse "Internal server error - database operation failed"
// @Router /users/by-public-key [get]
func (h *UserHandler) GetUserByPublicKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)

	publicKey := r.URL.Query().Get("publicKey")
	if publicKey == "" {
		h.logger.Warn().
			Str("requestId", requestID).
			Msg("Missing publicKey query parameter")
		response.BadRequest(w, "Missing publicKey query parameter", map[string]string{
			"publicKey": "publicKey is required",
		})
		return
	}

	user, err := h.service.GetUserByPublicKey(ctx, publicKey)
	if err == apierrors.ErrNotFound {
		h.logger.Warn().
			Str("requestId", requestID).
			Str("publicKey", publicKey).
			Msg("User not found")
		response.NotFound(w, "User not found")
		return
	}
	if err != nil {
		h.logger.Error().
			Str("requestId", requestID).
			Err(err).
			Msg("Failed to lookup user")
		response.InternalError(w, "Failed to lookup user", "database_error")
		return
	}

	h.logger.Info().
		Str("requestId", requestID).
		Str("userId", user.ID.String()).
		Msg("User fetched successfully")

	response.OK(w, "User fetched successfully", user)
}

// Helper function to return toggle response
func returnToggleResponse(w http.ResponseWriter, user *model.User, action string) {
	resp := dto.ToggleResponse{
		UserID:        user.ID.String(),
		PublicKey:     user.PublicKey,
		Status:        user.Status,
		Action:        action,
		UpdatedAt:     user.UpdatedAt.Format(time.RFC3339),
		LastConnected: nil,
	}

	if user.LastConnected != nil {
		lastConnected := user.LastConnected.Format(time.RFC3339)
		resp.LastConnected = &lastConnected
	}

	response.OK(w, "Connection toggled successfully", resp)
}
