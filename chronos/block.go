package chronos

import "time"

type Block struct {
	ID        int64
	Name      string
	Client    string
	Project   string
	StartTime time.Time
	EndTime   time.Time
	Active    bool
	CreatedAt time.Time
}
