package service

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	apierrors "github.com/ElissaDesign/vpn-dev/internal/errors"
	"github.com/ElissaDesign/vpn-dev/internal/model"
	"github.com/ElissaDesign/vpn-dev/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// DeviceService handles business logic for devices
type DeviceService struct {
	deviceRepo *repository.DeviceRepository
	usageRepo  *repository.DeviceUsageRepository
	userRepo   *repository.UserRepository
	serverRepo *repository.ServerRepository
	wgAgent    *WireGuardAgentClient
	logger     *zerolog.Logger
}

// NewDeviceService creates a new device service
func NewDeviceService(
	deviceRepo *repository.DeviceRepository,
	usageRepo *repository.DeviceUsageRepository,
	userRepo *repository.UserRepository,
	serverRepo *repository.ServerRepository,
	logger *zerolog.Logger,
) *DeviceService {
	return &DeviceService{
		deviceRepo: deviceRepo,
		usageRepo:  usageRepo,
		userRepo:   userRepo,
		serverRepo: serverRepo,
		wgAgent:    NewWireGuardAgentClient(logger),
		logger:     logger,
	}
}

// RegisterDevice registers a new device for a user
func (s *DeviceService) RegisterDevice(ctx context.Context, userID uuid.UUID, deviceName, deviceType string) (*model.Device, error) {
	// Check user exists
	_, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Str("userId", userID.String()).Msg("User not found")
		return nil, apierrors.ErrNotFound
	}

	// Check device limit
	count, err := s.deviceRepo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= 5 {
		s.logger.Warn().Str("userId", userID.String()).Int("count", count).Msg("Device limit reached")
		return nil, fmt.Errorf("device limit reached (max 5 devices)")
	}

	// Get healthy servers for load balancing
	healthyServers, err := s.serverRepo.GetHealthyServers(ctx)
	if err != nil || len(healthyServers) == 0 {
		s.logger.Error().Err(err).Msg("No healthy servers available")
		return nil, fmt.Errorf("no healthy VPN servers available")
	}

	// Select first server (lowest load)
	selectedServer := &healthyServers[0]

	// Allocate IP from server's CIDR
	ipStr, err := s.allocateIP(ctx, selectedServer)
	if err != nil {
		return nil, err
	}

	// Generate WireGuard keys
	privateKey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate WireGuard keys")
		return nil, fmt.Errorf("failed to generate VPN keys")
	}

	pubKey := privateKey.PublicKey()
	privKeyStr := privateKey.String()
	pubKeyStr := pubKey.String()

	// Create device object
	device := &model.Device{
		ID:         uuid.New(),
		UserID:     userID,
		DeviceName: deviceName,
		DeviceType: deviceType,
		PublicKey:  pubKeyStr,
		PrivateKey: privKeyStr,
		IP:         ipStr,
		Status:     "disconnected",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Create device in database
	if err := s.deviceRepo.Create(ctx, device); err != nil {
		s.logger.Error().Err(err).Str("userId", userID.String()).Msg("Failed to create device")
		return nil, err
	}

	// Call WireGuard agent to add peer
	err = s.wgAgent.AddPeer(ctx,
		selectedServer.AgentURL,
		selectedServer.AgentAPIKey,
		device.PublicKey,
		device.IP,
		"")

	if err != nil {
		s.logger.Error().Err(err).Str("deviceId", device.ID.String()).Msg("Failed to add peer to WireGuard - rolling back")
		// Rollback: delete device from database
		if delErr := s.deviceRepo.Delete(ctx, device.ID); delErr != nil {
			s.logger.Error().Err(delErr).Msg("Failed to rollback device creation")
		}
		return nil, fmt.Errorf("failed to configure WireGuard: %w", err)
	}

	s.logger.Info().
		Str("deviceId", device.ID.String()).
		Str("userId", userID.String()).
		Str("deviceName", deviceName).
		Msg("Device registered successfully")

	return device, nil
}

// GetUserDevices returns all devices for a user
func (s *DeviceService) GetUserDevices(ctx context.Context, userID uuid.UUID) ([]model.Device, error) {
	devices, err := s.deviceRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if devices == nil {
		devices = []model.Device{}
	}
	return devices, nil
}

// GetDeviceByID returns a specific device
func (s *DeviceService) GetDeviceByID(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
	return s.deviceRepo.FindByID(ctx, deviceID)
}

// RemoveDevice removes a device
func (s *DeviceService) RemoveDevice(ctx context.Context, deviceID uuid.UUID) error {
	_, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return err
	}

	// Note: We don't have the server URL/key from device record
	// This would need to be stored or fetched from elsewhere
	// For now, we'll skip the agent call and just remove from DB
	s.logger.Warn().
		Str("deviceId", deviceID.String()).
		Msg("Device being removed - WireGuard cleanup requires additional context")

	// Delete device from database
	if err := s.deviceRepo.Delete(ctx, deviceID); err != nil {
		return err
	}

	s.logger.Info().Str("deviceId", deviceID.String()).Msg("Device removed successfully")

	return nil
}

// RecordUsage records usage for a device
func (s *DeviceService) RecordUsage(ctx context.Context, deviceID uuid.UUID, bytesSent, bytesReceived int64) error {
	device, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return err
	}

	usage := &model.DeviceUsage{
		ID:            uuid.New(),
		DeviceID:      deviceID,
		UserID:        device.UserID,
		BytesSent:     bytesSent,
		BytesReceived: bytesReceived,
		RecordedAt:    time.Now(),
	}

	if err := s.usageRepo.RecordUsage(ctx, usage); err != nil {
		s.logger.Error().Err(err).Str("deviceId", deviceID.String()).Msg("Failed to record usage")
		return err
	}

	return nil
}

// GetDeviceUsage returns usage summary for a device
func (s *DeviceService) GetDeviceUsage(ctx context.Context, deviceID uuid.UUID) (*model.DeviceUsageSummary, error) {
	// Get device info
	device, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	// Get usage summary
	summary, err := s.usageRepo.GetTotalUsageByDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	// Set device name
	summary.DeviceID = device.ID
	summary.DeviceName = device.DeviceName

	return summary, nil
}

// GetUserUsage returns usage summary for all user devices
func (s *DeviceService) GetUserUsage(ctx context.Context, userID uuid.UUID) (*model.UserUsageSummary, error) {
	return s.usageRepo.GetUserUsageSummary(ctx, userID)
}

// allocateIP allocates a new IP for a device from the server's CIDR pool
func (s *DeviceService) allocateIP(ctx context.Context, server *model.Server) (string, error) {
	// Parse the CIDR (trim whitespace)
	cidr := strings.TrimSpace(server.NetworkCIDR)
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("invalid server CIDR: %w", err)
	}

	// Get base IP
	baseIP := ipNet.IP.To4()
	if baseIP == nil {
		return "", fmt.Errorf("invalid IPv4 CIDR")
	}

	// Convert to number and add offset
	ipNum := uint32(baseIP[0])<<24 | uint32(baseIP[1])<<16 | uint32(baseIP[2])<<8 | uint32(baseIP[3])
	nextNum := ipNum + uint32(server.NextAvailableIP)

	// Convert back to IP
	newIP := net.IPv4(byte(nextNum>>24), byte(nextNum>>16), byte(nextNum>>8), byte(nextNum))

	// Update server's next_available_ip
	server.NextAvailableIP++
	if err := s.serverRepo.UpdateNextAvailableIP(ctx, server.ID, server.NextAvailableIP); err != nil {
		return "", fmt.Errorf("failed to update server IP counter: %w", err)
	}

	return newIP.String(), nil
}

// GenerateWireGuardConfig generates a WireGuard configuration file for a device
func (s *DeviceService) GenerateWireGuardConfig(ctx context.Context, deviceID, userID uuid.UUID, serverID string) (string, error) {
	// Get device
	device, err := s.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		s.logger.Error().Err(err).Str("deviceId", deviceID.String()).Msg("Device not found")
		return "", err
	}

	// Verify device belongs to user
	if device.UserID != userID {
		s.logger.Warn().
			Str("deviceId", deviceID.String()).
			Str("userId", userID.String()).
			Str("deviceOwnerId", device.UserID.String()).
			Msg("User does not own this device")
		return "", fmt.Errorf("unauthorized: user does not own this device")
	}

	// Get server
	server, err := s.serverRepo.GetByID(ctx, serverID)
	if err != nil {
		s.logger.Error().Err(err).Str("serverId", serverID).Msg("Server not found")
		return "", err
	}

	// Generate WireGuard config
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32
DNS = 1.1.1.1, 8.8.8.8

[Peer]
PublicKey = %s
Endpoint = %s:%d
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
`,
		device.PrivateKey,
		device.IP,
		server.ServerPublicKey,
		server.PublicIP,
		server.WGPort,
	)

	s.logger.Info().
		Str("deviceId", deviceID.String()).
		Str("userId", userID.String()).
		Str("serverId", serverID).
		Msg("WireGuard config generated successfully")

	return config, nil
}

// GenerateConfig generates a VPN configuration file (supports multiple protocols)
func (s *DeviceService) GenerateConfig(ctx context.Context, deviceID, userID uuid.UUID, serverID, protocol string) (string, error) {
	if protocol == "" || protocol == "wireguard" {
		return s.GenerateWireGuardConfig(ctx, deviceID, userID, serverID)
	}

	return "", fmt.Errorf("unsupported protocol: %s (only 'wireguard' is supported)", protocol)
}
