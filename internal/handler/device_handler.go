package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ElissaDesign/vpn-dev/internal/middleware"
	"github.com/ElissaDesign/vpn-dev/internal/service"
	"github.com/ElissaDesign/vpn-dev/internal/utils/response"
	"github.com/ElissaDesign/vpn-dev/internal/validator"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// DeviceHandler handles device-related HTTP requests
type DeviceHandler struct {
	deviceService *service.DeviceService
	validator     *validator.Validator
	logger        *zerolog.Logger
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(deviceService *service.DeviceService, validator *validator.Validator, logger *zerolog.Logger) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
		validator:     validator,
		logger:        logger,
	}
}

// RegisterDeviceRequest represents device registration request
type RegisterDeviceRequest struct {
	DeviceName string `json:"deviceName" validate:"required,min=1,max=100"`
	DeviceType string `json:"deviceType" validate:"required,oneof=mobile desktop tablet router unknown"`
}

// DeviceResponse represents device in responses
type DeviceResponse struct {
	DeviceID      string  `json:"deviceId"`
	DeviceName    string  `json:"deviceName"`
	DeviceType    string  `json:"deviceType"`
	IP            string  `json:"ip"`
	PublicKey     string  `json:"publicKey"`
	Status        string  `json:"status"`
	LastConnected *string `json:"lastConnected,omitempty"`
	CreatedAt     string  `json:"createdAt"`
}

// RegisterDevice registers a new device for a user
// @Summary Register new VPN device
// @Description Register a new device for user with device name and type
// @Tags Devices
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param request body RegisterDeviceRequest true "Device registration request"
// @Success 201 {object} response.JSONResponse{data=DeviceResponse} "Device registered successfully"
// @Failure 400 {object} response.JSONResponse "Invalid input"
// @Failure 404 {object} response.JSONResponse "User not found"
// @Failure 409 {object} response.JSONResponse "Device limit reached"
// @Failure 500 {object} response.JSONResponse "Internal server error"
// @Router /users/{userId}/devices [post]
func (h *DeviceHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)
	userID := r.PathValue("userId")

	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Warn().Str("requestId", requestID).Str("userId", userID).Msg("Invalid user ID format")
		response.BadRequest(w, "Invalid user ID", map[string]string{"userId": "Invalid UUID format"})
		return
	}

	// Parse request
	var req RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn().Str("requestId", requestID).Err(err).Msg("Invalid request body")
		response.BadRequest(w, "Invalid request body", map[string]string{"body": "Invalid JSON"})
		return
	}

	// Validate request
	validationErrs := h.validator.Validate(req)
	if len(validationErrs) > 0 {
		h.logger.Warn().Str("requestId", requestID).Interface("errors", validationErrs).Msg("Validation errors")
		fieldErrors := make(map[string]string)
		for _, ve := range validationErrs {
			fieldErrors[ve.Field] = ve.Message
		}
		response.BadRequest(w, "Validation failed", fieldErrors)
		return
	}

	// Register device
	device, err := h.deviceService.RegisterDevice(ctx, uid, req.DeviceName, req.DeviceType)
	if err != nil {
		h.logger.Error().Str("requestId", requestID).Str("userId", userID).Err(err).Msg("Failed to register device")
		response.InternalError(w, err.Error(), "device_error")
		return
	}

	h.logger.Info().Str("requestId", requestID).Str("userId", userID).Str("deviceId", device.ID.String()).Msg("Device registered successfully")

	resp := DeviceResponse{
		DeviceID:   device.ID.String(),
		DeviceName: device.DeviceName,
		DeviceType: device.DeviceType,
		IP:         device.IP,
		PublicKey:  device.PublicKey,
		Status:     device.Status,
		CreatedAt:  device.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if device.LastConnected != nil {
		lastConnected := device.LastConnected.Format("2006-01-02T15:04:05Z07:00")
		resp.LastConnected = &lastConnected
	}

	w.Header().Set("Location", "/users/"+userID+"/devices/"+device.ID.String())
	response.Created(w, "Device registered successfully", resp)
}

// GetUserDevices lists all devices for a user
// @Summary List user devices
// @Description Get all devices registered for a user
// @Tags Devices
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} response.JSONResponse{data=[]DeviceResponse} "Devices retrieved successfully"
// @Failure 400 {object} response.JSONResponse "Invalid user ID"
// @Failure 500 {object} response.JSONResponse "Internal server error"
// @Router /users/{userId}/devices [get]
func (h *DeviceHandler) GetUserDevices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)
	userID := r.PathValue("userId")

	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Warn().Str("requestId", requestID).Str("userId", userID).Msg("Invalid user ID format")
		response.BadRequest(w, "Invalid user ID", map[string]string{"userId": "Invalid UUID format"})
		return
	}

	// Get devices
	devices, err := h.deviceService.GetUserDevices(ctx, uid)
	if err != nil {
		h.logger.Error().Str("requestId", requestID).Err(err).Msg("Failed to get devices")
		response.InternalError(w, "Failed to get devices", "device_error")
		return
	}

	// Convert to response format
	var deviceResponses []DeviceResponse
	for _, device := range devices {
		resp := DeviceResponse{
			DeviceID:   device.ID.String(),
			DeviceName: device.DeviceName,
			DeviceType: device.DeviceType,
			IP:         device.IP,
			PublicKey:  device.PublicKey,
			Status:     device.Status,
			CreatedAt:  device.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if device.LastConnected != nil {
			lastConnected := device.LastConnected.Format("2006-01-02T15:04:05Z07:00")
			resp.LastConnected = &lastConnected
		}
		deviceResponses = append(deviceResponses, resp)
	}

	h.logger.Info().Str("requestId", requestID).Str("userId", userID).Int("count", len(deviceResponses)).Msg("Devices retrieved successfully")

	response.OK(w, "Devices retrieved successfully", deviceResponses)
}

// GetDeviceByID gets a specific device
// @Summary Get device details
// @Description Get details for a specific device
// @Tags Devices
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param deviceId path string true "Device ID"
// @Success 200 {object} response.JSONResponse{data=DeviceResponse} "Device retrieved successfully"
// @Failure 400 {object} response.JSONResponse "Invalid ID format"
// @Failure 404 {object} response.JSONResponse "Device not found"
// @Failure 500 {object} response.JSONResponse "Internal server error"
// @Router /users/{userId}/devices/{deviceId} [get]
func (h *DeviceHandler) GetDeviceByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)
	deviceID := r.PathValue("deviceId")

	// Parse device ID
	did, err := uuid.Parse(deviceID)
	if err != nil {
		h.logger.Warn().Str("requestId", requestID).Str("deviceId", deviceID).Msg("Invalid device ID format")
		response.BadRequest(w, "Invalid device ID", map[string]string{"deviceId": "Invalid UUID format"})
		return
	}

	// Get device
	device, err := h.deviceService.GetDeviceByID(ctx, did)
	if err != nil {
		h.logger.Warn().Str("requestId", requestID).Str("deviceId", deviceID).Msg("Device not found")
		response.NotFound(w, "Device not found")
		return
	}

	resp := DeviceResponse{
		DeviceID:   device.ID.String(),
		DeviceName: device.DeviceName,
		DeviceType: device.DeviceType,
		IP:         device.IP,
		PublicKey:  device.PublicKey,
		Status:     device.Status,
		CreatedAt:  device.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if device.LastConnected != nil {
		lastConnected := device.LastConnected.Format("2006-01-02T15:04:05Z07:00")
		resp.LastConnected = &lastConnected
	}

	h.logger.Info().Str("requestId", requestID).Str("deviceId", deviceID).Msg("Device retrieved successfully")

	response.OK(w, "Device retrieved successfully", resp)
}

// DeleteDevice deletes a device
// @Summary Delete device
// @Description Remove a device from user account
// @Tags Devices
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param deviceId path string true "Device ID"
// @Success 200 {object} response.JSONResponse "Device deleted successfully"
// @Failure 400 {object} response.JSONResponse "Invalid ID format"
// @Failure 404 {object} response.JSONResponse "Device not found"
// @Failure 500 {object} response.JSONResponse "Internal server error"
// @Router /users/{userId}/devices/{deviceId} [delete]
func (h *DeviceHandler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)
	deviceID := r.PathValue("deviceId")

	// Parse device ID
	did, err := uuid.Parse(deviceID)
	if err != nil {
		h.logger.Warn().Str("requestId", requestID).Str("deviceId", deviceID).Msg("Invalid device ID format")
		response.BadRequest(w, "Invalid device ID", map[string]string{"deviceId": "Invalid UUID format"})
		return
	}

	// Delete device
	if err := h.deviceService.RemoveDevice(ctx, did); err != nil {
		h.logger.Warn().Str("requestId", requestID).Str("deviceId", deviceID).Msg("Device not found")
		response.NotFound(w, "Device not found")
		return
	}

	h.logger.Info().Str("requestId", requestID).Str("deviceId", deviceID).Msg("Device deleted successfully")

	response.OK(w, "Device deleted successfully", nil)
}

// GetDeviceUsage gets usage for a specific device
// @Summary Get device usage
// @Description Get bandwidth usage for a specific device
// @Tags Usage
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param deviceId path string true "Device ID"
// @Success 200 {object} response.JSONResponse{data=map[string]interface{}} "Device usage retrieved successfully"
// @Failure 400 {object} response.JSONResponse "Invalid ID format"
// @Failure 404 {object} response.JSONResponse "Device not found"
// @Failure 500 {object} response.JSONResponse "Internal server error"
// @Router /users/{userId}/devices/{deviceId}/usage [get]
func (h *DeviceHandler) GetDeviceUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)
	deviceID := r.PathValue("deviceId")

	// Parse device ID
	did, err := uuid.Parse(deviceID)
	if err != nil {
		h.logger.Warn().Str("requestId", requestID).Str("deviceId", deviceID).Msg("Invalid device ID format")
		response.BadRequest(w, "Invalid device ID", map[string]string{"deviceId": "Invalid UUID format"})
		return
	}

	// Get usage
	usage, err := h.deviceService.GetDeviceUsage(ctx, did)
	if err != nil {
		h.logger.Warn().Str("requestId", requestID).Str("deviceId", deviceID).Msg("Device not found")
		response.NotFound(w, "Device not found")
		return
	}

	h.logger.Info().Str("requestId", requestID).Str("deviceId", deviceID).Msg("Device usage retrieved successfully")

	response.OK(w, "Device usage retrieved successfully", usage)
}

// GetUserUsage gets total usage for all user devices
// @Summary Get user total usage
// @Description Get total bandwidth usage across all user devices
// @Tags Usage
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} response.JSONResponse{data=map[string]interface{}} "User usage retrieved successfully"
// @Failure 400 {object} response.JSONResponse "Invalid user ID"
// @Failure 500 {object} response.JSONResponse "Internal server error"
// @Router /users/{userId}/usage [get]
func (h *DeviceHandler) GetUserUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)
	userID := r.PathValue("userId")

	// Parse user ID
	uid, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Warn().Str("requestId", requestID).Str("userId", userID).Msg("Invalid user ID format")
		response.BadRequest(w, "Invalid user ID", map[string]string{"userId": "Invalid UUID format"})
		return
	}

	// Get usage
	usage, err := h.deviceService.GetUserUsage(ctx, uid)
	if err != nil {
		h.logger.Error().Str("requestId", requestID).Err(err).Msg("Failed to get user usage")
		response.InternalError(w, "Failed to get user usage", "usage_error")
		return
	}

	h.logger.Info().Str("requestId", requestID).Str("userId", userID).Msg("User usage retrieved successfully")

	response.OK(w, "User usage retrieved successfully", usage)
}

// GetDeviceConfig generates and returns VPN configuration file for a device
// @Summary Get VPN configuration file for a device
// @Description Generates a WireGuard configuration file that can be downloaded or scanned with QR code
// @Tags devices
// @Param userId path string true "User ID" format(uuid)
// @Param deviceId path string true "Device ID" format(uuid)
// @Param serverId query string true "Server ID"
// @Param protocol query string false "Protocol (default: wireguard)" default(wireguard) enum(wireguard)
// @Success 200 {file} string "WireGuard configuration file"
// @Failure 400 {object} response.JSONResponse "Invalid request"
// @Failure 404 {object} response.JSONResponse "Not found"
// @Failure 500 {object} response.JSONResponse "Internal server error"
// @Router /users/{userId}/devices/{deviceId}/config [get]
func (h *DeviceHandler) GetDeviceConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := middleware.GetRequestID(ctx)

	// Extract path parameters
	userID := r.PathValue("userId")
	deviceID := r.PathValue("deviceId")

	// Extract query parameters
	serverId := r.URL.Query().Get("serverId")
	protocol := r.URL.Query().Get("protocol")
	if protocol == "" {
		protocol = "wireguard"
	}

	// Validate IDs
	uid, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Warn().Str("requestId", requestID).Str("userId", userID).Err(err).Msg("Invalid user ID format")
		response.BadRequest(w, "Invalid user ID", map[string]string{"userId": "Invalid UUID format"})
		return
	}

	did, err := uuid.Parse(deviceID)
	if err != nil {
		h.logger.Warn().Str("requestId", requestID).Str("deviceId", deviceID).Err(err).Msg("Invalid device ID format")
		response.BadRequest(w, "Invalid device ID", map[string]string{"deviceId": "Invalid UUID format"})
		return
	}

	// Validate server ID
	if serverId == "" {
		h.logger.Warn().Str("requestId", requestID).Msg("Server ID is required")
		response.BadRequest(w, "Server ID is required", map[string]string{"serverId": "Required parameter"})
		return
	}

	// Generate configuration
	config, err := h.deviceService.GenerateConfig(ctx, did, uid, serverId, protocol)
	if err != nil {
		if err.Error() == "not found" || err.Error() == "unauthorized: user does not own this device" {
			h.logger.Warn().Str("requestId", requestID).Str("deviceId", deviceID).Err(err).Msg("Device not found or unauthorized")
			response.NotFound(w, "Device not found")
			return
		}

		h.logger.Error().Str("requestId", requestID).Str("deviceId", deviceID).Err(err).Msg("Failed to generate config")
		response.InternalError(w, "Failed to generate configuration", "config_error")
		return
	}

	// Return config file as downloadable content
	filename := "vpn-config.conf"
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)

	h.logger.Info().
		Str("requestId", requestID).
		Str("userId", userID).
		Str("deviceId", deviceID).
		Str("serverId", serverId).
		Str("protocol", protocol).
		Msg("VPN configuration generated and sent successfully")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(config))
}
