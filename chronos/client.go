package chronos

import (
	"database/sql" // Added for sql.ErrNoRows and sql.Rows
	"fmt"          // Added for fmt.Errorf
	"time"
	// "github.com/regiellis/chronos-go/db" // Removed db import
)

// Client represents a client in the system.
type Client struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	ContactInfo string    `json:"contact_info"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateClient adds a new client to the database.
// NOTE: The 'store' parameter is now a ClientStore interface.
// Internal DB calls will be compile errors until db.Store implements ClientStore.
func CreateClient(store ClientStore, client *Client) error {
	client.CreatedAt = time.Now()
	client.UpdatedAt = time.Now()

	query := `
		INSERT INTO clients (name, contact_info, created_at, updated_at)
		VALUES (?, ?, ?, ?)`
	// The following line will cause a compile error.
	res, err := store.DB.Exec(query, client.Name, client.ContactInfo, client.CreatedAt, client.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreateClient: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("CreateClient LastInsertId: %w", err)
	}
	client.ID = id
	return nil
}

// GetClientByID retrieves a client from the database by its ID.
// NOTE: The 'store' parameter is now a ClientStore interface.
func GetClientByID(store ClientStore, id int64) (*Client, error) {
	client := &Client{}
	query := `
		SELECT id, name, contact_info, created_at, updated_at
		FROM clients WHERE id = ?`
	// The following line will cause a compile error.
	err := store.DB.QueryRow(query, id).Scan(&client.ID, &client.Name, &client.ContactInfo, &client.CreatedAt, &client.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("GetClientByID: client with ID %d not found: %w", id, err)
		}
		return nil, fmt.Errorf("GetClientByID: %w", err)
	}
	return client, nil
}

// UpdateClient updates an existing client in the database.
// NOTE: The 'store' parameter is now a ClientStore interface.
func UpdateClient(store ClientStore, client *Client) error {
	client.UpdatedAt = time.Now()
	query := `
		UPDATE clients
		SET name = ?, contact_info = ?, updated_at = ?
		WHERE id = ?`
	// The following line will cause a compile error.
	_, err := store.DB.Exec(query, client.Name, client.ContactInfo, client.UpdatedAt, client.ID)
	if err != nil {
		return fmt.Errorf("UpdateClient: %w", err)
	}
	return err
}

// DeleteClient removes a client from the database by its ID.
// NOTE: The 'store' parameter is now a ClientStore interface.
func DeleteClient(store ClientStore, id int64) error {
	query := "DELETE FROM clients WHERE id = ?"
	// The following line will cause a compile error.
	_, err := store.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("DeleteClient: %w", err)
	}
	return err
}

// ListClients retrieves a list of all clients from the database.
// NOTE: The 'store' parameter is now a ClientStore interface.
func ListClients(store ClientStore) ([]*Client, error) {
	query := `
		SELECT id, name, contact_info, created_at, updated_at
		FROM clients ORDER BY name ASC` // Added ORDER BY
	// The following line will cause a compile error.
	rows, err := store.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ListClients query: %w", err)
	}
	defer rows.Close()

	clients := []*Client{}
	for rows.Next() {
		client := &Client{}
		err := rows.Scan(&client.ID, &client.Name, &client.ContactInfo, &client.CreatedAt, &client.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("ListClients scan: %w", err)
		}
		clients = append(clients, client)
	}
	if err = rows.Err(); err != nil { // Check for errors during iteration
		return nil, fmt.Errorf("ListClients iteration: %w", err)
	}
	return clients, nil
}
