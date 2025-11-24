package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/matinhimself/singbox-web-config/internal/types"
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

// Config is an alias to the generated type-safe Config type
type Config = types.Config

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
				Route: &types.RouteOptions{
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
		config.Route = &types.RouteOptions{}
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

// UpdateOutbounds updates the outbounds in the config
func (m *Manager) UpdateOutbounds(outbounds []interface{}) error {
	// Load current config
	config, err := m.LoadConfig()
	if err != nil {
		return err
	}

	// Update outbounds
	config.Outbounds = outbounds

	// Save config
	return m.SaveConfig(config)
}

// GetOutbounds returns the current outbounds
func (m *Manager) GetOutbounds() ([]interface{}, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, err
	}

	if config.Outbounds == nil {
		return []interface{}{}, nil
	}

	return config.Outbounds, nil
}

// GetOutboundTags returns a list of all outbound tags
func (m *Manager) GetOutboundTags() ([]string, error) {
	outbounds, err := m.GetOutbounds()
	if err != nil {
		return nil, err
	}

	var tags []string
	for _, outbound := range outbounds {
		if outboundMap, ok := outbound.(map[string]interface{}); ok {
			if tag, ok := outboundMap["tag"].(string); ok {
				tags = append(tags, tag)
			}
		}
	}

	return tags, nil
}

// RenameOutbound renames an outbound and updates all references to it
func (m *Manager) RenameOutbound(oldTag, newTag string) error {
	config, err := m.LoadConfig()
	if err != nil {
		return err
	}

	// Update outbound tag
	for _, outbound := range config.Outbounds {
		if outboundMap, ok := outbound.(map[string]interface{}); ok {
			if tag, ok := outboundMap["tag"].(string); ok && tag == oldTag {
				outboundMap["tag"] = newTag
			}

			// Update references in selector/urltest outbounds
			if outbounds, ok := outboundMap["outbounds"].([]interface{}); ok {
				for i, ob := range outbounds {
					if obTag, ok := ob.(string); ok && obTag == oldTag {
						outbounds[i] = newTag
					}
				}
			}
		}
	}

	// Update references in route rules
	if config.Route != nil && config.Route.Rules != nil {
		for _, rule := range config.Route.Rules {
			if ruleMap, ok := rule.(map[string]interface{}); ok {
				// Update outbound field
				if outbound, ok := ruleMap["outbound"].(string); ok && outbound == oldTag {
					ruleMap["outbound"] = newTag
				}
			}
		}

		// Update final outbound
		if config.Route.Final == oldTag {
			config.Route.Final = newTag
		}
	}

	return m.SaveConfig(config)
}
