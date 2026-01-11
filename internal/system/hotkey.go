package system

import (
	"fmt"
	"sync"

	hook "github.com/robotn/gohook"
)

// HotkeyManager handles global hotkey registration
type HotkeyManager struct {
	callback func()
	running  bool
	mu       sync.Mutex
}

// NewHotkeyManager creates a new hotkey manager
func NewHotkeyManager() *HotkeyManager {
	return &HotkeyManager{
		running: false,
	}
}

// SetCallback sets the function to call when hotkey is pressed
func (h *HotkeyManager) SetCallback(callback func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.callback = callback
}

// Start registers the global hotkey (Ctrl+Shift+V)
func (h *HotkeyManager) Start() error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return fmt.Errorf("hotkey already registered")
	}
	h.running = true
	h.mu.Unlock()

	go h.listenForHotkey()
	return nil
}

// Stop unregisters the global hotkey
func (h *HotkeyManager) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.running = false
	hook.End()
}

// listenForHotkey listens for Ctrl+Shift+V combination
func (h *HotkeyManager) listenForHotkey() {
	// Modifier key state tracking
	ctrlPressed := false
	shiftPressed := false

	// Create event channel
	evChan := hook.Start()
	defer hook.End()

	for ev := range evChan {
		// Check if we should stop
		h.mu.Lock()
		running := h.running
		h.mu.Unlock()

		if !running {
			return
		}

		if ev.Kind == hook.KeyDown {
			// Track Ctrl key (scan codes: 29 left, 3613 right OR virtual key codes: 162 left, 163 right)
			if ev.Rawcode == 29 || ev.Rawcode == 3613 || ev.Rawcode == 162 || ev.Rawcode == 163 || ev.Rawcode == 17 {
				ctrlPressed = true
			}
			// Track Shift key (scan codes: 42 left, 54 right OR virtual key codes: 160 left, 161 right)
			if ev.Rawcode == 42 || ev.Rawcode == 54 || ev.Rawcode == 160 || ev.Rawcode == 161 || ev.Rawcode == 16 {
				shiftPressed = true
			}
			// Check for V key (scan code: 47, virtual key: 86)
			if (ev.Rawcode == 47 || ev.Rawcode == 86) && ctrlPressed && shiftPressed {
				// Ctrl+Shift+V detected - trigger callback
				h.mu.Lock()
				callback := h.callback
				h.mu.Unlock()

				if callback != nil {
					go callback() // Run in goroutine to avoid blocking
				}
			}
		} else if ev.Kind == hook.KeyUp {
			// Reset Ctrl state when Ctrl key is released
			if ev.Rawcode == 29 || ev.Rawcode == 3613 || ev.Rawcode == 162 || ev.Rawcode == 163 || ev.Rawcode == 17 {
				ctrlPressed = false
			}
			// Reset Shift state when Shift key is released
			if ev.Rawcode == 42 || ev.Rawcode == 54 || ev.Rawcode == 160 || ev.Rawcode == 161 || ev.Rawcode == 16 {
				shiftPressed = false
			}
		}
	}
}
