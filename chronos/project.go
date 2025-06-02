package chronos

import (
	"time"

	"github.com/regiellis/chronos-go/db"
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
func CreateProject(store *db.Store, project *Project) error {
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()

	query := `
		INSERT INTO projects (name, client_id, rate, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)`
	res, err := store.DB.Exec(query, project.Name, project.ClientID, project.Rate, project.CreatedAt, project.UpdatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	project.ID = id
	return nil
}

// GetProjectByID retrieves a project from the database by its ID.
func GetProjectByID(store *db.Store, id int64) (*Project, error) {
	project := &Project{}
	query := `
		SELECT id, name, client_id, rate, created_at, updated_at
		FROM projects WHERE id = ?`
	err := store.DB.QueryRow(query, id).Scan(&project.ID, &project.Name, &project.ClientID, &project.Rate, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return project, nil
}

// UpdateProject updates an existing project in the database.
func UpdateProject(store *db.Store, project *Project) error {
	project.UpdatedAt = time.Now()
	query := `
		UPDATE projects
		SET name = ?, client_id = ?, rate = ?, updated_at = ?
		WHERE id = ?`
	_, err := store.DB.Exec(query, project.Name, project.ClientID, project.Rate, project.UpdatedAt, project.ID)
	return err
}

// DeleteProject removes a project from the database by its ID.
func DeleteProject(store *db.Store, id int64) error {
	query := "DELETE FROM projects WHERE id = ?"
	_, err := store.DB.Exec(query, id)
	return err
}

// ListProjects retrieves a list of projects from the database, optionally filtered by clientID.
func ListProjects(store *db.Store, clientID *int64) ([]*Project, error) {
	var rows *db.Rows
	var err error

	if clientID != nil {
		query := `
			SELECT id, name, client_id, rate, created_at, updated_at
			FROM projects WHERE client_id = ?`
		rows, err = store.DB.Query(query, *clientID)
	} else {
		query := `
			SELECT id, name, client_id, rate, created_at, updated_at
			FROM projects`
		rows, err = store.DB.Query(query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := []*Project{}
	for rows.Next() {
		project := &Project{}
		err := rows.Scan(&project.ID, &project.Name, &project.ClientID, &project.Rate, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, nil
}
