package clash

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Config represents Clash API configuration
type Config struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

// ConfigManager handles Clash configuration persistence
type ConfigManager struct {
	configPath string
}

// NewConfigManager creates a new config manager
func NewConfigManager() (*ConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "singbox-web-config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &ConfigManager{
		configPath: filepath.Join(configDir, "clash.json"),
	}, nil
}

// Load loads the Clash configuration from file
func (cm *ConfigManager) Load() (*Config, error) {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save saves the Clash configuration to file
func (cm *ConfigManager) Save(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// TestConnection tests if a Clash API endpoint is accessible
func TestConnection(baseURL, secret string) error {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, err := http.NewRequest("GET", baseURL+"/proxies", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("authentication failed: invalid secret")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// AutoDetect attempts to detect Clash API on common ports
func AutoDetect() *Config {
	defaultURLs := []string{
		"http://127.0.0.1:9090",
		"http://localhost:9090",
	}

	for _, url := range defaultURLs {
		if err := TestConnection(url, ""); err == nil {
			return &Config{
				URL:    url,
				Secret: "",
			}
		}
	}

	return nil
}
