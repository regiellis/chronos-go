package chronos

import "time"

type Entry struct {
	ID          int64
	BlockID     int64
	Project     string
	Client      string
	Task        string
	Description string
	Duration    int64
	EntryTime   time.Time
	CreatedAt   time.Time
	Billable    bool
	Rate        float64
	Invoiced    bool
}
