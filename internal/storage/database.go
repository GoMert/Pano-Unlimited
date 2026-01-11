package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	MaxItems     = 100              // Maximum number of clipboard items
	MaxItemSize  = 20 * 1024 * 1024 // 20MB per item
	DatabaseFile = "clipboard.db"
)

// ClipboardItem represents a single clipboard entry
type ClipboardItem struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`    // "text" or "image"
	Content   string    `json:"content"` // Encrypted content
	Timestamp time.Time `json:"timestamp"`
	Pinned    bool      `json:"pinned"`
	Size      int       `json:"size"` // Original size in bytes
}

// Database manages clipboard items storage
type Database struct {
	Items []ClipboardItem `json:"items"`
	key   []byte          // Encryption key (not stored in JSON)
}

// NewDatabase creates or loads the database
func NewDatabase() (*Database, error) {
	key, err := GetHardwareKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get hardware key: %w", err)
	}

	db := &Database{
		Items: make([]ClipboardItem, 0),
		key:   key,
	}

	// Try to load existing database
	if err := db.Load(); err != nil {
		// If file doesn't exist, that's okay - we'll create it on first save
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return db, nil
}

// GetDatabasePath returns the full path to the database file
func GetDatabasePath() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("APPDATA environment variable not set")
	}

	panoDir := filepath.Join(appData, "Pano")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(panoDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create Pano directory: %w", err)
	}

	return filepath.Join(panoDir, DatabaseFile), nil
}

// Load loads the database from disk
func (db *Database) Load() error {
	dbPath, err := GetDatabasePath()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(dbPath)
	if err != nil {
		return err
	}

	// Decrypt the entire database
	decrypted, err := Decrypt(string(data), db.key)
	if err != nil {
		return fmt.Errorf("failed to decrypt database: %w", err)
	}

	// Parse JSON
	if err := json.Unmarshal(decrypted, &db.Items); err != nil {
		return fmt.Errorf("failed to parse database: %w", err)
	}

	return nil
}

// Save saves the database to disk
func (db *Database) Save() error {
	dbPath, err := GetDatabasePath()
	if err != nil {
		return err
	}

	// Convert to JSON
	jsonData, err := json.Marshal(db.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal database: %w", err)
	}

	// Encrypt the entire database
	encrypted, err := Encrypt(jsonData, db.key)
	if err != nil {
		return fmt.Errorf("failed to encrypt database: %w", err)
	}

	// Write to file
	if err := os.WriteFile(dbPath, []byte(encrypted), 0600); err != nil {
		return fmt.Errorf("failed to write database: %w", err)
	}

	return nil
}

// AddItem adds a new clipboard item
func (db *Database) AddItem(itemType string, content []byte) error {
	// Check size limit
	if len(content) > MaxItemSize {
		return fmt.Errorf("item size (%d bytes) exceeds maximum (%d bytes)", len(content), MaxItemSize)
	}

	// Encrypt content
	encrypted, err := Encrypt(content, db.key)
	if err != nil {
		return fmt.Errorf("failed to encrypt content: %w", err)
	}

	// Create new item
	item := ClipboardItem{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:      itemType,
		Content:   encrypted,
		Timestamp: time.Now(),
		Pinned:    false,
		Size:      len(content),
	}

	// Add to beginning of list
	db.Items = append([]ClipboardItem{item}, db.Items...)

	// Remove oldest unpinned items if we exceed the limit
	db.enforceLimit()

	// Save to disk
	return db.Save()
}

// enforceLimit removes oldest unpinned items to stay within MaxItems
func (db *Database) enforceLimit() {
	if len(db.Items) <= MaxItems {
		return
	}

	// Separate pinned and unpinned items
	pinnedItems := make([]ClipboardItem, 0)
	unpinnedItems := make([]ClipboardItem, 0)

	for _, item := range db.Items {
		if item.Pinned {
			pinnedItems = append(pinnedItems, item)
		} else {
			unpinnedItems = append(unpinnedItems, item)
		}
	}

	// If pinned items exceed MaxItems, keep only the newest pinned items
	// This shouldn't happen normally, but handle it as a safeguard
	if len(pinnedItems) > MaxItems {
		pinnedItems = pinnedItems[:MaxItems]
	}

	// Calculate how many unpinned items we can keep
	availableSlots := MaxItems - len(pinnedItems)

	// Keep the newest unpinned items (they are already in order from newest to oldest)
	if len(unpinnedItems) > availableSlots {
		unpinnedItems = unpinnedItems[:availableSlots]
	}

	// Combine: pinned items first, then unpinned items
	db.Items = append(pinnedItems, unpinnedItems...)
}

// GetItem retrieves and decrypts an item by ID
func (db *Database) GetItem(id string) (*ClipboardItem, []byte, error) {
	for i, item := range db.Items {
		if item.ID == id {
			// Decrypt content
			decrypted, err := Decrypt(item.Content, db.key)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decrypt item: %w", err)
			}
			return &db.Items[i], decrypted, nil
		}
	}
	return nil, nil, fmt.Errorf("item not found")
}

// TogglePin toggles the pinned status of an item
func (db *Database) TogglePin(id string) error {
	for i, item := range db.Items {
		if item.ID == id {
			db.Items[i].Pinned = !item.Pinned
			return db.Save()
		}
	}
	return fmt.Errorf("item not found")
}

// DeleteItem removes an item from the database
func (db *Database) DeleteItem(id string) error {
	for i, item := range db.Items {
		if item.ID == id {
			db.Items = append(db.Items[:i], db.Items[i+1:]...)
			return db.Save()
		}
	}
	return fmt.Errorf("item not found")
}

// GetAllItems returns all items (metadata only, no decrypted content)
func (db *Database) GetAllItems() []ClipboardItem {
	return db.Items
}

// ClearAll removes all items from the database
func (db *Database) ClearAll() error {
	db.Items = make([]ClipboardItem, 0)
	return db.Save()
}

// GetItemCount returns the number of items in the database
func (db *Database) GetItemCount() int {
	return len(db.Items)
}

// GetPinnedCount returns the number of pinned items
func (db *Database) GetPinnedCount() int {
	count := 0
	for _, item := range db.Items {
		if item.Pinned {
			count++
		}
	}
	return count
}
