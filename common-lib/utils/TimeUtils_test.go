package utils

import (
	"testing"
	"time"
)

func TestParseAndValidateTimeRange(t *testing.T) {
	// Test nil input
	t.Run("nil input", func(t *testing.T) {
		var timeRange *TimeRangeRequest
		result, err := timeRange.ParseAndValidateTimeRange()
		if err == nil {
			t.Error("Expected error for nil input")
		}
		if result == nil {
			t.Error("Expected non-nil result")
		}
	})

	// Test TimeWindow cases
	t.Run("Today timeWindow", func(t *testing.T) {
		timeWindow := Today
		timeRange := &TimeRangeRequest{TimeWindow: &timeWindow}
		result, err := timeRange.ParseAndValidateTimeRange()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.From == nil || result.To == nil {
			t.Error("Expected non-nil From and To times")
		}
	})

	t.Run("Yesterday timeWindow", func(t *testing.T) {
		timeWindow := Yesterday
		timeRange := &TimeRangeRequest{TimeWindow: &timeWindow}
		result, err := timeRange.ParseAndValidateTimeRange()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.From == nil || result.To == nil {
			t.Error("Expected non-nil From and To times")
		}
	})

	t.Run("Week timeWindow", func(t *testing.T) {
		timeWindow := Week
		timeRange := &TimeRangeRequest{TimeWindow: &timeWindow}
		result, err := timeRange.ParseAndValidateTimeRange()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.From == nil || result.To == nil {
			t.Error("Expected non-nil From and To times")
		}
	})

	t.Run("Month timeWindow", func(t *testing.T) {
		timeWindow := Month
		timeRange := &TimeRangeRequest{TimeWindow: &timeWindow}
		result, err := timeRange.ParseAndValidateTimeRange()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.From == nil || result.To == nil {
			t.Error("Expected non-nil From and To times")
		}
	})

	t.Run("Quarter timeWindow", func(t *testing.T) {
		timeWindow := Quarter
		timeRange := &TimeRangeRequest{TimeWindow: &timeWindow}
		result, err := timeRange.ParseAndValidateTimeRange()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.From == nil || result.To == nil {
			t.Error("Expected non-nil From and To times")
		}
	})

	t.Run("LastWeek timeWindow", func(t *testing.T) {
		timeWindow := LastWeek
		timeRange := &TimeRangeRequest{TimeWindow: &timeWindow}
		result, err := timeRange.ParseAndValidateTimeRange()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.From == nil || result.To == nil {
			t.Error("Expected non-nil From and To times")
		}
	})

	t.Run("LastMonth timeWindow", func(t *testing.T) {
		timeWindow := LastMonth
		timeRange := &TimeRangeRequest{TimeWindow: &timeWindow}
		result, err := timeRange.ParseAndValidateTimeRange()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if result.From == nil || result.To == nil {
			t.Error("Expected non-nil From and To times")
		}
	})

	t.Run("invalid timeWindow", func(t *testing.T) {
		invalidWindow := TimeWindows("invalid")
		timeRange := &TimeRangeRequest{TimeWindow: &invalidWindow}
		_, err := timeRange.ParseAndValidateTimeRange()
		if err == nil {
			t.Error("Expected error for invalid timeWindow")
		}
	})

	// Test From/To date cases
	t.Run("valid From and To dates", func(t *testing.T) {
		from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2023, 1, 31, 23, 59, 59, 0, time.UTC)
		timeRange := &TimeRangeRequest{From: &from, To: &to}
		result, err := timeRange.ParseAndValidateTimeRange()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !result.From.Equal(from) || !result.To.Equal(to) {
			t.Error("From and To dates should match input")
		}
	})

	t.Run("From date after To date", func(t *testing.T) {
		from := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
		to := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		timeRange := &TimeRangeRequest{From: &from, To: &to}
		_, err := timeRange.ParseAndValidateTimeRange()
		if err == nil {
			t.Error("Expected error when From date is after To date")
		}
	})

	t.Run("missing From date", func(t *testing.T) {
		to := time.Date(2023, 1, 31, 0, 0, 0, 0, time.UTC)
		timeRange := &TimeRangeRequest{To: &to}
		_, err := timeRange.ParseAndValidateTimeRange()
		if err == nil {
			t.Error("Expected error when From date is missing")
		}
	})

	t.Run("missing To date", func(t *testing.T) {
		from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
		timeRange := &TimeRangeRequest{From: &from}
		_, err := timeRange.ParseAndValidateTimeRange()
		if err == nil {
			t.Error("Expected error when To date is missing")
		}
	})

	t.Run("no timeWindow and no From/To dates", func(t *testing.T) {
		timeRange := &TimeRangeRequest{}
		_, err := timeRange.ParseAndValidateTimeRange()
		if err == nil {
			t.Error("Expected error when neither timeWindow nor From/To dates are provided")
		}
	})
}

func TestGetWeeklyTimeBoundaries(t *testing.T) {
	t.Run("zero iterations", func(t *testing.T) {
		result := GetWeeklyTimeBoundaries(0)
		if len(result) != 0 {
			t.Errorf("Expected empty slice for 0 iterations, got %d boundaries", len(result))
		}
	})

	t.Run("single iteration", func(t *testing.T) {
		result := GetWeeklyTimeBoundaries(1)
		if len(result) != 1 {
			t.Errorf("Expected 1 boundary for 1 iteration, got %d", len(result))
		}

		boundary := result[0]
		if boundary.StartTime.After(boundary.EndTime) {
			t.Error("StartTime should not be after EndTime")
		}

		// Check that the week starts on Monday
		if boundary.StartTime.Weekday() != time.Monday {
			t.Errorf("Expected week to start on Monday, got %v", boundary.StartTime.Weekday())
		}
	})

	t.Run("multiple iterations", func(t *testing.T) {
		iterations := 3
		result := GetWeeklyTimeBoundaries(iterations)

		if len(result) != iterations {
			t.Errorf("Expected %d boundaries, got %d", iterations, len(result))
		}

		// Check that boundaries are consecutive weeks going backwards
		for i := 0; i < len(result)-1; i++ {
			currentWeek := result[i]
			nextWeek := result[i+1]

			// Current week's start should equal next week's end
			if !currentWeek.StartTime.Equal(nextWeek.EndTime) {
				t.Errorf("Week %d start (%v) should equal week %d end (%v)",
					i, currentWeek.StartTime, i+1, nextWeek.EndTime)
			}

			// Each week should start on Monday
			if currentWeek.StartTime.Weekday() != time.Monday {
				t.Errorf("Week %d should start on Monday, got %v", i, currentWeek.StartTime.Weekday())
			}
		}
	})

	t.Run("negative iterations", func(t *testing.T) {
		result := GetWeeklyTimeBoundaries(-1)
		if len(result) != 0 {
			t.Errorf("Expected empty slice for negative iterations, got %d boundaries", len(result))
		}
	})

	t.Run("boundary validation", func(t *testing.T) {
		result := GetWeeklyTimeBoundaries(2)

		for i, boundary := range result {
			// StartTime should not be after EndTime
			if boundary.StartTime.After(boundary.EndTime) {
				t.Errorf("Boundary %d: StartTime (%v) should not be after EndTime (%v)",
					i, boundary.StartTime, boundary.EndTime)
			}

			// StartTime should be on Monday
			if boundary.StartTime.Weekday() != time.Monday {
				t.Errorf("Boundary %d: StartTime should be on Monday, got %v",
					i, boundary.StartTime.Weekday())
			}

			// Time difference should be less than 7 days (one week)
			duration := boundary.EndTime.Sub(boundary.StartTime)
			if duration >= 7*24*time.Hour && boundary.EndTime.Sub(time.Now()) > 10*time.Minute {
				t.Errorf("Boundary %d: Duration (%v) should be less than 7 days", i, duration)
			}
		}
	})

	t.Run("chronological order", func(t *testing.T) {
		result := GetWeeklyTimeBoundaries(4)

		// Boundaries should be in reverse chronological order (most recent first)
		for i := 0; i < len(result)-1; i++ {
			current := result[i]
			next := result[i+1]

			if current.StartTime.Before(next.StartTime) {
				t.Errorf("Boundaries should be in reverse chronological order: boundary %d start (%v) should be after boundary %d start (%v)",
					i, current.StartTime, i+1, next.StartTime)
			}
		}
	})
}

func TestGetMonthlyTimeBoundaries(t *testing.T) {
	t.Run("zero iterations", func(t *testing.T) {
		result := GetMonthlyTimeBoundaries(0)
		if len(result) != 0 {
			t.Errorf("Expected empty slice for 0 iterations, got %d boundaries", len(result))
		}
	})

	t.Run("single iteration", func(t *testing.T) {
		result := GetMonthlyTimeBoundaries(1)
		if len(result) != 1 {
			t.Errorf("Expected 1 boundary for 1 iteration, got %d", len(result))
		}

		boundary := result[0]
		if boundary.StartTime.After(boundary.EndTime) {
			t.Error("StartTime should not be after EndTime")
		}

		// Check that the month starts on the 1st
		if boundary.StartTime.Day() != 1 {
			t.Errorf("Expected month to start on the 1st, got day %d", boundary.StartTime.Day())
		}
	})

	t.Run("multiple iterations", func(t *testing.T) {
		iterations := 3
		result := GetMonthlyTimeBoundaries(iterations)

		if len(result) != iterations {
			t.Errorf("Expected %d boundaries, got %d", iterations, len(result))
		}

		// Check that boundaries are consecutive months going backwards
		for i := 0; i < len(result)-1; i++ {
			currentMonth := result[i]
			nextMonth := result[i+1]

			// Current month's start should equal next month's end
			if !currentMonth.StartTime.Equal(nextMonth.EndTime) {
				t.Errorf("Month %d start (%v) should equal month %d end (%v)",
					i, currentMonth.StartTime, i+1, nextMonth.EndTime)
			}

			// Each month should start on the 1st
			if currentMonth.StartTime.Day() != 1 {
				t.Errorf("Month %d should start on the 1st, got day %d", i, currentMonth.StartTime.Day())
			}
		}
	})

	t.Run("negative iterations", func(t *testing.T) {
		result := GetMonthlyTimeBoundaries(-1)
		if len(result) != 0 {
			t.Errorf("Expected empty slice for negative iterations, got %d boundaries", len(result))
		}
	})

	t.Run("boundary validation", func(t *testing.T) {
		result := GetMonthlyTimeBoundaries(2)

		for i, boundary := range result {
			// StartTime should not be after EndTime
			if boundary.StartTime.After(boundary.EndTime) {
				t.Errorf("Boundary %d: StartTime (%v) should not be after EndTime (%v)",
					i, boundary.StartTime, boundary.EndTime)
			}

			// StartTime should be on the 1st
			if boundary.StartTime.Day() != 1 {
				t.Errorf("Boundary %d: StartTime should be on the 1st, got day %d",
					i, boundary.StartTime.Day())
			}

			// EndTime should be the 1st of the next month
			expectedEnd := boundary.StartTime.AddDate(0, 1, 0)
			now := time.Now()
			// For current month, end might be 'now' if we're still in the month
			if i == 0 && now.Before(expectedEnd) {
				expectedEnd = now
			}
			if !boundary.EndTime.Equal(expectedEnd) && boundary.EndTime.Sub(time.Now()) > 10*time.Minute {
				t.Errorf("Boundary %d: EndTime (%v) should be %v",
					i, boundary.EndTime, expectedEnd)
			}
		}
	})

	t.Run("chronological order", func(t *testing.T) {
		result := GetMonthlyTimeBoundaries(4)

		// Boundaries should be in reverse chronological order (most recent first)
		for i := 0; i < len(result)-1; i++ {
			current := result[i]
			next := result[i+1]

			if current.StartTime.Before(next.StartTime) {
				t.Errorf("Boundaries should be in reverse chronological order: boundary %d start (%v) should be after boundary %d start (%v)",
					i, current.StartTime, i+1, next.StartTime)
			}
		}
	})

	t.Run("current month adjustment", func(t *testing.T) {
		result := GetMonthlyTimeBoundaries(1)
		boundary := result[0]
		now := time.Now()

		// For the current month, end time should not exceed now
		if boundary.EndTime.After(now) {
			t.Errorf("Current month end time (%v) should not be after now (%v)",
				boundary.EndTime, now)
		}

		// Start time should be the beginning of current month
		expectedStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		if !boundary.StartTime.Equal(expectedStart) {
			t.Errorf("Current month start time (%v) should be %v",
				boundary.StartTime, expectedStart)
		}
	})
}

func TestGetQuarterlyTimeBoundaries(t *testing.T) {
	t.Run("zero iterations", func(t *testing.T) {
		result := GetQuarterlyTimeBoundaries(0)
		if len(result) != 0 {
			t.Errorf("Expected empty slice for 0 iterations, got %d boundaries", len(result))
		}
	})

	t.Run("single iteration", func(t *testing.T) {
		result := GetQuarterlyTimeBoundaries(1)
		if len(result) != 1 {
			t.Errorf("Expected 1 boundary for 1 iteration, got %d", len(result))
		}

		boundary := result[0]
		if boundary.StartTime.After(boundary.EndTime) {
			t.Error("StartTime should not be after EndTime")
		}

		// Check that the quarter starts on the 1st
		if boundary.StartTime.Day() != 1 {
			t.Errorf("Expected quarter to start on the 1st, got day %d", boundary.StartTime.Day())
		}

		// Check that it starts at the beginning of a quarter month (Jan, Apr, Jul, Oct)
		expectedQuarterMonths := []time.Month{time.January, time.April, time.July, time.October}
		found := false
		for _, month := range expectedQuarterMonths {
			if boundary.StartTime.Month() == month {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected quarter to start in Jan/Apr/Jul/Oct, got %v", boundary.StartTime.Month())
		}
	})

	t.Run("multiple iterations", func(t *testing.T) {
		iterations := 3
		result := GetQuarterlyTimeBoundaries(iterations)

		if len(result) != iterations {
			t.Errorf("Expected %d boundaries, got %d", iterations, len(result))
		}

		// Check that boundaries are consecutive quarters going backwards
		for i := 0; i < len(result)-1; i++ {
			currentQuarter := result[i]
			nextQuarter := result[i+1]

			// Current quarter's start should equal next quarter's end
			if !currentQuarter.StartTime.Equal(nextQuarter.EndTime) {
				t.Errorf("Quarter %d start (%v) should equal quarter %d end (%v)",
					i, currentQuarter.StartTime, i+1, nextQuarter.EndTime)
			}

			// Each quarter should start on the 1st
			if currentQuarter.StartTime.Day() != 1 {
				t.Errorf("Quarter %d should start on the 1st, got day %d", i, currentQuarter.StartTime.Day())
			}
		}
	})

	t.Run("negative iterations", func(t *testing.T) {
		result := GetQuarterlyTimeBoundaries(-1)
		if len(result) != 0 {
			t.Errorf("Expected empty slice for negative iterations, got %d boundaries", len(result))
		}
	})

	t.Run("boundary validation", func(t *testing.T) {
		result := GetQuarterlyTimeBoundaries(2)

		for i, boundary := range result {
			// StartTime should not be after EndTime
			if boundary.StartTime.After(boundary.EndTime) {
				t.Errorf("Boundary %d: StartTime (%v) should not be after EndTime (%v)",
					i, boundary.StartTime, boundary.EndTime)
			}

			// StartTime should be on the 1st
			if boundary.StartTime.Day() != 1 {
				t.Errorf("Boundary %d: StartTime should be on the 1st, got day %d",
					i, boundary.StartTime.Day())
			}

			// StartTime should be at beginning of quarter (Jan/Apr/Jul/Oct)
			expectedQuarterMonths := []time.Month{time.January, time.April, time.July, time.October}
			found := false
			for _, month := range expectedQuarterMonths {
				if boundary.StartTime.Month() == month {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Boundary %d: StartTime should be in Jan/Apr/Jul/Oct, got %v",
					i, boundary.StartTime.Month())
			}

			// Duration should be approximately 3 months
			expectedEnd := boundary.StartTime.AddDate(0, 3, 0)
			now := time.Now()
			// For current quarter, end might be 'now' if we're still in the quarter
			if i == 0 && now.Before(expectedEnd) {
				expectedEnd = now
			}
			if !boundary.EndTime.Equal(expectedEnd) && boundary.EndTime.Sub(time.Now()) > 10*time.Minute {
				t.Errorf("Boundary %d: EndTime (%v) should be %v",
					i, boundary.EndTime, expectedEnd)
			}
		}
	})

	t.Run("chronological order", func(t *testing.T) {
		result := GetQuarterlyTimeBoundaries(4)

		// Boundaries should be in reverse chronological order (most recent first)
		for i := 0; i < len(result)-1; i++ {
			current := result[i]
			next := result[i+1]

			if current.StartTime.Before(next.StartTime) {
				t.Errorf("Boundaries should be in reverse chronological order: boundary %d start (%v) should be after boundary %d start (%v)",
					i, current.StartTime, i+1, next.StartTime)
			}
		}
	})

	t.Run("current quarter adjustment", func(t *testing.T) {
		result := GetQuarterlyTimeBoundaries(1)
		boundary := result[0]
		now := time.Now()

		// For the current quarter, end time should not exceed now
		if boundary.EndTime.After(now) {
			t.Errorf("Current quarter end time (%v) should not be after now (%v)",
				boundary.EndTime, now)
		}

		// Start time should be the beginning of current quarter
		quarter := ((int(now.Month()) - 1) / 3) + 1
		quarterMonth := time.Month((quarter-1)*3 + 1)
		expectedStart := time.Date(now.Year(), quarterMonth, 1, 0, 0, 0, 0, now.Location())
		if !boundary.StartTime.Equal(expectedStart) {
			t.Errorf("Current quarter start time (%v) should be %v",
				boundary.StartTime, expectedStart)
		}
	})

	t.Run("quarter calculation accuracy", func(t *testing.T) {
		result := GetQuarterlyTimeBoundaries(4)

		for i, boundary := range result {
			// Verify quarter months are correct
			month := boundary.StartTime.Month()
			switch month {
			case time.January:
				// Q1: Jan-Mar
				expectedEnd := boundary.StartTime.AddDate(0, 3, 0)
				if i == 0 && time.Now().Before(expectedEnd) {
					expectedEnd = time.Now()
				}
				if expectedEnd.Month() != time.April && expectedEnd.Sub(time.Now()) > 10*time.Minute {
					t.Errorf("Q1 boundary %d should end in April, got %v", i, expectedEnd.Month())
				}
			case time.April:
				// Q2: Apr-Jun
				expectedEnd := boundary.StartTime.AddDate(0, 3, 0)
				if i == 0 && time.Now().Before(expectedEnd) {
					expectedEnd = time.Now()
				}
				if expectedEnd.Month() != time.July && expectedEnd.Sub(time.Now()) > 10*time.Minute {
					t.Errorf("Q2 boundary %d should end in July, got %v", i, expectedEnd.Month())
				}
			case time.July:
				// Q3: Jul-Sep
				expectedEnd := boundary.StartTime.AddDate(0, 3, 0)
				if i == 0 && time.Now().Before(expectedEnd) {
					expectedEnd = time.Now()
				}
				if expectedEnd.Month() != time.October && expectedEnd.Sub(time.Now()) > 10*time.Minute {
					t.Errorf("Q3 boundary %d should end in October, got %v", i, expectedEnd.Month())
				}
			case time.October:
				// Q4: Oct-Dec
				expectedEnd := boundary.StartTime.AddDate(0, 3, 0)
				if i == 0 && time.Now().Before(expectedEnd) {
					expectedEnd = time.Now()
				}
				if expectedEnd.Month() != time.January && expectedEnd.Sub(time.Now()) > 10*time.Minute {
					t.Errorf("Q4 boundary %d should end in January of next year, got %v", i, expectedEnd.Month())
				}
			default:
				t.Errorf("Boundary %d starts in invalid quarter month: %v", i, month)
			}
		}
	})
}

func TestGetYearlyTimeBoundaries(t *testing.T) {
	t.Run("zero iterations", func(t *testing.T) {
		result := GetYearlyTimeBoundaries(0)
		if len(result) != 0 {
			t.Errorf("Expected empty slice for 0 iterations, got %d boundaries", len(result))
		}
	})

	t.Run("single iteration", func(t *testing.T) {
		result := GetYearlyTimeBoundaries(1)
		if len(result) != 1 {
			t.Errorf("Expected 1 boundary for 1 iteration, got %d", len(result))
		}

		boundary := result[0]
		if boundary.StartTime.After(boundary.EndTime) {
			t.Error("StartTime should not be after EndTime")
		}

		// Check that the year starts on January 1st
		if boundary.StartTime.Month() != time.January || boundary.StartTime.Day() != 1 {
			t.Errorf("Expected year to start on January 1st, got %v %d", boundary.StartTime.Month(), boundary.StartTime.Day())
		}
	})

	t.Run("multiple iterations", func(t *testing.T) {
		iterations := 3
		result := GetYearlyTimeBoundaries(iterations)

		if len(result) != iterations {
			t.Errorf("Expected %d boundaries, got %d", iterations, len(result))
		}

		// Check that boundaries are consecutive years going backwards
		for i := 0; i < len(result)-1; i++ {
			currentYear := result[i]
			nextYear := result[i+1]

			// Current year's start should equal next year's end
			if !currentYear.StartTime.Equal(nextYear.EndTime) {
				t.Errorf("Year %d start (%v) should equal year %d end (%v)",
					i, currentYear.StartTime, i+1, nextYear.EndTime)
			}

			// Each year should start on January 1st
			if currentYear.StartTime.Month() != time.January || currentYear.StartTime.Day() != 1 {
				t.Errorf("Year %d should start on January 1st, got %v %d", i, currentYear.StartTime.Month(), currentYear.StartTime.Day())
			}
		}
	})

	t.Run("negative iterations", func(t *testing.T) {
		result := GetYearlyTimeBoundaries(-1)
		if len(result) != 0 {
			t.Errorf("Expected empty slice for negative iterations, got %d boundaries", len(result))
		}
	})

	t.Run("boundary validation", func(t *testing.T) {
		result := GetYearlyTimeBoundaries(2)

		for i, boundary := range result {
			// StartTime should not be after EndTime
			if boundary.StartTime.After(boundary.EndTime) {
				t.Errorf("Boundary %d: StartTime (%v) should not be after EndTime (%v)",
					i, boundary.StartTime, boundary.EndTime)
			}

			// StartTime should be on January 1st
			if boundary.StartTime.Month() != time.January || boundary.StartTime.Day() != 1 {
				t.Errorf("Boundary %d: StartTime should be on January 1st, got %v %d",
					i, boundary.StartTime.Month(), boundary.StartTime.Day())
			}

			// EndTime should be January 1st of the next year
			expectedEnd := boundary.StartTime.AddDate(1, 0, 0)
			now := time.Now()
			// For current year, end might be 'now' if we're still in the year
			if i == 0 && now.Before(expectedEnd) {
				expectedEnd = now
			}
			if !boundary.EndTime.Equal(expectedEnd) && boundary.EndTime.Sub(time.Now()) > 10*time.Minute {
				t.Errorf("Boundary %d: EndTime (%v) should be %v",
					i, boundary.EndTime, expectedEnd)
			}
		}
	})

	t.Run("chronological order", func(t *testing.T) {
		result := GetYearlyTimeBoundaries(4)

		// Boundaries should be in reverse chronological order (most recent first)
		for i := 0; i < len(result)-1; i++ {
			current := result[i]
			next := result[i+1]

			if current.StartTime.Before(next.StartTime) {
				t.Errorf("Boundaries should be in reverse chronological order: boundary %d start (%v) should be after boundary %d start (%v)",
					i, current.StartTime, i+1, next.StartTime)
			}
		}
	})

	t.Run("current year adjustment", func(t *testing.T) {
		result := GetYearlyTimeBoundaries(1)
		boundary := result[0]
		now := time.Now()

		// For the current year, end time should not exceed now
		if boundary.EndTime.After(now) {
			t.Errorf("Current year end time (%v) should not be after now (%v)",
				boundary.EndTime, now)
		}

		// Start time should be the beginning of current year
		expectedStart := time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
		if !boundary.StartTime.Equal(expectedStart) {
			t.Errorf("Current year start time (%v) should be %v",
				boundary.StartTime, expectedStart)
		}
	})

	t.Run("year progression accuracy", func(t *testing.T) {
		result := GetYearlyTimeBoundaries(5)
		now := time.Now()

		for i, boundary := range result {
			expectedYear := now.Year() - i
			if boundary.StartTime.Year() != expectedYear {
				t.Errorf("Boundary %d should be for year %d, got %d",
					i, expectedYear, boundary.StartTime.Year())
			}

			// Verify it's exactly January 1st of that year
			expectedStart := time.Date(expectedYear, time.January, 1, 0, 0, 0, 0, now.Location())
			if !boundary.StartTime.Equal(expectedStart) {
				t.Errorf("Boundary %d start time (%v) should be exactly %v",
					i, boundary.StartTime, expectedStart)
			}
		}
	})
}
