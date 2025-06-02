package chronos

import (
	"sort"
	"time"
)

// CalculateProjectTotalsByProjectID aggregates total duration for each ProjectID.
// Duration is calculated in minutes.
func CalculateProjectTotalsByProjectID(entries []*Entry) map[int64]float64 {
	totals := make(map[int64]float64)
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		duration := entry.EndTime.Sub(entry.StartTime).Minutes()
		totals[entry.ProjectID] += duration
	}
	return totals
}

// CalculateReviewPeriodTotals calculates total duration of entries within a given period.
// Entries are considered within the period if their StartTime is on or after periodStart.
// Duration is calculated in minutes.
func CalculateReviewPeriodTotals(entries []*Entry, periodStart time.Time) float64 {
	var totalDuration float64
	for _, entry := range entries {
		if entry == nil {
			continue
		}
		// Ensure StartTime is valid and after or equal to periodStart
		if !entry.StartTime.IsZero() && (entry.StartTime.Equal(periodStart) || entry.StartTime.After(periodStart)) {
			totalDuration += entry.EndTime.Sub(entry.StartTime).Minutes()
		}
	}
	return totalDuration
}

// SortEntriesByStartTime sorts a slice of Entry pointers by their StartTime.
// Pass asc=true for ascending, asc=false for descending.
// Nil entries or entries with zero StartTime are pushed towards the beginning for ascending sort,
// and towards the end for descending sort, to handle potentially incomplete data.
func SortEntriesByStartTime(entries []*Entry, asc bool) {
	sort.SliceStable(entries, func(i, j int) bool {
		entryI := entries[i]
		entryJ := entries[j]

		// Handle nil entries or zero StartTime
		iValid := entryI != nil && !entryI.StartTime.IsZero()
		jValid := entryJ != nil && !entryJ.StartTime.IsZero()

		if !iValid && !jValid { return false } // Both invalid, keep order
		if !iValid { return asc } // Only i is invalid, if asc, i is "less" (comes first)
		if !jValid { return !asc } // Only j is invalid, if asc, j is "greater" (i comes first)

		// Both are valid
		if asc {
			return entryI.StartTime.Before(entryJ.StartTime)
		}
		return entryI.StartTime.After(entryJ.StartTime)
	})
}

// IdleGap represents a detected period of inactivity between two entries.
type IdleGap struct {
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// DetectIdleGaps identifies periods of inactivity between entries longer than minGapDuration.
// It sorts the entries by StartTime ASC before processing.
func DetectIdleGaps(entries []*Entry, minGapDuration time.Duration) []IdleGap {
	if len(entries) < 2 {
		return nil
	}

	// Create a defensive copy to sort, ensuring the original slice order is untouched.
	sortedEntries := make([]*Entry, len(entries))
	copy(sortedEntries, entries)
	SortEntriesByStartTime(sortedEntries, true) // true for ascending

	var gaps []IdleGap
	for i := 0; i < len(sortedEntries)-1; i++ {
		prevEntry := sortedEntries[i]
		nextEntry := sortedEntries[i+1]

		// Ensure entries and their relevant time fields are valid
		if prevEntry == nil || nextEntry == nil || prevEntry.EndTime.IsZero() || nextEntry.StartTime.IsZero() {
			continue 
		}
		
		// A gap occurs only if the next entry starts after the previous one ended.
		if nextEntry.StartTime.After(prevEntry.EndTime) {
			gapDuration := nextEntry.StartTime.Sub(prevEntry.EndTime)
			if gapDuration >= minGapDuration {
				gaps = append(gaps, IdleGap{
					StartTime: prevEntry.EndTime,
					EndTime:   nextEntry.StartTime,
					Duration:  gapDuration,
				})
			}
		}
	}
	return gaps
}

// Future enhancements could include:
// - CalculateClientTotals: Would require entries to have ClientID or means to derive it.
//   (e.g., func CalculateClientTotals(projects []*Project, entries []*Entry) map[int64]float64)
// - CalculateTaskTotals: Would require parsing Summary field or a dedicated Task field in Entry.
// - Functions to return lists of top N projects/clients/tasks.
// - More sophisticated filtering options within calculations.
