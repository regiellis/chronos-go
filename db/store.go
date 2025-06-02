package db

import (
	"database/sql"
	"strings"
	"time"

	log "github.com/charmbracelet/log"
	_ "github.com/mattn/go-sqlite3"
	"github.com/regiellis/chronos-go/chronos"
	"github.com/regiellis/chronos-go/utils"
)

// Store wraps the SQLite DB connection.
type Store struct {
	DB *sql.DB
}

// NewStore opens (or creates) the SQLite database.
func NewStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Error("[DB] Failed to open database: %v", err)
		return nil, err
	}
	// Only log errors, not info or routine feedback
	return &Store{DB: db}, nil
}

// InitSchema creates the entries and blocks tables if they don't exist.
func (s *Store) InitSchema() error {
	_, err := s.DB.Exec(`CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		block_id INTEGER,
		project TEXT,
		client TEXT,
		task TEXT,
		description TEXT,
		duration INTEGER,
		entry_time DATETIME,
		created_at DATETIME,
		billable BOOLEAN DEFAULT 1,
		rate REAL DEFAULT 0,
		invoiced BOOLEAN DEFAULT 0
	);`)
	if err != nil {
		log.Error("[DB] Failed to create entries table: %v", err)
		return err
	}
	_, err = s.DB.Exec(`CREATE TABLE IF NOT EXISTS blocks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		client TEXT,
		project TEXT,
		start_time DATETIME,
		end_time DATETIME,
		active BOOLEAN,
		created_at DATETIME
	);`)
	if err != nil {
		log.Error("[DB] Failed to create blocks table: %v", err)
		return err
	}
	return nil
}

// AddEntry inserts a new entry into the database.
func (s *Store) AddEntry(e *chronos.Entry) error {
	e.Project = utils.SanitizeString(e.Project)
	e.Client = utils.SanitizeString(e.Client)
	e.Task = utils.SanitizeString(e.Task)
	e.Description = utils.SanitizeDescription(e.Description)
	_, err := s.DB.Exec(`INSERT INTO entries (block_id, project, client, task, description, duration, entry_time, created_at, billable, rate, invoiced) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.BlockID, e.Project, e.Client, e.Task, e.Description, e.Duration, e.EntryTime, e.CreatedAt, e.Billable, e.Rate, e.Invoiced)
	if err != nil {
		log.Error("[DB] Failed to add entry: %v", err)
		return err
	}
	return nil
}

// AddBlock inserts a new block into the database.
func (s *Store) AddBlock(b *chronos.Block) error {
	b.Name = utils.SanitizeString(b.Name)
	b.Client = utils.SanitizeString(b.Client)
	b.Project = utils.SanitizeString(b.Project)
	_, err := s.DB.Exec(`INSERT INTO blocks (name, client, project, start_time, end_time, active, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		b.Name, b.Client, b.Project, b.StartTime, b.EndTime, b.Active, b.CreatedAt)
	if err != nil {
		log.Error("[DB] Failed to add block: %v", err)
		return err
	}
	return nil
}

// SetActiveBlock sets the specified block as active and deactivates others.
func (s *Store) SetActiveBlock(blockID int64) error {
	_, err := s.DB.Exec(`UPDATE blocks SET active = (id = ?)`, blockID)
	return err
}

// GetActiveBlock returns the currently active block, if any.
func (s *Store) GetActiveBlock() (*chronos.Block, error) {
	row := s.DB.QueryRow(`SELECT id, name, client, project, start_time, end_time, active, created_at FROM blocks WHERE active = 1 LIMIT 1`)
	var b chronos.Block
	err := row.Scan(&b.ID, &b.Name, &b.Client, &b.Project, &b.StartTime, &b.EndTime, &b.Active, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// ListEntries returns all entries, optionally filtered by block, project, client, task, date range, billable, rate, and invoiced.
func (s *Store) ListEntries(filter map[string]interface{}) ([]*chronos.Entry, error) {
	query := `SELECT id, block_id, project, client, task, description, duration, entry_time, created_at, billable, rate, invoiced FROM entries`
	args := []interface{}{}
	clauses := []string{}
	if filter != nil {
		if v, ok := filter["block_id"]; ok {
			clauses = append(clauses, "block_id = ?")
			args = append(args, v)
		}
		if v, ok := filter["project"]; ok {
			clauses = append(clauses, "project = ?")
			args = append(args, v)
		}
		if v, ok := filter["client"]; ok {
			clauses = append(clauses, "client = ?")
			args = append(args, v)
		}
		if v, ok := filter["task"]; ok {
			clauses = append(clauses, "task = ?")
			args = append(args, v)
		}
		if v, ok := filter["from"]; ok {
			clauses = append(clauses, "entry_time >= ?")
			args = append(args, v)
		}
		if v, ok := filter["to"]; ok {
			clauses = append(clauses, "entry_time <= ?")
			args = append(args, v)
		}
		if v, ok := filter["min_duration"]; ok {
			clauses = append(clauses, "duration >= ?")
			args = append(args, v)
		}
		if v, ok := filter["max_duration"]; ok {
			clauses = append(clauses, "duration <= ?")
			args = append(args, v)
		}
		if v, ok := filter["billable"]; ok {
			clauses = append(clauses, "billable = ?")
			args = append(args, v)
		}
		if v, ok := filter["min_rate"]; ok {
			clauses = append(clauses, "rate >= ?")
			args = append(args, v)
		}
		if v, ok := filter["max_rate"]; ok {
			clauses = append(clauses, "rate <= ?")
			args = append(args, v)
		}
		if v, ok := filter["invoiced"]; ok {
			clauses = append(clauses, "invoiced = ?")
			args = append(args, v)
		}
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY entry_time DESC"
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []*chronos.Entry
	for rows.Next() {
		var e chronos.Entry
		var entryTime, createdAt string
		if err := rows.Scan(&e.ID, &e.BlockID, &e.Project, &e.Client, &e.Task, &e.Description, &e.Duration, &entryTime, &createdAt, &e.Billable, &e.Rate, &e.Invoiced); err != nil {
			return nil, err
		}
		e.EntryTime, _ = time.Parse(time.RFC3339, entryTime)
		e.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		entries = append(entries, &e)
	}
	return entries, nil
}

// ListBlocks returns all blocks, optionally filtered by client or project.
func (s *Store) ListBlocks(filter map[string]interface{}) ([]*chronos.Block, error) {
	query := `SELECT id, name, client, project, start_time, end_time, active, created_at FROM blocks`
	args := []interface{}{}
	clauses := []string{}
	if filter != nil {
		if v, ok := filter["client"]; ok {
			clauses = append(clauses, "client = ?")
			args = append(args, v)
		}
		if v, ok := filter["project"]; ok {
			clauses = append(clauses, "project = ?")
			args = append(args, v)
		}
	}
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY start_time DESC"
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var blocks []*chronos.Block
	for rows.Next() {
		var b chronos.Block
		var start, end, created string
		if err := rows.Scan(&b.ID, &b.Name, &b.Client, &b.Project, &start, &end, &b.Active, &b.CreatedAt); err != nil {
			return nil, err
		}
		b.StartTime, _ = time.Parse(time.RFC3339, start)
		b.EndTime, _ = time.Parse(time.RFC3339, end)
		b.CreatedAt, _ = time.Parse(time.RFC3339, created)
		blocks = append(blocks, &b)
	}
	return blocks, nil
}

// QueryHistory returns the last N user queries for quick re-use.
func (s *Store) QueryHistory(limit int) ([]string, error) {
	_, err := s.DB.Exec(`CREATE TABLE IF NOT EXISTS query_history (id INTEGER PRIMARY KEY AUTOINCREMENT, query TEXT, created_at DATETIME)`)
	if err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(`SELECT query FROM query_history ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var queries []string
	for rows.Next() {
		var q string
		if err := rows.Scan(&q); err != nil {
			return nil, err
		}
		queries = append(queries, q)
	}
	return queries, nil
}

// SaveQuery stores a user query in the history.
func (s *Store) SaveQuery(query string) error {
	_, err := s.DB.Exec(`INSERT INTO query_history (query, created_at) VALUES (?, ?)`, query, time.Now())
	return err
}

// UpdateEntry updates an existing entry in the database.
func (s *Store) UpdateEntry(e *chronos.Entry) error {
	e.Project = utils.SanitizeString(e.Project)
	e.Client = utils.SanitizeString(e.Client)
	e.Task = utils.SanitizeString(e.Task)
	e.Description = utils.SanitizeDescription(e.Description)
	_, err := s.DB.Exec(`UPDATE entries SET block_id=?, project=?, client=?, task=?, description=?, duration=?, entry_time=?, created_at=?, billable=?, rate=?, invoiced=? WHERE id=?`,
		e.BlockID, e.Project, e.Client, e.Task, e.Description, e.Duration, e.EntryTime, e.CreatedAt, e.Billable, e.Rate, e.Invoiced, e.ID)
	if err != nil {
		log.Error("[DB] Failed to update entry: %v", err)
		return err
	}
	return nil
}

// DeleteEntry deletes an entry by ID.
func (s *Store) DeleteEntry(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM entries WHERE id=?`, id)
	if err != nil {
		log.Error("[DB] Failed to delete entry %d: %v", id, err)
		return err
	}
	return nil
}

// MarkEntriesInvoiced marks entries as invoiced by IDs.
func (s *Store) MarkEntriesInvoiced(ids []int64) error {
	for _, id := range ids {
		_, err := s.DB.Exec(`UPDATE entries SET invoiced = 1 WHERE id=?`, id)
		if err != nil {
			log.Error("[DB] Failed to mark entry %d as invoiced: %v", id, err)
			return err
		}
	}
	return nil
}

// FindUnbilledEntries returns entries not marked as invoiced.
func (s *Store) FindUnbilledEntries() ([]*chronos.Entry, error) {
	return s.ListEntries(map[string]interface{}{"billable": true, "invoiced": false})
}
