package chronos

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/regiellis/chronos-go/db" // Assumed import path for db.Store
)

type Block struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Client    string    `json:"client"`    // Consider if this should be ClientID int64
	Project   string    `json:"project"`   // Consider if this should be ProjectID int64
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"` // Nullable / Zero time if not ended
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"` // Added UpdatedAt for consistency
}

// CreateBlock adds a new block to the database.
func CreateBlock(store *db.Store, block *Block) error {
	block.CreatedAt = time.Now()
	block.UpdatedAt = time.Now()

	// Ensure EndTime is NULL if it's zero, otherwise use the provided time.
	var endTime interface{}
	if block.EndTime.IsZero() {
		endTime = nil
	} else {
		endTime = block.EndTime
	}

	query := `
		INSERT INTO blocks (name, client, project, start_time, end_time, active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := store.DB.Exec(query, block.Name, block.Client, block.Project, block.StartTime, endTime, block.Active, block.CreatedAt, block.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreateBlock: failed to execute insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("CreateBlock: failed to get last insert ID: %w", err)
	}
	block.ID = id
	return nil
}

// GetBlockByID retrieves a block from the database by its ID.
func GetBlockByID(store *db.Store, id int64) (*Block, error) {
	block := &Block{}
	// Handle potential NULL EndTime from DB
	var endTime sql.NullTime

	query := `
		SELECT id, name, client, project, start_time, end_time, active, created_at, updated_at
		FROM blocks WHERE id = ?`
	err := store.DB.QueryRow(query, id).Scan(
		&block.ID, &block.Name, &block.Client, &block.Project,
		&block.StartTime, &endTime, &block.Active, &block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("GetBlockByID: no block found with ID %d: %w", id, err)
		}
		return nil, fmt.Errorf("GetBlockByID: failed to scan row: %w", err)
	}
	if endTime.Valid {
		block.EndTime = endTime.Time
	}
	return block, nil
}

// UpdateBlock updates an existing block in the database.
func UpdateBlock(store *db.Store, block *Block) error {
	block.UpdatedAt = time.Now()

	var endTime interface{}
	if block.EndTime.IsZero() {
		endTime = nil
	} else {
		endTime = block.EndTime
	}

	query := `
		UPDATE blocks
		SET name = ?, client = ?, project = ?, start_time = ?, end_time = ?, active = ?, updated_at = ?
		WHERE id = ?`
	_, err := store.DB.Exec(query, block.Name, block.Client, block.Project, block.StartTime, endTime, block.Active, block.UpdatedAt, block.ID)
	if err != nil {
		return fmt.Errorf("UpdateBlock: failed to execute update: %w", err)
	}
	return nil
}

// DeleteBlock removes a block from the database by its ID.
func DeleteBlock(store *db.Store, id int64) error {
	query := "DELETE FROM blocks WHERE id = ?"
	_, err := store.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("DeleteBlock: failed to execute delete: %w", err)
	}
	return nil
}

// ListBlocks retrieves a list of blocks from the database, optionally filtered.
// Example filters: "active" (bool), "client" (string), "project" (string), "start_date", "end_date"
func ListBlocks(store *db.Store, filters map[string]interface{}) ([]*Block, error) {
	baseQuery := "SELECT id, name, client, project, start_time, end_time, active, created_at, updated_at FROM blocks"
	var conditions []string
	var args []interface{}

	for key, value := range filters {
		switch key {
		case "active":
			conditions = append(conditions, "active = ?")
			args = append(args, value)
		case "client":
			conditions = append(conditions, "client = ?") // Or client_id if schema changes
			args = append(args, value)
		case "project":
			conditions = append(conditions, "project = ?") // Or project_id if schema changes
			args = append(args, value)
		case "start_date": // Blocks started on or after this date
			conditions = append(conditions, "date(start_time) >= date(?)")
			args = append(args, value)
		case "end_date": // Blocks started on or before this date (or active blocks if end_time is also checked)
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
		return nil, fmt.Errorf("ListBlocks: failed to execute query: %w", err)
	}
	defer rows.Close()

	blocks := []*Block{}
	for rows.Next() {
		block := &Block{}
		var endTime sql.NullTime
		err := rows.Scan(
			&block.ID, &block.Name, &block.Client, &block.Project,
			&block.StartTime, &endTime, &block.Active, &block.CreatedAt, &block.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("ListBlocks: failed to scan row: %w", err)
		}
		if endTime.Valid {
			block.EndTime = endTime.Time
		}
		blocks = append(blocks, block)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ListBlocks: error during rows iteration: %w", err)
	}
	return blocks, nil
}

// GetActiveBlock retrieves the currently active block, if any.
// Assumes only one block can be active at a time.
func GetActiveBlock(store *db.Store) (*Block, error) {
	block := &Block{}
	var endTime sql.NullTime
	query := `
		SELECT id, name, client, project, start_time, end_time, active, created_at, updated_at
		FROM blocks WHERE active = TRUE LIMIT 1` // Or active = 1 for SQLite
	err := store.DB.QueryRow(query).Scan(
		&block.ID, &block.Name, &block.Client, &block.Project,
		&block.StartTime, &endTime, &block.Active, &block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active block is not an error in this context
		}
		return nil, fmt.Errorf("GetActiveBlock: failed to scan row: %w", err)
	}
	if endTime.Valid {
		block.EndTime = endTime.Time
	}
	return block, nil
}

// SetActiveBlock sets a block as active and deactivates others.
// This is a common operation, so good to have a dedicated function.
func SetActiveBlock(store *db.Store, id int64) error {
	tx, err := store.DB.Begin()
	if err != nil {
		return fmt.Errorf("SetActiveBlock: failed to begin transaction: %w", err)
	}

	// Deactivate all other blocks
	_, err = tx.Exec("UPDATE blocks SET active = FALSE, updated_at = ? WHERE active = TRUE", time.Now())
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("SetActiveBlock: failed to deactivate other blocks: %w", err)
	}

	// Activate the specified block
	_, err = tx.Exec("UPDATE blocks SET active = TRUE, updated_at = ? WHERE id = ?", time.Now(), id)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("SetActiveBlock: failed to activate block ID %d: %w", id, err)
	}

	return tx.Commit()
}
