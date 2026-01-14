package clipboard

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"image"
	"image/png"
	"sync"
	"time"

	"pano/internal/storage"

	"github.com/atotto/clipboard"
)

// Monitor handles clipboard monitoring
type Monitor struct {
	db            *storage.Database
	lastTextHash  []byte
	lastImageHash []byte
	running       bool
	mu            sync.Mutex
	onChange      func(itemType string, content []byte)
	onLimitWarn   func(remaining int)
	pollInterval  time.Duration
}

// NewMonitor creates a new clipboard monitor
func NewMonitor(db *storage.Database) *Monitor {
	return &Monitor{
		db:           db,
		pollInterval: 200 * time.Millisecond, // Faster polling
		running:      false,
	}
}

// SetOnChange sets the callback function for clipboard changes
func (m *Monitor) SetOnChange(callback func(itemType string, content []byte)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onChange = callback
}

// SetOnLimitWarn sets the callback for limit warnings
func (m *Monitor) SetOnLimitWarn(callback func(remaining int)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onLimitWarn = callback
}

// Start begins monitoring the clipboard
func (m *Monitor) Start() error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("monitor already running")
	}
	m.running = true
	m.mu.Unlock()

	go m.monitorLoop()
	return nil
}

// Stop stops monitoring the clipboard
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.running = false
}

// monitorLoop continuously checks for clipboard changes
func (m *Monitor) monitorLoop() {
	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			running := m.running
			m.mu.Unlock()

			if !running {
				return
			}

			m.checkClipboard()
		}
	}
}

// checkClipboard checks if clipboard content has changed
func (m *Monitor) checkClipboard() {
	// Try to read image first (if available)
	// Images are checked first because text might be empty but image could be present
	if img, err := ReadClipboardImage(); err == nil && img != nil {
		m.handleImage(img)
		return
	}

	// Try to read text
	text, err := clipboard.ReadAll()
	if err == nil && text != "" {
		m.handleText(text)
		return
	}
}

// handleText processes new text content
func (m *Monitor) handleText(text string) {
	content := []byte(text)
	hash := sha256.Sum256(content)

	// Check if content has changed
	if bytes.Equal(hash[:], m.lastTextHash) {
		return
	}

	m.lastTextHash = hash[:]

	// Add to database
	err := m.db.AddItem("text", content)

	// Check for limit warnings
	m.mu.Lock()
	limitCallback := m.onLimitWarn
	changeCallback := m.onChange
	m.mu.Unlock()

	if err != nil {
		errStr := err.Error()
		if len(errStr) >= 10 && errStr[:10] == "LIMIT_FULL" {
			if limitCallback != nil {
				go limitCallback(0)
			}
			return
		} else if len(errStr) >= 10 && errStr[:10] == "LIMIT_WARN" {
			var remaining int
			fmt.Sscanf(errStr, "LIMIT_WARN:%d", &remaining)
			if limitCallback != nil {
				go limitCallback(remaining)
			}
			// Continue to trigger onChange since item was added
		} else {
			return // Silently ignore other errors
		}
	}

	if changeCallback != nil {
		changeCallback("text", content)
	}
}

// handleImage processes new image content
func (m *Monitor) handleImage(img image.Image) {
	// Convert image to PNG bytes
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		fmt.Printf("Error encoding image: %v\n", err)
		return
	}

	content := buf.Bytes()
	hash := sha256.Sum256(content)

	// Check if content has changed
	if bytes.Equal(hash[:], m.lastImageHash) {
		return
	}

	m.lastImageHash = hash[:]

	// Add to database
	err := m.db.AddItem("image", content)

	// Check for limit warnings
	m.mu.Lock()
	limitCallback := m.onLimitWarn
	changeCallback := m.onChange
	m.mu.Unlock()

	if err != nil {
		errStr := err.Error()
		if len(errStr) >= 10 && errStr[:10] == "LIMIT_FULL" {
			if limitCallback != nil {
				go limitCallback(0)
			}
			return
		} else if len(errStr) >= 10 && errStr[:10] == "LIMIT_WARN" {
			var remaining int
			fmt.Sscanf(errStr, "LIMIT_WARN:%d", &remaining)
			if limitCallback != nil {
				go limitCallback(remaining)
			}
		} else {
			return // Silently ignore other errors
		}
	}

	if changeCallback != nil {
		changeCallback("image", content)
	}
}
