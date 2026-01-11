package storage

import (
	"crypto/sha256"
	"fmt"
	"runtime"

	"github.com/denisbrodbeck/machineid"
)

// GetHardwareKey generates a unique SHA-256 key based on hardware information
// This key is used for encrypting clipboard data
func GetHardwareKey() ([]byte, error) {
	// Get machine ID (based on CPU, motherboard, etc.)
	machineID, err := machineid.ProtectedID("pano-clipboard")
	if err != nil {
		return nil, fmt.Errorf("failed to get machine ID: %w", err)
	}

	// Combine machine ID with OS and architecture for additional uniqueness
	combined := fmt.Sprintf("%s-%s-%s", machineID, runtime.GOOS, runtime.GOARCH)

	// Generate SHA-256 hash
	hash := sha256.Sum256([]byte(combined))

	return hash[:], nil
}

// GetKeyFingerprint returns a human-readable fingerprint of the hardware key
// This can be used for debugging (first 8 chars only)
func GetKeyFingerprint() (string, error) {
	key, err := GetHardwareKey()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", key[:4]), nil
}
