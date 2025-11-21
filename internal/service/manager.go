package service

import (
	"fmt"
	"os/exec"
	"strings"
)

// Manager manages the sing-box systemd service
type Manager struct {
	serviceName string
}

// NewManager creates a new service manager
func NewManager(serviceName string) *Manager {
	return &Manager{
		serviceName: serviceName,
	}
}

// Status represents service status
type Status struct {
	Active    bool
	Running   bool
	Enabled   bool
	Message   string
}

// GetStatus returns the current status of the service
func (m *Manager) GetStatus() (*Status, error) {
	cmd := exec.Command("systemctl", "is-active", m.serviceName)
	output, _ := cmd.Output()
	isActive := strings.TrimSpace(string(output)) == "active"

	cmd = exec.Command("systemctl", "is-enabled", m.serviceName)
	output, _ = cmd.Output()
	isEnabled := strings.TrimSpace(string(output)) == "enabled"

	// Get detailed status
	cmd = exec.Command("systemctl", "status", m.serviceName)
	statusOutput, _ := cmd.CombinedOutput()

	return &Status{
		Active:  isActive,
		Running: isActive,
		Enabled: isEnabled,
		Message: string(statusOutput),
	}, nil
}

// Start starts the service
func (m *Manager) Start() error {
	cmd := exec.Command("systemctl", "start", m.serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %w, output: %s", err, output)
	}
	return nil
}

// Stop stops the service
func (m *Manager) Stop() error {
	cmd := exec.Command("systemctl", "stop", m.serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop service: %w, output: %s", err, output)
	}
	return nil
}

// Restart restarts the service
func (m *Manager) Restart() error {
	cmd := exec.Command("systemctl", "restart", m.serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to restart service: %w, output: %s", err, output)
	}
	return nil
}

// Reload reloads the service configuration
func (m *Manager) Reload() error {
	cmd := exec.Command("systemctl", "reload-or-restart", m.serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload service: %w, output: %s", err, output)
	}
	return nil
}

// Enable enables the service to start on boot
func (m *Manager) Enable() error {
	cmd := exec.Command("systemctl", "enable", m.serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to enable service: %w, output: %s", err, output)
	}
	return nil
}

// Disable disables the service from starting on boot
func (m *Manager) Disable() error {
	cmd := exec.Command("systemctl", "disable", m.serviceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to disable service: %w, output: %s", err, output)
	}
	return nil
}

// GetLogs returns recent service logs
func (m *Manager) GetLogs(lines int) (string, error) {
	cmd := exec.Command("journalctl", "-u", m.serviceName, "-n", fmt.Sprintf("%d", lines), "--no-pager")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}
	return string(output), nil
}
