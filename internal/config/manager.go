package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manager handles sing-box configuration management
type Manager struct {
	configPath string
	backupDir  string
}

// NewManager creates a new config manager
func NewManager(configPath string) (*Manager, error) {
	backupDir := filepath.Join(filepath.Dir(configPath), "backups")

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	return &Manager{
		configPath: configPath,
		backupDir:  backupDir,
	}, nil
}

// Config represents a sing-box configuration
type Config struct {
	Log       *LogConfig       `json:"log,omitempty"`
	DNS       *DNSConfig       `json:"dns,omitempty"`
	Inbounds  []interface{}    `json:"inbounds,omitempty"`
	Outbounds []interface{}    `json:"outbounds,omitempty"`
	Route     *RouteConfig     `json:"route,omitempty"`
}

type LogConfig struct {
	Level      string `json:"level,omitempty"`
	Output     string `json:"output,omitempty"`
	Timestamp  bool   `json:"timestamp,omitempty"`
}

type DNSConfig struct {
	Servers []interface{} `json:"servers,omitempty"`
	Rules   []interface{} `json:"rules,omitempty"`
	Final   string        `json:"final,omitempty"`
}

type RouteConfig struct {
	Rules []interface{} `json:"rules,omitempty"`
	Final string        `json:"final,omitempty"`
}

// LoadConfig loads the current configuration
func (m *Manager) LoadConfig() (*Config, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return &Config{
				Route: &RouteConfig{
					Rules: []interface{}{},
					Final: "direct",
				},
			}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration with backup
func (m *Manager) SaveConfig(config *Config) error {
	// Create backup first
	if err := m.BackupConfig(); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// BackupConfig creates a backup of the current configuration
func (m *Manager) BackupConfig() error {
	// Check if config exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return nil // No config to backup
	}

	// Read current config
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	// Create backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(m.backupDir, fmt.Sprintf("config-%s.json", timestamp))

	// Write backup
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// UpdateRules updates only the routing rules in the config
func (m *Manager) UpdateRules(rules []interface{}) error {
	// Load current config
	config, err := m.LoadConfig()
	if err != nil {
		return err
	}

	// Ensure route section exists
	if config.Route == nil {
		config.Route = &RouteConfig{}
	}

	// Update rules
	config.Route.Rules = rules

	// Save config
	return m.SaveConfig(config)
}

// GetRules returns the current routing rules
func (m *Manager) GetRules() ([]interface{}, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, err
	}

	if config.Route == nil {
		return []interface{}{}, nil
	}

	return config.Route.Rules, nil
}

// ListBackups returns a list of available backups
func (m *Manager) ListBackups() ([]string, error) {
	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			backups = append(backups, entry.Name())
		}
	}

	return backups, nil
}

// RestoreBackup restores a configuration from a backup
func (m *Manager) RestoreBackup(backupName string) error {
	backupPath := filepath.Join(m.backupDir, backupName)

	// Read backup
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	// Validate JSON
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("invalid backup file: %w", err)
	}

	// Create backup of current config before restoring
	if err := m.BackupConfig(); err != nil {
		return fmt.Errorf("failed to backup current config: %w", err)
	}

	// Write restored config
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to restore config: %w", err)
	}

	return nil
}
