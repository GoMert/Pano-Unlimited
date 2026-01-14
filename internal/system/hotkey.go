package system

import (
	"fmt"
	"sync"

	hook "github.com/robotn/gohook"
)

// Key codes for Windows
const (
	// Ctrl key codes (scan codes and virtual key codes)
	scCtrlLeft   = 29
	scCtrlRight  = 3613
	vkCtrlLeft   = 162
	vkCtrlRight  = 163
	vkCtrl       = 17

	// Shift key codes (scan codes and virtual key codes)
	scShiftLeft  = 42
	scShiftRight = 54
	vkShiftLeft  = 160
	vkShiftRight = 161
	vkShift      = 16

	// V key codes
	scV = 47
	vkV = 86
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

// isCtrlKey checks if the rawcode is a Ctrl key
func isCtrlKey(rawcode uint16) bool {
	return rawcode == scCtrlLeft || rawcode == scCtrlRight ||
		rawcode == vkCtrlLeft || rawcode == vkCtrlRight || rawcode == vkCtrl
}

// isShiftKey checks if the rawcode is a Shift key
func isShiftKey(rawcode uint16) bool {
	return rawcode == scShiftLeft || rawcode == scShiftRight ||
		rawcode == vkShiftLeft || rawcode == vkShiftRight || rawcode == vkShift
}

// isVKey checks if the rawcode is the V key
func isVKey(rawcode uint16) bool {
	return rawcode == scV || rawcode == vkV
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
			// Track Ctrl key
			if isCtrlKey(ev.Rawcode) {
				ctrlPressed = true
			}
			// Track Shift key
			if isShiftKey(ev.Rawcode) {
				shiftPressed = true
			}
			// Check for V key with modifiers
			if isVKey(ev.Rawcode) && ctrlPressed && shiftPressed {
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
			if isCtrlKey(ev.Rawcode) {
				ctrlPressed = false
			}
			// Reset Shift state when Shift key is released
			if isShiftKey(ev.Rawcode) {
				shiftPressed = false
			}
		}
	}
}
