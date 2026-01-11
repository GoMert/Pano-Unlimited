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
	pollInterval  time.Duration
}

// NewMonitor creates a new clipboard monitor
func NewMonitor(db *storage.Database) *Monitor {
	return &Monitor{
		db:           db,
		pollInterval: 500 * time.Millisecond,
		running:      false,
	}
}

// SetOnChange sets the callback function for clipboard changes
func (m *Monitor) SetOnChange(callback func(itemType string, content []byte)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onChange = callback
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
		m.mu.Lock()
		if !m.running {
			m.mu.Unlock()
			return
		}
		m.mu.Unlock()

		m.checkClipboard()
		<-ticker.C
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
	if err := m.db.AddItem("text", content); err != nil {
		fmt.Printf("Error adding clipboard item: %v\n", err)
		return
	}

	// Trigger callback
	m.mu.Lock()
	callback := m.onChange
	m.mu.Unlock()

	if callback != nil {
		callback("text", content)
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
	if err := m.db.AddItem("image", content); err != nil {
		fmt.Printf("Error adding clipboard item: %v\n", err)
		return
	}

	// Trigger callback
	m.mu.Lock()
	callback := m.onChange
	m.mu.Unlock()

	if callback != nil {
		callback("image", content)
	}
}
