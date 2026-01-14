package clipboard

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"pano/internal/storage"

	"github.com/atotto/clipboard"
)

// decodePNGImage decodes PNG bytes to image.Image
func decodePNGImage(data []byte) (image.Image, error) {
	return png.Decode(bytes.NewReader(data))
}

// Manager handles clipboard operations
type Manager struct {
	db *storage.Database
}

// NewManager creates a new clipboard manager
func NewManager(db *storage.Database) *Manager {
	return &Manager{
		db: db,
	}
}

// CopyToClipboard copies an item to the system clipboard
func (m *Manager) CopyToClipboard(id string) error {
	item, content, err := m.db.GetItem(id)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	switch item.Type {
	case "text":
		if err := clipboard.WriteAll(string(content)); err != nil {
			return fmt.Errorf("failed to write to clipboard: %w", err)
		}
	case "image":
		// Decode PNG image and write to clipboard
		img, err := decodePNGImage(content)
		if err != nil {
			return fmt.Errorf("failed to decode image: %w", err)
		}
		if err := WriteClipboardImage(img); err != nil {
			return fmt.Errorf("failed to write image to clipboard: %w", err)
		}
	default:
		return fmt.Errorf("unknown item type: %s", item.Type)
	}

	return nil
}

// PinItem toggles the pinned status of an item
func (m *Manager) PinItem(id string) error {
	return m.db.TogglePin(id)
}

// DeleteItem removes an item from the database
func (m *Manager) DeleteItem(id string) error {
	return m.db.DeleteItem(id)
}

// GetAllItems returns all clipboard items
func (m *Manager) GetAllItems() []storage.ClipboardItem {
	return m.db.GetAllItems()
}

// GetItemContent retrieves the decrypted content of an item
func (m *Manager) GetItemContent(id string) ([]byte, error) {
	_, content, err := m.db.GetItem(id)
	return content, err
}

// ClearAll removes all items from the database
func (m *Manager) ClearAll() error {
	return m.db.ClearAll()
}

// GetItemCount returns the number of items
func (m *Manager) GetItemCount() int {
	return m.db.GetItemCount()
}

// GetPinnedCount returns the number of pinned items
func (m *Manager) GetPinnedCount() int {
	return m.db.GetPinnedCount()
}

// SetMaxItems sets the maximum number of items
func (m *Manager) SetMaxItems(max int) {
	m.db.SetMaxItems(max)
}

// GetMaxItems returns the current maximum items limit
func (m *Manager) GetMaxItems() int {
	return m.db.GetMaxItems()
}

// IsNearLimit returns true if item count is within 10 of the limit
func (m *Manager) IsNearLimit() bool {
	return m.db.IsNearLimit()
}

// GetRemainingSlots returns how many more items can be added
func (m *Manager) GetRemainingSlots() int {
	return m.db.GetRemainingSlots()
}

// SetOnLimitWarn sets callback for limit warning
func (m *Manager) SetOnLimitWarn(callback func(remaining int)) {
	m.db.SetOnLimitWarn(callback)
}

// IsFull returns true if at or over limit
func (m *Manager) IsFull() bool {
	return m.db.IsFull()
}
