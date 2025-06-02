package chronos

import "time"

// Entry represents a time entry in the system.
type Entry struct {
	ID          int64     `json:"id"`
	BlockID     int64     `json:"block_id,omitempty"`
	ProjectID   int64     `json:"project_id,omitempty"`
	Summary     string    `json:"summary"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Invoiced    bool      `json:"invoiced"`
}
