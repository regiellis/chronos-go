package chronos

import (
	"database/sql"
	"fmt"
	"github.com/regiellis/chronos-go/db" // Assumed import path
)

const createTemplatesTableSQL = `
CREATE TABLE IF NOT EXISTS templates (
    name TEXT PRIMARY KEY,
    entry TEXT NOT NULL
);`

// EnsureTemplatesTable creates the templates table if it doesn't already exist.
func EnsureTemplatesTable(store *db.Store) error {
	_, err := store.DB.Exec(createTemplatesTableSQL)
	if err != nil {
		return fmt.Errorf("failed to ensure templates table: %w", err)
	}
	return nil
}

// SaveTemplate saves or updates an entry template.
func SaveTemplate(store *db.Store, name string, entryText string) error {
	// Ensure table exists first
	if err := EnsureTemplatesTable(store); err != nil {
		return err
	}

	query := `INSERT OR REPLACE INTO templates (name, entry) VALUES (?, ?)`
	_, err := store.DB.Exec(query, name, entryText)
	if err != nil {
		return fmt.Errorf("failed to save template '%s': %w", name, err)
	}
	return nil
}

// GetTemplate retrieves an entry template by its name.
func GetTemplate(store *db.Store, name string) (string, error) {
	// Ensure table exists first (though in read-only scenario, might not be strictly necessary if creation is guaranteed elsewhere)
	// However, for robustness, especially if GetTemplate could be called before any SaveTemplate.
	if err := EnsureTemplatesTable(store); err != nil { 
		return "", err
	}

	var entryText string
	query := `SELECT entry FROM templates WHERE name = ?`
	err := store.DB.QueryRow(query, name).Scan(&entryText)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("template '%s' not found: %w", name, err)
		}
		return "", fmt.Errorf("failed to get template '%s': %w", name, err)
	}
	return entryText, nil
}

// ListTemplates (Optional - if needed in future)
// func ListTemplates(store *db.Store) (map[string]string, error) { ... }

// DeleteTemplate (Optional - if needed in future)
// func DeleteTemplate(store *db.Store, name string) error { ... }
