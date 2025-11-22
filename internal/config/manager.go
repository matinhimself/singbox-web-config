package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	Rules      []interface{} `json:"rules,omitempty"`
	RuleAction []interface{} `json:"rule_action,omitempty"`
	Final      string        `json:"final,omitempty"`
}

// BackupMetadata stores information about a backup
type BackupMetadata struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	ConfigFile  string    `json:"config_file"`
	Version     string    `json:"version,omitempty"`
}

// BackupInfo combines backup filename with its metadata
type BackupInfo struct {
	Filename string
	Metadata BackupMetadata
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

	timestamp := time.Now()
	name := fmt.Sprintf("Auto backup %s", timestamp.Format("2006-01-02 15:04:05"))
	return m.CreateBackupWithName(name, "Automatic backup")
}

// CreateBackupWithName creates a backup with a custom name and metadata
func (m *Manager) CreateBackupWithName(name, description string) error {
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
	timestamp := time.Now()
	// Sanitize name for filename
	safeName := sanitizeFilename(name)
	if safeName == "" {
		safeName = "backup"
	}
	backupFilename := fmt.Sprintf("%s-%s.json", safeName, timestamp.Format("20060102-150405"))
	backupPath := filepath.Join(m.backupDir, backupFilename)

	// Write backup
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	// Create metadata
	metadata := BackupMetadata{
		Name:        name,
		Description: description,
		Timestamp:   timestamp,
		ConfigFile:  backupFilename,
		Version:     "1.0", // You can update this to track config version
	}

	// Write metadata file
	metadataPath := filepath.Join(m.backupDir, backupFilename+".meta")
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, metadataJSON, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	// Replace invalid characters with dash
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "-")
	}
	// Replace spaces with dash
	result = strings.ReplaceAll(result, " ", "-")
	// Remove leading/trailing dashes
	result = strings.Trim(result, "-")
	return result
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

// ListBackups returns a list of available backups sorted by timestamp (newest first)
func (m *Manager) ListBackups() ([]BackupInfo, error) {
	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			backupInfo := BackupInfo{
				Filename: entry.Name(),
			}

			// Try to load metadata
			metadataPath := filepath.Join(m.backupDir, entry.Name()+".meta")
			if metadataData, err := os.ReadFile(metadataPath); err == nil {
				var metadata BackupMetadata
				if err := json.Unmarshal(metadataData, &metadata); err == nil {
					backupInfo.Metadata = metadata
				}
			}

			// If no metadata, create default from filename
			if backupInfo.Metadata.ConfigFile == "" {
				info, _ := entry.Info()
				backupInfo.Metadata = BackupMetadata{
					Name:       entry.Name(),
					Timestamp:  info.ModTime(),
					ConfigFile: entry.Name(),
				}
			}

			backups = append(backups, backupInfo)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Metadata.Timestamp.After(backups[j].Metadata.Timestamp)
	})

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
