package emby

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// CreateUser creates a new user on the media server (Emby/Jellyfin)
func (es *embyService) CreateUser(ctx context.Context, username, password string) (string, error) {
	baseURL, apiKey := es.getConfig()
	
	// Prepare the request payload
	userPayload := map[string]interface{}{
		"Name":     username,
		"Password": password,
	}
	
	payloadBytes, err := json.Marshal(userPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user creation payload: %w", err)
	}
	
	// Create HTTP request
	url := fmt.Sprintf("%s/users/new?api_key=%s", baseURL, apiKey)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	
	resp, err := es.client.Do(req)
	if err != nil {
		slog.Error("Failed to create media server user", "error", err, "url", url, "username", username)
		return "", fmt.Errorf("failed to contact media server: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		slog.Error("Media server user creation failed", 
			"status", resp.StatusCode, 
			"response", string(respBody),
			"username", username)
		return "", fmt.Errorf("media server rejected user creation: %s", string(respBody))
	}
	
	// Parse response to get user ID
	var userResponse struct {
		ID   string `json:"Id"`
		Name string `json:"Name"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&userResponse); err != nil {
		return "", fmt.Errorf("failed to decode media server response: %w", err)
	}
	
	slog.Info("Media server user created successfully", 
		"user_id", userResponse.ID, 
		"username", username)
	
	return userResponse.ID, nil
}