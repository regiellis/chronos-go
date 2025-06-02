package chronos_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/regiellis/chronos-go/chronos"
)

func TestCalculateProjectTotalsByProjectID(t *testing.T) {
	now := time.Now()
	entries := []*chronos.Entry{
		{ProjectID: 1, StartTime: now, EndTime: now.Add(1 * time.Hour)},                      // 60 mins
		{ProjectID: 2, StartTime: now, EndTime: now.Add(30 * time.Minute)},                    // 30 mins
		{ProjectID: 1, StartTime: now.Add(2 * time.Hour), EndTime: now.Add(3 * time.Hour)},    // 60 mins
		{ProjectID: 3, StartTime: now, EndTime: now.Add(0 * time.Minute)},                     // 0 mins
		nil, // test nil entry
	}

	expectedTotals := map[int64]float64{
		1: 120.0, // 60 + 60
		2: 30.0,
		3: 0.0,
	}

	actualTotals := chronos.CalculateProjectTotalsByProjectID(entries)

	if !reflect.DeepEqual(actualTotals, expectedTotals) {
		t.Errorf("CalculateProjectTotalsByProjectID: expected %v, got %v", expectedTotals, actualTotals)
	}
}

func TestCalculateReviewPeriodTotals(t *testing.T) {
	now := time.Now()
	periodStart := now.Add(-24 * time.Hour) // Last 24 hours

	entries := []*chronos.Entry{
		// Inside period
		{StartTime: now.Add(-1 * time.Hour), EndTime: now},                                  // 60 mins
		{StartTime: periodStart, EndTime: periodStart.Add(30 * time.Minute)},               // 30 mins
		// Outside period (before)
		{StartTime: periodStart.Add(-2 * time.Hour), EndTime: periodStart.Add(-1 * time.Hour)}, // Should be ignored
		// Partially inside (if logic were more complex, but current logic is based on StartTime >= periodStart)
		{StartTime: now.Add(1 * time.Hour), EndTime: now.Add(2 * time.Hour)}, // 60 mins (future, but after periodStart)
		nil, // test nil entry
		{StartTime: time.Time{}}, // test zero time entry
	}

	// Expected total: 60 + 30 + 60 = 150
	expectedTotalMinutes := 150.0
	actualTotalMinutes := chronos.CalculateReviewPeriodTotals(entries, periodStart)

	if actualTotalMinutes != expectedTotalMinutes {
		t.Errorf("CalculateReviewPeriodTotals: expected %.2f minutes, got %.2f minutes", expectedTotalMinutes, actualTotalMinutes)
	}

	// Test with empty entries
	if total := chronos.CalculateReviewPeriodTotals([]*chronos.Entry{}, periodStart); total != 0 {
		t.Errorf("CalculateReviewPeriodTotals with empty slice: expected 0, got %.2f", total)
	}
}

func TestSortEntriesByStartTime(t *testing.T) {
	t1 := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	t3 := time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC)
	tZero := time.Time{}

	tests := []struct {
		name     string
		asc      bool
		input    []*chronos.Entry
		expected []*chronos.Entry
	}{
		{
			name: "Ascending sort",
			asc:  true,
			input: []*chronos.Entry{
				{ID: 1, Summary: "Entry 1", StartTime: t1},
				{ID: 2, Summary: "Entry 2", StartTime: t2},
				nil, // Nil entry
				{ID: 3, Summary: "Entry 3", StartTime: t3},
				{ID: 4, Summary: "Zero time", StartTime: tZero},
			},
			expected: []*chronos.Entry{
				nil, // Nil entries pushed first by current sort logic
				{ID: 4, Summary: "Zero time", StartTime: tZero}, // Zero times after nils
				{ID: 3, Summary: "Entry 3", StartTime: t3},
				{ID: 1, Summary: "Entry 1", StartTime: t1},
				{ID: 2, Summary: "Entry 2", StartTime: t2},
			},
		},
		{
			name: "Descending sort",
			asc:  false,
			input: []*chronos.Entry{
				{ID: 1, Summary: "Entry 1", StartTime: t1},
				{ID: 2, Summary: "Entry 2", StartTime: t2},
				{ID: 3, Summary: "Entry 3", StartTime: t3},
			},
			expected: []*chronos.Entry{
				{ID: 2, Summary: "Entry 2", StartTime: t2},
				{ID: 1, Summary: "Entry 1", StartTime: t1},
				{ID: 3, Summary: "Entry 3", StartTime: t3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chronos.SortEntriesByStartTime(tt.input, tt.asc)
			if !reflect.DeepEqual(tt.input, tt.expected) {
				t.Errorf("SortEntriesByStartTime (%s) incorrect.\nExpected order IDs:", tt.name)
				for _, e := range tt.expected { if e != nil { t.Logf("  ID: %d, Time: %v", e.ID, e.StartTime) } else { t.Logf("  ID: nil")} }
				t.Errorf("Got order IDs:")
				for _, e := range tt.input { if e != nil { t.Logf("  ID: %d, Time: %v", e.ID, e.StartTime) } else { t.Logf("  ID: nil")} }
			}
		})
	}
}


func TestDetectIdleGaps(t *testing.T) {
	y2023m1d1 := func(hour, min int) time.Time {
		return time.Date(2023, 1, 1, hour, min, 0, 0, time.UTC)
	}

	entries := []*chronos.Entry{
		{ID: 1, StartTime: y2023m1d1(9, 0), EndTime: y2023m1d1(10, 0)},  // Ends 10:00
		{ID: 2, StartTime: y2023m1d1(10, 30), EndTime: y2023m1d1(11, 0)}, // Starts 10:30 (30 min gap)
		{ID: 3, StartTime: y2023m1d1(14, 0), EndTime: y2023m1d1(15, 0)}, // Starts 14:00 (3 hour gap after entry 2 ends at 11:00)
		{ID: 4, StartTime: y2023m1d1(15, 0), EndTime: y2023m1d1(16, 0)}, // Starts 15:00 (0 min gap - consecutive)
		// Unsorted entry to test sorting
		{ID: 5, StartTime: y2023m1d1(11, 30), EndTime: y2023m1d1(12, 0)}, // Starts 11:30 (30 min gap after entry 2, before entry 3)
	}

	minGapDuration := 1 * time.Hour

	expectedGaps := []chronos.IdleGap{
		{
			StartTime: y2023m1d1(12, 0), // End of (sorted) entry ID 5
			EndTime:   y2023m1d1(14, 0), // Start of entry ID 3
			Duration:  2 * time.Hour,
		},
	}

	actualGaps := chronos.DetectIdleGaps(entries, minGapDuration)

	if len(actualGaps) != len(expectedGaps) {
		t.Fatalf("DetectIdleGaps: expected %d gaps, got %d gaps", len(expectedGaps), len(actualGaps))
	}

	// Note: reflect.DeepEqual might be tricky with time.Time if locations differ.
	// For simplicity, comparing fields.
	for i, eg := range expectedGaps {
		ag := actualGaps[i]
		if !ag.StartTime.Equal(eg.StartTime) || !ag.EndTime.Equal(eg.EndTime) || ag.Duration != eg.Duration {
			t.Errorf("DetectIdleGaps: gap %d mismatch.\nExpected: %+v\nActual:   %+v", i, eg, ag)
		}
	}

	// Test with insufficient entries
	if gaps := chronos.DetectIdleGaps(entries[:1], minGapDuration); len(gaps) != 0 {
		t.Errorf("DetectIdleGaps with 1 entry: expected 0 gaps, got %d", len(gaps))
	}
}
