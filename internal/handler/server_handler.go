package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ElissaDesign/vpn-dev/internal/errors"
	"github.com/ElissaDesign/vpn-dev/internal/handler/dto"
	"github.com/ElissaDesign/vpn-dev/internal/middleware"
	"github.com/ElissaDesign/vpn-dev/internal/model"
	"github.com/ElissaDesign/vpn-dev/internal/repository"
	"github.com/ElissaDesign/vpn-dev/internal/utils/response"
	"github.com/ElissaDesign/vpn-dev/internal/validator"
	"github.com/rs/zerolog"
)

// ServerHandler handles server-related HTTP requests
type ServerHandler struct {
	serverRepo *repository.ServerRepository
	validator  *validator.Validator
	logger     *zerolog.Logger
}

// NewServerHandler creates a new server handler
func NewServerHandler(sr *repository.ServerRepository, val *validator.Validator, logger *zerolog.Logger) *ServerHandler {
	return &ServerHandler{
		serverRepo: sr,
		validator:  val,
		logger:     logger,
	}
}

// GetAllServers retrieves all VPN servers
// @Summary Get all VPN servers
// @Description Retrieve a list of all registered VPN servers with their status and capacity
// @Tags Servers
// @Accept json
// @Produce json
// @Success 200 {object} response.JSONResponse{data=[]dto.ServerResponse} "List of all servers"
// @Failure 500 {object} response.JSONResponse "Internal server error"
// @Router /admin/servers [get]
func (h *ServerHandler) GetAllServers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)

	servers, err := h.serverRepo.GetAll(ctx)
	if err != nil {
		h.logger.Error().
			Str("requestId", requestID).
			Err(err).
			Msg("Failed to fetch servers")
		response.InternalError(w, "Failed to fetch servers", "database_error")
		return
	}

	h.logger.Info().
		Str("requestId", requestID).
		Int("count", len(servers)).
		Msg("Successfully fetched servers")

	// Convert to response DTOs
	serverResponses := make([]dto.ServerResponse, len(servers))
	for i, s := range servers {
		serverResponses[i] = h.modelToDTO(&s)
	}

	response.OK(w, "Servers fetched successfully", serverResponses)
}

// GetHealthyServers retrieves servers that are healthy and have capacity
// @Summary Get healthy VPN servers
// @Description Retrieve only servers that are active, healthy, and have available capacity
// @Tags Servers
// @Accept json
// @Produce json
// @Success 200 {object} response.JSONResponse{data=[]dto.ServerResponse} "List of healthy servers"
// @Failure 500 {object} response.JSONResponse "Internal server error"
// @Router /admin/servers/healthy [get]
func (h *ServerHandler) GetHealthyServers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)

	servers, err := h.serverRepo.GetHealthyServers(ctx)
	if err != nil {
		h.logger.Error().
			Str("requestId", requestID).
			Err(err).
			Msg("Failed to fetch healthy servers")
		response.InternalError(w, "Failed to fetch servers", "database_error")
		return
	}

	h.logger.Info().
		Str("requestId", requestID).
		Int("count", len(servers)).
		Msg("Successfully fetched healthy servers")

	// Convert to response DTOs
	serverResponses := make([]dto.ServerResponse, len(servers))
	for i, s := range servers {
		serverResponses[i] = h.modelToDTO(&s)
	}

	response.OK(w, "Healthy servers fetched successfully", serverResponses)
}

// CreateServer registers a new VPN server
// @Summary Register a new VPN server
// @Description Register a new VPN server node with WireGuard agent connection details
// @Tags Servers
// @Accept json
// @Produce json
// @Param request body dto.CreateServerRequest true "Server registration request"
// @Success 201 {object} response.JSONResponse{data=dto.ServerResponse} "Server created successfully"
// @Failure 400 {object} response.JSONResponse "Invalid input"
// @Failure 409 {object} response.JSONResponse "Server ID already exists"
// @Failure 500 {object} response.JSONResponse "Internal server error"
// @Router /admin/servers [post]
func (h *ServerHandler) CreateServer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)

	// Parse request
	var req dto.CreateServerRequest
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

	// Create server object
	server := &model.Server{
		ID:                req.ID,
		Name:              req.Name,
		Location:          req.Location,
		RegionCode:        req.RegionCode,
		CountryCode:       req.CountryCode,
		PublicIP:          req.PublicIP,
		AgentURL:          req.AgentURL,
		AgentAPIKey:       req.AgentAPIKey,
		WGPort:            req.WireGuardPort,
		ServerPublicKey:   req.ServerPublicKey,
		NetworkCIDR:       req.NetworkCIDR,
		MaxClients:        req.MaxClients,
		CurrentClients:    0,
		BandwidthLimitMbp: req.BandwidthLimit,
		Status:            "active",
		HealthStatus:      "unknown",
		Latitude:          req.Latitude,
		Longitude:         req.Longitude,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Create server in database
	if err := h.serverRepo.Create(ctx, server); err != nil {
		h.logger.Error().
			Str("requestId", requestID).
			Err(err).
			Msg("Failed to create server in database")
		response.InternalError(w, "Failed to create server", "database_error")
		return
	}

	h.logger.Info().
		Str("requestId", requestID).
		Str("serverId", server.ID).
		Str("region", server.RegionCode).
		Msg("Server registered successfully")

	// Return response
	resp := h.modelToDTO(server)
	w.Header().Set("Location", "/admin/servers/"+server.ID)
	response.Created(w, "Server registered successfully", resp)
}

// UpdateServerStatus updates a server's status and health
// @Summary Update server status
// @Description Update a VPN server's operational status and health status
// @Tags Servers
// @Accept json
// @Produce json
// @Param serverId path string true "Server ID"
// @Param request body dto.UpdateServerStatusRequest true "Status update request"
// @Success 200 {object} response.JSONResponse{data=dto.ServerResponse} "Server status updated successfully"
// @Failure 400 {object} response.JSONResponse "Invalid input"
// @Failure 404 {object} response.JSONResponse "Server not found"
// @Failure 500 {object} response.JSONResponse "Internal server error"
// @Router /admin/servers/{serverId}/status [put]
func (h *ServerHandler) UpdateServerStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)

	// Get server ID from URL
	serverId := r.PathValue("serverId")
	if serverId == "" {
		h.logger.Warn().
			Str("requestId", requestID).
			Msg("Missing serverId in path")
		response.BadRequest(w, "Missing serverId path parameter", map[string]string{
			"serverId": "serverId is required",
		})
		return
	}

	// Parse request
	var req dto.UpdateServerStatusRequest
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

	// Fetch server
	server, err := h.serverRepo.GetByID(ctx, serverId)
	if err != nil {
		if err == errors.ErrNotFound {
			h.logger.Warn().
				Str("requestId", requestID).
				Str("serverId", serverId).
				Msg("Server not found")
			response.NotFound(w, "Server not found")
			return
		}
		h.logger.Error().
			Str("requestId", requestID).
			Err(err).
			Msg("Failed to lookup server")
		response.InternalError(w, "Failed to lookup server", "database_error")
		return
	}

	// Update server status
	if err := h.serverRepo.UpdateStatus(ctx, serverId, req.Status, req.HealthStatus); err != nil {
		h.logger.Error().
			Str("requestId", requestID).
			Err(err).
			Str("serverId", serverId).
			Msg("Failed to update server status")
		response.InternalError(w, "Failed to update server status", "database_error")
		return
	}

	// Update local object for response
	server.Status = req.Status
	server.HealthStatus = req.HealthStatus
	server.UpdatedAt = time.Now()

	h.logger.Info().
		Str("requestId", requestID).
		Str("serverId", serverId).
		Str("status", req.Status).
		Str("healthStatus", req.HealthStatus).
		Msg("Server status updated successfully")

	resp := h.modelToDTO(server)
	response.OK(w, "Server status updated successfully", resp)
}

// Helper function to convert Server model to DTO
func (h *ServerHandler) modelToDTO(server *model.Server) dto.ServerResponse {
	return dto.ServerResponse{
		ID:              server.ID,
		Name:            server.Name,
		Location:        server.Location,
		RegionCode:      server.RegionCode,
		CountryCode:     server.CountryCode,
		PublicIP:        server.PublicIP,
		AgentURL:        server.AgentURL,
		WireGuardPort:   server.WGPort,
		ServerPublicKey: server.ServerPublicKey,
		NetworkCIDR:     server.NetworkCIDR,
		MaxClients:      server.MaxClients,
		CurrentClients:  server.CurrentClients,
		BandwidthLimit:  server.BandwidthLimitMbp,
		Status:          server.Status,
		HealthStatus:    server.HealthStatus,
		Latitude:        server.Latitude,
		Longitude:       server.Longitude,
		CreatedAt:       server.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       server.UpdatedAt.Format(time.RFC3339),
	}
}
