//go:build !windows
// +build !windows

package ui

// BringWindowToFront is a no-op on non-Windows platforms
func BringWindowToFront(windowTitle string) {
	// No-op on non-Windows
}
