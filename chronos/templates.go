package chronos

import (
	"database/sql"
	"fmt"
	// "github.com/regiellis/chronos-go/db" // Removed db import
)

const createTemplatesTableSQL = `
CREATE TABLE IF NOT EXISTS templates (
    name TEXT PRIMARY KEY,
    entry TEXT NOT NULL
);`

// EnsureTemplatesTable creates the templates table if it doesn't already exist.
// NOTE: The 'store' parameter is now a TemplateCreator interface.
// Internal DB calls will be compile errors until db.Store implements TemplateCreator.
func EnsureTemplatesTable(store TemplateCreator) error {
	// The following line will cause a compile error.
	_, err := store.DB.Exec(createTemplatesTableSQL)
	if err != nil {
		return fmt.Errorf("failed to ensure templates table: %w", err)
	}
	return nil
}

// SaveTemplate saves or updates an entry template.
// NOTE: The 'store' parameter is now a TemplateCreator interface.
func SaveTemplate(store TemplateCreator, name string, entryText string) error {
	// Ensure table exists first. This call now uses the interface method.
	// If EnsureTemplatesTable itself is part of the interface, this is fine.
	// If not, this specific call needs to be re-evaluated or moved to the concrete type.
	// For now, assuming EnsureTemplatesTable is also part of TemplateCreator for consistency.
	if err := store.EnsureTemplatesTable(); err != nil {
		// This will be a compile error if EnsureTemplatesTable is not on TemplateCreator.
		// It is on TemplateCreator as per interfaces.go.
		return err
	}

	query := `INSERT OR REPLACE INTO templates (name, entry) VALUES (?, ?)`
	// The following line will cause a compile error.
	_, err := store.DB.Exec(query, name, entryText)
	if err != nil {
		return fmt.Errorf("failed to save template '%s': %w", name, err)
	}
	return nil
}

// GetTemplate retrieves an entry template by its name.
// NOTE: The 'store' parameter is now a TemplateRetriever interface.
func GetTemplate(store TemplateRetriever, name string) (string, error) {
	// Ensuring table exists before get is tricky if GetTemplate only has TemplateRetriever.
	// The EnsureTemplatesTable method is on TemplateCreator.
	// This implies that for GetTemplate to ensure the table, the passed store would need to
	// also satisfy TemplateCreator, or table creation is guaranteed elsewhere.
	// For this refactor, we assume that by the time GetTemplate is called,
	// the table should exist (e.g. ensured by InitSchema or a SaveTemplate call).
	// If strict separation is needed, EnsureTemplatesTable might need to be called
	// by the concrete db.Store's GetTemplate method before its own query.

	var entryText string
	query := `SELECT entry FROM templates WHERE name = ?`
	// The following line will cause a compile error.
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
// func ListTemplates(store TemplateRetriever) (map[string]string, error) { ... }

// DeleteTemplate (Optional - if needed in future)
// func DeleteTemplate(store TemplateCreator, name string) error { ... }
