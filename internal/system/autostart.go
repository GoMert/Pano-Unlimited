package system

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const (
	registryPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	appName      = "Pano"
)

// AutostartManager handles Windows startup registration
type AutostartManager struct {
	exePath string
}

// NewAutostartManager creates a new autostart manager
func NewAutostartManager() (*AutostartManager, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	return &AutostartManager{
		exePath: exePath,
	}, nil
}

// IsEnabled checks if autostart is enabled
func (a *AutostartManager) IsEnabled() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.QUERY_VALUE)
	if err != nil {
		return false, nil // Key doesn't exist, autostart not enabled
	}
	defer key.Close()

	_, _, err = key.GetStringValue(appName)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// Enable adds the application to Windows startup
func (a *AutostartManager) Enable() error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	// Use quoted path to handle spaces
	quotedPath := fmt.Sprintf(`"%s"`, filepath.Clean(a.exePath))

	if err := key.SetStringValue(appName, quotedPath); err != nil {
		return fmt.Errorf("failed to set registry value: %w", err)
	}

	return nil
}

// Disable removes the application from Windows startup
func (a *AutostartManager) Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key: %w", err)
	}
	defer key.Close()

	if err := key.DeleteValue(appName); err != nil {
		if err == registry.ErrNotExist {
			return nil // Already disabled
		}
		return fmt.Errorf("failed to delete registry value: %w", err)
	}

	return nil
}

// Toggle toggles the autostart status
func (a *AutostartManager) Toggle() error {
	enabled, err := a.IsEnabled()
	if err != nil {
		return err
	}

	if enabled {
		return a.Disable()
	}
	return a.Enable()
}
