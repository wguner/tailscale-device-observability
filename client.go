package tailscale

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://api.tailscale.com/api/v2"

// Client wraps the Tailscale HTTP API.
type Client struct {
	apiKey  string
	tailnet string
	http    *http.Client
}

// Device represents a single node in the tailnet as returned by the API.
type Device struct {
	ID                string    `json:"id"`
	Hostname          string    `json:"hostname"`
	OS                string    `json:"os"`
	LastSeen          time.Time `json:"lastSeen"`
	KeyExpiryDisabled bool      `json:"keyExpiryDisabled"`
	Expires           time.Time `json:"keyExpiry"`
	Authorized        bool      `json:"authorized"`
	UpdateAvailable   bool      `json:"updateAvailable"`
}

type devicesResponse struct {
	Devices []Device `json:"devices"`
}

// NewClient returns a configured Tailscale API client.
func NewClient(apiKey, tailnet string) *Client {
	return &Client{
		apiKey:  apiKey,
		tailnet: tailnet,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// GetDevices fetches all devices registered to the tailnet.
func (c *Client) GetDevices() ([]Device, error) {
	url := fmt.Sprintf("%s/tailnet/%s/devices", baseURL, c.tailnet)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tailscale API returned status %d", resp.StatusCode)
	}

	var result devicesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Devices, nil
}
