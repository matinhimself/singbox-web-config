package clash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents a Clash API client
type Client struct {
	baseURL    string
	secret     string
	httpClient *http.Client
}

// ProxyGroup represents a proxy group
type ProxyGroup struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Now     string   `json:"now"`
	All     []string `json:"all"`
	History []struct {
		Name  string    `json:"name"`
		Delay int       `json:"delay"`
		Time  time.Time `json:"time"`
	} `json:"history"`
}

// Proxy represents a single proxy
type Proxy struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	History []struct {
		Delay int       `json:"delay"`
		Time  time.Time `json:"time"`
	} `json:"history"`
	All  []string `json:"all,omitempty"`
	Now  string   `json:"now,omitempty"`
	UDP  bool     `json:"udp"`
}

// ProxiesResponse represents the response from /proxies endpoint
type ProxiesResponse struct {
	Proxies map[string]Proxy `json:"proxies"`
}

// DelayTestResponse represents the response from delay test
type DelayTestResponse struct {
	Delay int `json:"delay"`
}

// NewClient creates a new Clash API client
func NewClient(baseURL, secret string) *Client {
	return &Client{
		baseURL: baseURL,
		secret:  secret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with auth headers
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.secret != "" {
		req.Header.Set("Authorization", "Bearer "+c.secret)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// GetProxies fetches all proxies from the Clash API
func (c *Client) GetProxies() (map[string]Proxy, error) {
	resp, err := c.doRequest("GET", "/proxies", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var result ProxiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Proxies, nil
}

// GetProxy fetches a specific proxy by name
func (c *Client) GetProxy(name string) (*Proxy, error) {
	resp, err := c.doRequest("GET", "/proxies/"+name, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var proxy Proxy
	if err := json.NewDecoder(resp.Body).Decode(&proxy); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &proxy, nil
}

// SwitchProxy switches the active proxy in a group
func (c *Client) SwitchProxy(groupName, proxyName string) error {
	body := map[string]string{"name": proxyName}

	resp, err := c.doRequest("PUT", "/proxies/"+groupName, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to switch proxy: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// TestProxyDelay tests the latency of a proxy
func (c *Client) TestProxyDelay(proxyName string, testURL string, timeout int) (int, error) {
	if testURL == "" {
		testURL = "http://www.gstatic.com/generate_204"
	}
	if timeout == 0 {
		timeout = 5000
	}

	path := fmt.Sprintf("/proxies/%s/delay?timeout=%d&url=%s", proxyName, timeout, testURL)
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("delay test failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result DelayTestResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Delay, nil
}

// GetProxyGroups returns only proxies that are groups (have "all" field)
func (c *Client) GetProxyGroups() (map[string]Proxy, error) {
	proxies, err := c.GetProxies()
	if err != nil {
		return nil, err
	}

	groups := make(map[string]Proxy)
	for name, proxy := range proxies {
		if len(proxy.All) > 0 {
			groups[name] = proxy
		}
	}

	return groups, nil
}
