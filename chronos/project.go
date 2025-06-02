package chronos

import (
	"database/sql" // Added for sql.ErrNoRows, and db.Rows if used directly (though it's store.DB.Rows)
	"fmt"          // Added for fmt.Errorf
	"time"
	// "github.com/regiellis/chronos-go/db" // Removed db import
)

// Project represents a project in the system.
type Project struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	ClientID  int64     `json:"client_id"`
	Rate      float64   `json:"rate"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateProject adds a new project to the database.
// NOTE: The 'store' parameter is now a ProjectStore interface.
// Internal DB calls will be compile errors until db.Store implements ProjectStore.
func CreateProject(store ProjectStore, project *Project) error {
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	query := `
		INSERT INTO projects (name, client_id, rate, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`
	// The following line will cause a compile error.
	res, err := store.DB.Exec(query, project.Name, project.ClientID, project.Rate, project.CreatedAt, project.UpdatedAt)
	if err != nil {
		return fmt.Errorf("CreateProject: %w", err) // Wrap error
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("CreateProject LastInsertId: %w", err) // Wrap error
	}
	project.ID = id
	return nil
}

// GetProjectByID retrieves a project from the database by its ID.
// NOTE: The 'store' parameter is now a ProjectStore interface.
func GetProjectByID(store ProjectStore, id int64) (*Project, error) {
	project := &Project{}
	query := `
		SELECT id, name, client_id, rate, created_at, updated_at
		FROM projects WHERE id = ?`
	// The following line will cause a compile error.
	err := store.DB.QueryRow(query, id).Scan(&project.ID, &project.Name, &project.ClientID, &project.Rate, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows { // Added sql.ErrNoRows check
			return nil, fmt.Errorf("GetProjectByID: project with ID %d not found: %w", id, err)
		}
		return nil, fmt.Errorf("GetProjectByID: %w", err) // Wrap error
	}
	return project, nil
}

// UpdateProject updates an existing project in the database.
// NOTE: The 'store' parameter is now a ProjectStore interface.
func UpdateProject(store ProjectStore, project *Project) error {
	project.UpdatedAt = time.Now()
	query := `
		UPDATE projects
		SET name = ?, client_id = ?, rate = ?, updated_at = ?
		WHERE id = ?`
	// The following line will cause a compile error.
	_, err := store.DB.Exec(query, project.Name, project.ClientID, project.Rate, project.UpdatedAt, project.ID)
	if err != nil {
		return fmt.Errorf("UpdateProject: %w", err) // Wrap error
	}
	return err
}

// DeleteProject removes a project from the database by its ID.
// NOTE: The 'store' parameter is now a ProjectStore interface.
func DeleteProject(store ProjectStore, id int64) error {
	query := "DELETE FROM projects WHERE id = ?"
	// The following line will cause a compile error.
	_, err := store.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("DeleteProject: %w", err) // Wrap error
	}
	return err
}

// ListProjects retrieves a list of projects from the database, optionally filtered by clientID.
// NOTE: The 'store' parameter is now a ProjectStore interface.
func ListProjects(store ProjectStore, clientID *int64) ([]*Project, error) {
	var rows *sql.Rows // Changed from db.Rows to sql.Rows
	var err error

	var query string
	var queryArgs []interface{}

	if clientID != nil {
		query = `
			SELECT id, name, client_id, rate, created_at, updated_at
			FROM projects WHERE client_id = ? ORDER BY name ASC`
		queryArgs = append(queryArgs, *clientID)
	} else {
		query = `
			SELECT id, name, client_id, rate, created_at, updated_at
			FROM projects ORDER BY name ASC`
	}

	// The following line will cause a compile error.
	rows, err = store.DB.Query(query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("ListProjects query: %w", err) // Wrap error
	}
	defer rows.Close()

	projects := []*Project{}
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(&project.ID, &project.Name, &project.ClientID, &project.Rate, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("ListProjects scan: %w", err) // Wrap error
		}
		projects = append(projects, project)
	}
	if err = rows.Err(); err != nil { // Check for errors during iteration
		return nil, fmt.Errorf("ListProjects iteration: %w", err)
	}
	return projects, nil
}
