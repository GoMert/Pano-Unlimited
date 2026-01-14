package storage

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	DefaultMaxItems = 100              // Default maximum number of clipboard items
	MaxItemSize     = 20 * 1024 * 1024 // 20MB per item
	DatabaseFile    = "clipboard.db"
)

// ClipboardItem represents a single clipboard entry
type ClipboardItem struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`    // "text" or "image"
	Content   string    `json:"content"` // Encrypted content
	Timestamp time.Time `json:"timestamp"`
	Pinned    bool      `json:"pinned"`
	Size      int       `json:"size"` // Original size in bytes
	Hash      string    `json:"hash"` // Content hash for duplicate detection
}

// Database manages clipboard items storage
type Database struct {
	Items       []ClipboardItem     `json:"items"`
	key         []byte              // Encryption key (not stored in JSON)
	mu          sync.RWMutex        // Mutex for thread-safe operations
	maxItems    int                 // Configurable max items limit
	onLimitWarn func(remaining int) // Callback when near limit
}

// NewDatabase creates or loads the database
func NewDatabase() (*Database, error) {
	key, err := GetHardwareKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get hardware key: %w", err)
	}

	db := &Database{
		Items:    make([]ClipboardItem, 0),
		key:      key,
		maxItems: DefaultMaxItems,
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

// SetMaxItems sets the maximum number of items
func (db *Database) SetMaxItems(max int) {
	db.mu.Lock()
	defer db.mu.Unlock()
	if max < 10 {
		max = 10
	}
	if max > 500 {
		max = 500
	}
	db.maxItems = max
	db.enforceLimit()
	db.saveInternal()
}

// GetMaxItems returns the current maximum items limit
func (db *Database) GetMaxItems() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.maxItems
}

// IsNearLimit returns true if item count is within 10 of the limit
func (db *Database) IsNearLimit() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.Items) >= db.maxItems-10
}

// GetRemainingSlots returns how many more items can be added
func (db *Database) GetRemainingSlots() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	remaining := db.maxItems - len(db.Items)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// SetOnLimitWarn sets callback for limit warning
func (db *Database) SetOnLimitWarn(callback func(remaining int)) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.onLimitWarn = callback
}

// IsFull returns true if at or over limit
func (db *Database) IsFull() bool {
	db.mu.RLock()
	defer db.mu.RUnlock()
	// Count unpinned items only (pinned don't count toward limit)
	unpinnedCount := 0
	for _, item := range db.Items {
		if !item.Pinned {
			unpinnedCount++
		}
	}
	return unpinnedCount >= db.maxItems
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
	db.mu.Lock()
	defer db.mu.Unlock()

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

// Save saves the database to disk (thread-safe)
func (db *Database) Save() error {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.saveInternal()
}

// saveInternal saves the database without locking (caller must hold lock)
func (db *Database) saveInternal() error {
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
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check size limit
	if len(content) > MaxItemSize {
		return fmt.Errorf("item size (%d bytes) exceeds maximum (%d bytes)", len(content), MaxItemSize)
	}

	// Calculate content hash for duplicate detection
	contentHash := fmt.Sprintf("%x", sha256.Sum256(content))

	// Check for duplicate (same content already exists)
	for i, existing := range db.Items {
		if existing.Hash == contentHash && existing.Type == itemType {
			// Move existing item to top instead of creating duplicate
			db.Items = append([]ClipboardItem{existing}, append(db.Items[:i], db.Items[i+1:]...)...)
			db.Items[0].Timestamp = time.Now()
			return db.saveInternal()
		}
	}

	// Count current unpinned items
	unpinnedCount := 0
	for _, item := range db.Items {
		if !item.Pinned {
			unpinnedCount++
		}
	}

	// Check if we're at the limit - don't add new items if full
	if unpinnedCount >= db.maxItems {
		return fmt.Errorf("LIMIT_FULL:0")
	}

	// Calculate remaining slots for warning
	remaining := db.maxItems - unpinnedCount - 1
	warnNeeded := remaining <= 10 && remaining >= 0

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
		Hash:      contentHash,
	}

	// Add to beginning of list
	db.Items = append([]ClipboardItem{item}, db.Items...)

	// Save to disk
	if err := db.saveInternal(); err != nil {
		return err
	}

	// Return warning signal if near limit
	if warnNeeded {
		return fmt.Errorf("LIMIT_WARN:%d", remaining)
	}
	return nil
}

// enforceLimit removes oldest unpinned items to stay within maxItems
func (db *Database) enforceLimit() {
	if len(db.Items) <= db.maxItems {
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

	// If pinned items exceed maxItems, keep only the newest pinned items
	if len(pinnedItems) > db.maxItems {
		pinnedItems = pinnedItems[:db.maxItems]
	}

	// Calculate how many unpinned items we can keep
	availableSlots := db.maxItems - len(pinnedItems)

	// Keep the newest unpinned items
	if len(unpinnedItems) > availableSlots {
		unpinnedItems = unpinnedItems[:availableSlots]
	}

	// Combine: pinned items first, then unpinned items
	db.Items = append(pinnedItems, unpinnedItems...)
}

// GetItem retrieves and decrypts an item by ID
func (db *Database) GetItem(id string) (*ClipboardItem, []byte, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	for i, item := range db.Items {
		if item.ID == id {
			// Decrypt content
			decrypted, err := Decrypt(item.Content, db.key)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to decrypt item: %w", err)
			}
			// Return a copy to avoid race conditions
			itemCopy := db.Items[i]
			return &itemCopy, decrypted, nil
		}
	}
	return nil, nil, fmt.Errorf("item not found")
}

// TogglePin toggles the pinned status of an item
func (db *Database) TogglePin(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i, item := range db.Items {
		if item.ID == id {
			db.Items[i].Pinned = !item.Pinned
			return db.saveInternal()
		}
	}
	return fmt.Errorf("item not found")
}

// DeleteItem removes an item from the database
func (db *Database) DeleteItem(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i, item := range db.Items {
		if item.ID == id {
			db.Items = append(db.Items[:i], db.Items[i+1:]...)
			return db.saveInternal()
		}
	}
	return fmt.Errorf("item not found")
}

// GetAllItems returns all items (metadata only, no decrypted content)
// Pinned items are returned first, then unpinned items by timestamp
func (db *Database) GetAllItems() []ClipboardItem {
	db.mu.RLock()
	defer db.mu.RUnlock()

	// Separate pinned and unpinned items
	pinned := make([]ClipboardItem, 0)
	unpinned := make([]ClipboardItem, 0)

	for _, item := range db.Items {
		if item.Pinned {
			pinned = append(pinned, item)
		} else {
			unpinned = append(unpinned, item)
		}
	}

	// Return pinned first, then unpinned
	result := make([]ClipboardItem, 0, len(db.Items))
	result = append(result, pinned...)
	result = append(result, unpinned...)
	return result
}

// ClearAll removes all items from the database
func (db *Database) ClearAll() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.Items = make([]ClipboardItem, 0)
	return db.saveInternal()
}

// GetItemCount returns the number of items in the database
func (db *Database) GetItemCount() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.Items)
}

// GetPinnedCount returns the number of pinned items
func (db *Database) GetPinnedCount() int {
	db.mu.RLock()
	defer db.mu.RUnlock()

	count := 0
	for _, item := range db.Items {
		if item.Pinned {
			count++
		}
	}
	return count
}
