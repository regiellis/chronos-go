package chronos

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	// "github.com/regiellis/chronos-go/db" // Removed db import
)

// CreateEntry adds a new entry to the database.
// Assumes entry.CreatedAt and entry.UpdatedAt will be set by the caller or here.
// NOTE: The 'store' parameter is now an EntryStore interface.
// The internal call store.DB.Exec will be a compile error until db.Store implements EntryStore.
func CreateEntry(store EntryStore, entry *Entry) error {
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	entry.UpdatedAt = time.Now()

	query := `
		INSERT INTO entries (block_id, project_id, summary, start_time, end_time, created_at, updated_at, invoiced)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	// The following line will cause a compile error until EntryStore is implemented by a type
	// that has a DB field or the CreateEntry method itself is moved to the implementing type.
	res, err := store.DB.Exec(query, entry.BlockID, entry.ProjectID, entry.Summary, entry.StartTime, entry.EndTime, entry.CreatedAt, entry.UpdatedAt, entry.Invoiced)
	if err != nil {
		return fmt.Errorf("CreateEntry: failed to execute insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("CreateEntry: failed to get last insert ID: %w", err)
	}
	entry.ID = id
	return nil
}

// GetEntryByID retrieves an entry from the database by its ID.
// NOTE: The 'store' parameter is now an EntryStore interface.
// The internal call store.DB.QueryRow will be a compile error.
func GetEntryByID(store EntryStore, id int64) (*Entry, error) {
	entry := &Entry{}
	query := `
		SELECT id, block_id, project_id, summary, start_time, end_time, created_at, updated_at, invoiced
		FROM entries WHERE id = ?`
	// The following line will cause a compile error.
	err := store.DB.QueryRow(query, id).Scan(
		&entry.ID, &entry.BlockID, &entry.ProjectID, &entry.Summary,
		&entry.StartTime, &entry.EndTime, &entry.CreatedAt, &entry.UpdatedAt, &entry.Invoiced,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("GetEntryByID: no entry found with ID %d: %w", id, err)
		}
		return nil, fmt.Errorf("GetEntryByID: failed to scan row: %w", err)
	}
	return entry, nil
}

// UpdateEntry updates an existing entry in the database.
// NOTE: The 'store' parameter is now an EntryStore interface.
// The internal call store.DB.Exec will be a compile error.
func UpdateEntry(store EntryStore, entry *Entry) error {
	entry.UpdatedAt = time.Now()
	query := `
		UPDATE entries
		SET block_id = ?, project_id = ?, summary = ?, start_time = ?, end_time = ?, updated_at = ?, invoiced = ?
		WHERE id = ?`
	// The following line will cause a compile error.
	_, err := store.DB.Exec(query, entry.BlockID, entry.ProjectID, entry.Summary, entry.StartTime, entry.EndTime, entry.UpdatedAt, entry.Invoiced, entry.ID)
	if err != nil {
		return fmt.Errorf("UpdateEntry: failed to execute update: %w", err)
	}
	return nil
}

// DeleteEntry removes an entry from the database by its ID.
// NOTE: The 'store' parameter is now an EntryStore interface.
// The internal call store.DB.Exec will be a compile error.
func DeleteEntry(store EntryStore, id int64) error {
	query := "DELETE FROM entries WHERE id = ?"
	// The following line will cause a compile error.
	_, err := store.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("DeleteEntry: failed to execute delete: %w", err)
	}
	return nil
}

// ListEntries retrieves a list of entries from the database, optionally filtered.
// NOTE: The 'store' parameter is now an EntryStore interface.
// The internal call store.DB.Query will be a compile error.
func ListEntries(store EntryStore, filters map[string]interface{}) ([]*Entry, error) {
	baseQuery := "SELECT id, block_id, project_id, summary, start_time, end_time, created_at, updated_at, invoiced FROM entries"
	var conditions []string
	var args []interface{}

	for key, value := range filters {
		switch key {
		case "block_id":
			conditions = append(conditions, "block_id = ?")
			args = append(args, value)
		case "project_id":
			conditions = append(conditions, "project_id = ?")
			args = append(args, value)
		case "invoiced":
			conditions = append(conditions, "invoiced = ?")
			args = append(args, value)
		case "start_date":
			conditions = append(conditions, "date(start_time) >= date(?)")
			args = append(args, value)
		case "end_date":
			conditions = append(conditions, "date(start_time) <= date(?)")
			args = append(args, value)
		}
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY start_time DESC"

	// The following line will cause a compile error.
	rows, err := store.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("ListEntries: failed to execute query: %w", err)
	}
	defer rows.Close()

	entries := []*Entry{}
	for rows.Next() {
		entry := &Entry{}
		err := rows.Scan(
			&entry.ID, &entry.BlockID, &entry.ProjectID, &entry.Summary,
			&entry.StartTime, &entry.EndTime, &entry.CreatedAt, &entry.UpdatedAt, &entry.Invoiced,
		)
		if err != nil {
			return nil, fmt.Errorf("ListEntries: failed to scan row: %w", err)
		}
		entries = append(entries, entry)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ListEntries: error during rows iteration: %w", err)
	}
	return entries, nil
}
