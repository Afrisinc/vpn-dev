package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// WireGuardAgentClient handles communication with WireGuard agents on VPS servers
type WireGuardAgentClient struct {
	httpClient *http.Client
	logger     *zerolog.Logger
}

// AddPeerRequest is the request payload for adding a peer to WireGuard
type AddPeerRequest struct {
	PublicKey string `json:"public_key"`
	AllowedIP string `json:"allowed_ip"`
	Endpoint  string `json:"endpoint"`
}

// RemovePeerRequest is the request payload for removing a peer from WireGuard
type RemovePeerRequest struct {
	PublicKey string `json:"public_key"`
}

// WireGuardResponse is the response from WireGuard agent
type WireGuardResponse struct {
	Success bool   `json:"success"` // Fallback for standard responses
	Status  string `json:"status"`  // Agent uses "status" field instead
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// IsSuccess returns true if the operation was successful
func (w *WireGuardResponse) IsSuccess() bool {
	// Check if success field is true
	if w.Success {
		return true
	}
	// Also check if status field is "success"
	return strings.ToLower(w.Status) == "success"
}

// NewWireGuardAgentClient creates a new WireGuard agent client
func NewWireGuardAgentClient(logger *zerolog.Logger) *WireGuardAgentClient {
	return &WireGuardAgentClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// AddPeer adds a new peer to the WireGuard interface on the VPS
// agentURL: https://wg-agent-nyc.example.com
// agentAPIKey: secret-key-nyc
// publicKey: user's WireGuard public key
// allowedIP: user's VPN IP with /32 (e.g., 192.168.88.2/32)
// endpoint: optional user's public endpoint for peer communication
func (w *WireGuardAgentClient) AddPeer(ctx context.Context, agentURL, agentAPIKey, publicKey, allowedIP, endpoint string) error {
	payload := AddPeerRequest{
		PublicKey: publicKey,
		AllowedIP: allowedIP,
		Endpoint:  endpoint,
	}

	return w.doRequest(ctx, agentURL, "/add-peer", agentAPIKey, payload)
}

// RemovePeer removes a peer from the WireGuard interface on the VPS
// agentURL: https://wg-agent-nyc.example.com
// agentAPIKey: secret-key-nyc
// publicKey: user's WireGuard public key
func (w *WireGuardAgentClient) RemovePeer(ctx context.Context, agentURL, agentAPIKey, publicKey string) error {
	payload := RemovePeerRequest{
		PublicKey: publicKey,
	}

	return w.doRequest(ctx, agentURL, "/remove-peer", agentAPIKey, payload)
}

// doRequest makes the HTTP request to the WireGuard agent
func (w *WireGuardAgentClient) doRequest(ctx context.Context, agentURL, endpoint, apiKey string, payload interface{}) error {
	// Prepare request body
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		w.logger.Error().Err(err).Msg("Failed to marshal WireGuard agent request")
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := agentURL + endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		w.logger.Error().Err(err).Str("url", url).Msg("Failed to create WireGuard agent request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	// Execute request
	resp, err := w.httpClient.Do(req)
	if err != nil {
		w.logger.Error().
			Err(err).
			Str("url", url).
			Str("endpoint", endpoint).
			Msg("WireGuard agent request failed")
		return fmt.Errorf("agent request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		w.logger.Error().
			Err(err).
			Str("url", url).
			Msg("Failed to read WireGuard agent response")
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var wgResp WireGuardResponse
	if err := json.Unmarshal(respBody, &wgResp); err != nil {
		w.logger.Warn().
			Err(err).
			Str("url", url).
			Str("response", string(respBody)).
			Msg("Failed to parse WireGuard agent response")
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := wgResp.Error
		if errMsg == "" {
			errMsg = wgResp.Message
		}
		w.logger.Error().
			Int("statusCode", resp.StatusCode).
			Str("url", url).
			Str("endpoint", endpoint).
			Str("message", wgResp.Message).
			Str("error", wgResp.Error).
			Msg("WireGuard agent returned error status")
		return fmt.Errorf("agent error (status %d): %s", resp.StatusCode, errMsg)
	}

	// Check response success flag (supports both success and status fields)
	if !wgResp.IsSuccess() {
		errMsg := wgResp.Error
		if errMsg == "" {
			errMsg = wgResp.Message
		}
		w.logger.Error().
			Str("url", url).
			Str("endpoint", endpoint).
			Str("status", wgResp.Status).
			Bool("success", wgResp.Success).
			Str("message", wgResp.Message).
			Str("error", errMsg).
			Msg("WireGuard agent operation failed")
		return fmt.Errorf("agent operation failed: %s", errMsg)
	}

	w.logger.Info().
		Str("url", url).
		Str("endpoint", endpoint).
		Str("message", wgResp.Message).
		Msg("WireGuard agent operation succeeded")

	return nil
}
