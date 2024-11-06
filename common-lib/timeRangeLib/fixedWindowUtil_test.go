package timeRangeLib

import (
	"testing"
	"time"
)

func TestGetWindowForFixedTime(t *testing.T) {
	// Define the time range in UTC , as db will be having this time in UTC/GMT
	tr := TimeRange{
		TimeFrom: time.Date(2024, 10, 14, 9, 0, 0, 0, time.UTC),  // 9:00 AM UTC
		TimeTo:   time.Date(2024, 10, 14, 17, 0, 0, 0, time.UTC), // 5:00 PM UTC
	}

	// Define the time zones with GMT offsets
	locBangui, _ := time.LoadLocation("Africa/Bangui")          // GMT+01:00
	locParamaribo, _ := time.LoadLocation("America/Paramaribo") // GMT-03:00
	// Define test cases with different time zones
	testCases := []struct {
		name          string
		targetTime    time.Time
		expectedStart time.Time
		expectedEnd   time.Time
	}{

		{
			name:          "Target time in Africa/Bangui (GMT+01:00), within range",
			targetTime:    time.Date(2024, 10, 14, 11, 30, 0, 0, locBangui), // 11:30 AM GMT+01 (within range)
			expectedStart: time.Date(2024, 10, 14, 10, 0, 0, 0, locBangui),  // 10:00 AM GMT+01 (adjusted)
			expectedEnd:   time.Date(2024, 10, 14, 18, 0, 0, 0, locBangui),  // 6:00 PM GMT+01 (adjusted)
		},
		{
			name:          "Target time in Africa/Bangui (GMT+01:00), before range",
			targetTime:    time.Date(2024, 10, 14, 8, 30, 0, 0, locBangui), // 11:30 AM GMT+01 (before range)
			expectedStart: time.Date(2024, 10, 14, 10, 0, 0, 0, locBangui), // 10:00 AM GMT+01 (adjusted)
			expectedEnd:   time.Date(2024, 10, 14, 18, 0, 0, 0, locBangui), // 6:00 PM GMT+01 (adjusted)
		},
		{
			name:          "Target time in Africa/Bangui (GMT+01:00), after the range (Outside)",
			targetTime:    time.Date(2024, 10, 14, 19, 30, 0, 0, locBangui), // 7:30 AM GMT+01 (outside)
			expectedStart: time.Date(2024, 10, 14, 10, 0, 0, 0, locBangui),  // zeroTime
			expectedEnd:   time.Date(2024, 10, 14, 18, 0, 0, 0, locBangui),  // zeroTime
		},

		{
			name:          "Target time in America/Paramaribo (GMT-03:00), within range",
			targetTime:    time.Date(2024, 10, 14, 07, 00, 0, 0, locParamaribo), // 7:00 AM GMT-03 (within range)
			expectedStart: time.Date(2024, 10, 14, 06, 00, 0, 0, locParamaribo), // 6:00 AM GMT-03 (adjusted)
			expectedEnd:   time.Date(2024, 10, 14, 14, 00, 0, 0, locParamaribo), // 2:00 PM GMT-03 (adjusted)
		},
		{
			name:          "Target time in America/Paramaribo (GMT-03:00), before range",
			targetTime:    time.Date(2024, 10, 14, 04, 00, 0, 0, locParamaribo), // 4:00 AM GMT-03 (within range)
			expectedStart: time.Date(2024, 10, 14, 06, 00, 0, 0, locParamaribo), // 6:00 AM GMT-03 (adjusted)
			expectedEnd:   time.Date(2024, 10, 14, 14, 00, 0, 0, locParamaribo), // 2:00 PM GMT-03 (adjusted)
		},
		{
			name:          "Target time in America/Paramaribo (GMT-03:00), after the range (Outside)",
			targetTime:    time.Date(2024, 10, 14, 14, 30, 0, 0, locParamaribo), // 2:30 PM GMT-03 (outside)
			expectedStart: time.Date(2024, 10, 14, 6, 0, 0, 0, locParamaribo),   // zeroTime
			expectedEnd:   time.Date(2024, 10, 14, 14, 0, 0, 0, locParamaribo),
		},
	}

	// Loop through the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			windowStart, windowEnd := tr.getWindowForFixedTime(tc.targetTime)
			if windowStart != tc.expectedStart || windowEnd != tc.expectedEnd {
				t.Errorf("Failed %s: got start=%v, end=%v; want start=%v, end=%v",
					tc.name, windowStart, windowEnd, tc.expectedStart, tc.expectedEnd)
			}
		})
	}
}
