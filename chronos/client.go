package chronos

import (
	"time"

	"github.com/regiellis/chronos-go/db"
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
func CreateClient(store *db.Store, client *Client) error {
	client.CreatedAt = time.Now()
	client.UpdatedAt = time.Now()

	query := `
		INSERT INTO clients (name, contact_info, created_at, updated_at)
		VALUES (?, ?, ?, ?)`
	res, err := store.DB.Exec(query, client.Name, client.ContactInfo, client.CreatedAt, client.UpdatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	client.ID = id
	return nil
}

// GetClientByID retrieves a client from the database by its ID.
func GetClientByID(store *db.Store, id int64) (*Client, error) {
	client := &Client{}
	query := `
		SELECT id, name, contact_info, created_at, updated_at
		FROM clients WHERE id = ?`
	err := store.DB.QueryRow(query, id).Scan(&client.ID, &client.Name, &client.ContactInfo, &client.CreatedAt, &client.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// UpdateClient updates an existing client in the database.
func UpdateClient(store *db.Store, client *Client) error {
	client.UpdatedAt = time.Now()
	query := `
		UPDATE clients
		SET name = ?, contact_info = ?, updated_at = ?
		WHERE id = ?`
	_, err := store.DB.Exec(query, client.Name, client.ContactInfo, client.UpdatedAt, client.ID)
	return err
}

// DeleteClient removes a client from the database by its ID.
func DeleteClient(store *db.Store, id int64) error {
	query := "DELETE FROM clients WHERE id = ?"
	_, err := store.DB.Exec(query, id)
	return err
}

// ListClients retrieves a list of all clients from the database.
func ListClients(store *db.Store) ([]*Client, error) {
	query := `
		SELECT id, name, contact_info, created_at, updated_at
		FROM clients`
	rows, err := store.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clients := []*Client{}
	for rows.Next() {
		client := &Client{}
		err := rows.Scan(&client.ID, &client.Name, &client.ContactInfo, &client.CreatedAt, &client.UpdatedAt)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	return clients, nil
}
