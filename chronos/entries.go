package chronos

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/regiellis/chronos-go/db" // Assumed import path for db.Store
	// Assuming chronos.Entry is defined in the same package or imported appropriately.
	// If chronos.Entry is in "github.com/regiellis/chronos-go/chronos", it would be just "Entry" here.
)

// CreateEntry adds a new entry to the database.
// Assumes entry.CreatedAt and entry.UpdatedAt will be set by the caller or here.
func CreateEntry(store *db.Store, entry *Entry) error {
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	entry.UpdatedAt = time.Now()

	query := `
		INSERT INTO entries (block_id, project_id, summary, start_time, end_time, created_at, updated_at, invoiced)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
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
func GetEntryByID(store *db.Store, id int64) (*Entry, error) {
	entry := &Entry{}
	query := `
		SELECT id, block_id, project_id, summary, start_time, end_time, created_at, updated_at, invoiced
		FROM entries WHERE id = ?`
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
func UpdateEntry(store *db.Store, entry *Entry) error {
	entry.UpdatedAt = time.Now()
	query := `
		UPDATE entries
		SET block_id = ?, project_id = ?, summary = ?, start_time = ?, end_time = ?, updated_at = ?, invoiced = ?
		WHERE id = ?`
	_, err := store.DB.Exec(query, entry.BlockID, entry.ProjectID, entry.Summary, entry.StartTime, entry.EndTime, entry.UpdatedAt, entry.Invoiced, entry.ID)
	if err != nil {
		return fmt.Errorf("UpdateEntry: failed to execute update: %w", err)
	}
	return nil
}

// DeleteEntry removes an entry from the database by its ID.
func DeleteEntry(store *db.Store, id int64) error {
	query := "DELETE FROM entries WHERE id = ?"
	_, err := store.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("DeleteEntry: failed to execute delete: %w", err)
	}
	return nil
}

// ListEntries retrieves a list of entries from the database, optionally filtered.
// Example filters: "block_id", "project_id", "invoiced", "start_date", "end_date"
func ListEntries(store *db.Store, filters map[string]interface{}) ([]*Entry, error) {
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
		case "start_date": // Assumes value is time.Time or string parsable to time
			conditions = append(conditions, "date(start_time) >= date(?)")
			args = append(args, value)
		case "end_date": // Assumes value is time.Time or string parsable to time
			conditions = append(conditions, "date(start_time) <= date(?)")
			args = append(args, value)
		// Add more filters as needed
		}
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY start_time DESC" // Default ordering

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
