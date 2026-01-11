//go:build !windows
// +build !windows

package clipboard

import (
	"fmt"
	"image"
)

// ReadClipboardImage is a stub for non-Windows platforms
func ReadClipboardImage() (image.Image, error) {
	return nil, fmt.Errorf("image clipboard support is only available on Windows")
}
