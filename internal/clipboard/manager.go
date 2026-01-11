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
