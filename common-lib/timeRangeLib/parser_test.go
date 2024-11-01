/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package timeRangeLib

import (
	"testing"
	"time"
)

func TestGetScheduleSpec_FixedFrequency(t *testing.T) {

	testCases := []struct {
		description        string
		timeRange          TimeRange
		targetTime         time.Time
		expectedWindowEdge time.Time
		expectedIsBetween  bool
		expectedIsExpired  bool
		expectErr          bool
	}{
		{
			description: "Target time within the time range",
			timeRange: TimeRange{
				TimeFrom:  time.Date(2024, time.February, 26, 8, 0, 0, 0, time.UTC),
				TimeTo:    time.Date(2024, time.February, 26, 17, 0, 0, 0, time.UTC),
				Frequency: Fixed,
			},
			targetTime:         time.Date(2024, time.February, 26, 10, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.February, 26, 17, 0, 0, 0, time.UTC),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Target time before the time range",
			timeRange: TimeRange{
				TimeFrom:  time.Date(2024, time.February, 26, 8, 0, 0, 0, time.UTC),
				TimeTo:    time.Date(2024, time.February, 26, 17, 0, 0, 0, time.UTC),
				Frequency: Fixed,
			},
			targetTime:         time.Date(2024, time.February, 26, 7, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.February, 26, 8, 0, 0, 0, time.UTC),
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Target time after the time range",
			timeRange: TimeRange{
				TimeFrom:  time.Date(2024, time.February, 26, 8, 0, 0, 0, time.UTC),
				TimeTo:    time.Date(2024, time.February, 26, 17, 0, 0, 0, time.UTC),
				Frequency: Fixed,
			},
			targetTime:         time.Date(2024, time.February, 26, 18, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.February, 26, 17, 0, 0, 0, time.UTC),
			expectedIsBetween:  false,
			expectedIsExpired:  true,
			expectErr:          false,
		},
		{
			description: "Invalid time range (TimeFrom after TimeTo)",
			timeRange: TimeRange{
				TimeFrom:  time.Date(2024, time.February, 26, 18, 0, 0, 0, time.UTC),
				TimeTo:    time.Date(2024, time.February, 26, 8, 0, 0, 0, time.UTC),
				Frequency: Fixed,
			},
			targetTime:         time.Date(2024, time.February, 26, 10, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Time{},
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          true,
		},
		{
			description: "Target time exactly at TimeFrom boundary",
			timeRange: TimeRange{
				TimeFrom:  time.Date(2024, time.February, 26, 8, 0, 0, 0, time.UTC),
				TimeTo:    time.Date(2024, time.February, 26, 17, 0, 0, 0, time.UTC),
				Frequency: Fixed,
			},
			targetTime:         time.Date(2024, time.February, 26, 8, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.February, 26, 17, 0, 0, 0, time.UTC),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "Target time within daily window",
			timeRange: TimeRange{
				TimeFrom:       time.Date(2024, time.October, 2, 0, 0, 0, 0, time.UTC),
				TimeTo:         time.Date(2024, time.October, 13, 0, 0, 0, 0, time.UTC),
				Frequency:      Daily,
				HourMinuteFrom: "02:00",
				HourMinuteTo:   "18:00",
			},
			targetTime:         time.Date(2024, time.October, 3, 10, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.October, 3, 18, 0, 0, 0, time.UTC),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Target time before daily window",
			timeRange: TimeRange{
				TimeFrom:       time.Date(2024, time.October, 2, 0, 0, 0, 0, time.UTC),
				TimeTo:         time.Date(2024, time.October, 13, 0, 0, 0, 0, time.UTC),
				Frequency:      Daily,
				HourMinuteFrom: "02:00",
				HourMinuteTo:   "18:00",
			},
			targetTime:         time.Date(2024, time.October, 3, 1, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.October, 3, 2, 0, 0, 0, time.UTC),
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Target time after daily window",
			timeRange: TimeRange{
				TimeFrom:       time.Date(2024, time.October, 2, 0, 0, 0, 0, time.UTC),
				TimeTo:         time.Date(2024, time.October, 13, 0, 0, 0, 0, time.UTC),
				Frequency:      Daily,
				HourMinuteFrom: "02:00",
				HourMinuteTo:   "18:00",
			},
			targetTime:         time.Date(2024, time.October, 3, 19, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.October, 4, 02, 0, 0, 0, time.UTC),
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Target time outside the overall time range",
			timeRange: TimeRange{
				TimeFrom:       time.Date(2024, time.October, 2, 02, 0, 0, 0, time.UTC),
				TimeTo:         time.Date(2024, time.October, 13, 18, 0, 0, 0, time.UTC),
				Frequency:      Daily,
				HourMinuteFrom: "02:00",
				HourMinuteTo:   "18:00",
			},
			targetTime:         time.Date(2024, time.October, 14, 10, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.October, 13, 18, 0, 0, 0, time.UTC),
			expectedIsBetween:  false,
			expectedIsExpired:  true,
			expectErr:          false,
		},
		{
			description: "Target time exactly at start of daily window",
			timeRange: TimeRange{
				TimeFrom:       time.Date(2024, time.October, 2, 0, 0, 0, 0, time.UTC),
				TimeTo:         time.Date(2024, time.October, 13, 0, 0, 0, 0, time.UTC),
				Frequency:      Daily,
				HourMinuteFrom: "02:00",
				HourMinuteTo:   "18:00",
			},
			targetTime:         time.Date(2024, time.October, 3, 2, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.October, 3, 18, 0, 0, 0, time.UTC),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Target time exactly at end of daily window",
			timeRange: TimeRange{
				TimeFrom:       time.Date(2024, time.October, 2, 0, 0, 0, 0, time.UTC),
				TimeTo:         time.Date(2024, time.October, 13, 0, 0, 0, 0, time.UTC),
				Frequency:      Daily,
				HourMinuteFrom: "02:00",
				HourMinuteTo:   "18:00",
			},
			targetTime:         time.Date(2024, time.October, 3, 18, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.October, 4, 02, 0, 0, 0, time.UTC),
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "Infinite start time, target time within daily window",
			timeRange: TimeRange{
				TimeFrom:       time.Time{}, // Infinite start
				TimeTo:         time.Date(2024, time.October, 13, 0, 0, 0, 0, time.UTC),
				Frequency:      Daily,
				HourMinuteFrom: "02:00",
				HourMinuteTo:   "18:00",
			},
			targetTime:         time.Date(2024, time.October, 3, 10, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.October, 3, 18, 0, 0, 0, time.UTC),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Infinite end time, target time after daily window",
			timeRange: TimeRange{
				TimeFrom:       time.Date(2024, time.October, 2, 0, 0, 0, 0, time.UTC),
				TimeTo:         time.Time{}, // Infinite end
				Frequency:      Daily,
				HourMinuteFrom: "02:00",
				HourMinuteTo:   "18:00",
			},
			targetTime:         time.Date(2024, time.October, 3, 19, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.October, 4, 02, 0, 0, 0, time.UTC),
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Infinite start and end times, target within daily window",
			timeRange: TimeRange{
				TimeFrom:       time.Time{}, // Infinite start
				TimeTo:         time.Time{}, // Infinite end
				Frequency:      Daily,
				HourMinuteFrom: "02:00",
				HourMinuteTo:   "18:00",
			},
			targetTime:         time.Date(2024, time.October, 3, 10, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.October, 3, 18, 0, 0, 0, time.UTC),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Infinite start and end times, target outside daily window",
			timeRange: TimeRange{
				TimeFrom:       time.Time{}, // Infinite start
				TimeTo:         time.Time{}, // Infinite end
				Frequency:      Daily,
				HourMinuteFrom: "02:00",
				HourMinuteTo:   "18:00",
			},
			targetTime:         time.Date(2024, time.October, 3, 20, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.October, 4, 02, 0, 0, 0, time.UTC),
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "Weekly infinite recurrence - inside time range, HourMinuteFrom < HourMinuteTo",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "09:00",
				Weekdays:       []time.Weekday{time.Monday, time.Wednesday, time.Friday},
				Frequency:      Weekly,
			},
			targetTime:         time.Date(2024, time.February, 26, 8, 30, 0, 0, time.UTC), // Monday within range
			expectedWindowEdge: time.Date(2024, time.February, 26, 9, 0, 0, 0, time.UTC),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Weekly infinite recurrence - outside time range but not expired",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "09:00",
				Weekdays:       []time.Weekday{time.Monday, time.Wednesday, time.Friday},
				Frequency:      Weekly,
			},
			targetTime:         time.Date(2024, time.February, 26, 10, 0, 0, 0, time.UTC), // Monday but outside range
			expectedWindowEdge: time.Date(2024, time.February, 28, 8, 0, 0, 0, time.UTC),  // Next valid start on Wednesday
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Weekly recurrence with end date - expired",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "09:00",
				Weekdays:       []time.Weekday{time.Monday, time.Wednesday, time.Friday},
				Frequency:      Weekly,
				TimeTo:         time.Date(2024, time.February, 25, 9, 0, 0, 0, time.UTC), // Ended before targetTime
			},
			targetTime:         time.Date(2024, time.February, 26, 8, 0, 0, 0, time.UTC),
			expectedWindowEdge: time.Date(2024, time.February, 25, 9, 0, 0, 0, time.UTC), // Window expired, no valid edge
			expectedIsBetween:  false,
			expectedIsExpired:  true,
			expectErr:          false,
		},
		{
			description: "Weekly recurrence with end date - inside range and not expired",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "09:00",
				Weekdays:       []time.Weekday{time.Monday, time.Wednesday, time.Friday},
				Frequency:      Weekly,
				TimeTo:         time.Date(2024, time.February, 29, 0, 0, 0, 0, time.Local), // Ends after targetTime
			},
			targetTime:         time.Date(2024, time.February, 26, 8, 30, 0, 0, time.Local), // Monday within range
			expectedWindowEdge: time.Date(2024, time.February, 26, 9, 0, 0, 0, time.Local),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Weekly recurrence with end date - outside time range but not expired",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "09:00",
				Weekdays:       []time.Weekday{time.Monday, time.Wednesday, time.Friday},
				Frequency:      Weekly,
				TimeTo:         time.Date(2024, time.February, 29, 0, 0, 0, 0, time.Local), // Ends after targetTime
			},
			targetTime:         time.Date(2024, time.February, 26, 10, 0, 0, 0, time.Local), // Monday but outside range
			expectedWindowEdge: time.Date(2024, time.February, 28, 8, 0, 0, 0, time.Local),  // Next valid start on Wednesday
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Weekly recurrence - next valid weekday within range",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "09:00",
				Weekdays:       []time.Weekday{time.Monday, time.Wednesday, time.Friday},
				Frequency:      Weekly,
			},
			targetTime:         time.Date(2024, time.February, 27, 7, 0, 0, 0, time.Local), // Tuesday, before Wednesday window
			expectedWindowEdge: time.Date(2024, time.February, 28, 8, 0, 0, 0, time.Local), // Next window end on Wednesday
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Weekly recurrence - target time in previous window, next occurrence",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "09:00",
				Weekdays:       []time.Weekday{time.Monday, time.Wednesday, time.Friday},
				Frequency:      Weekly,
			},
			targetTime:         time.Date(2024, time.February, 26, 6, 0, 0, 0, time.Local), // Before Monday window
			expectedWindowEdge: time.Date(2024, time.February, 26, 8, 0, 0, 0, time.Local), // End of Monday window
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Weekly recurrence - target time after all valid windows in week but not expired",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "09:00",
				Weekdays:       []time.Weekday{time.Monday, time.Wednesday, time.Friday},
				Frequency:      Weekly,
			},
			targetTime:         time.Date(2024, time.March, 1, 10, 0, 0, 0, time.Local), // After last window in week (friday)
			expectedWindowEdge: time.Date(2024, time.March, 4, 8, 0, 0, 0, time.Local),  // Next Monday at 8:00
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "WeeklyRange recurrence without start or end dates - target within range",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "10:00",
				WeekdayFrom:    1, // Monday
				WeekdayTo:      3, // Wednesday
				Frequency:      WeeklyRange,
			},
			targetTime:         time.Date(2024, time.Month(2), 26, 8, 30, 0, 0, time.Local), // Within range on Monday
			expectedWindowEdge: time.Date(2024, time.Month(2), 28, 10, 0, 0, 0, time.Local), // End of range on Wednesday
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "WeeklyRange recurrence without start or end dates - target before range",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "10:00",
				WeekdayFrom:    1,
				WeekdayTo:      3,
				Frequency:      WeeklyRange,
			},
			targetTime:         time.Date(2024, time.Month(2), 25, 7, 30, 0, 0, time.Local), // Sunday before range starts
			expectedWindowEdge: time.Date(2024, time.Month(2), 26, 8, 0, 0, 0, time.Local),  // Start of next range on Monday
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "WeeklyRange recurrence without start or end dates - target after range",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "10:00",
				WeekdayFrom:    1,
				WeekdayTo:      3,
				Frequency:      WeeklyRange,
			},
			targetTime:         time.Date(2024, time.Month(2), 28, 11, 0, 0, 0, time.Local), // Wednesday after range ends
			expectedWindowEdge: time.Date(2024, time.Month(3), 4, 8, 0, 0, 0, time.Local),   // Start of next weekly range
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "WeeklyRange recurrence with end date - target within range before end date",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "10:00",
				WeekdayFrom:    1,
				WeekdayTo:      3,
				Frequency:      WeeklyRange,
				TimeTo:         time.Date(2024, time.Month(3), 1, 0, 0, 0, 0, time.Local), // End date is Friday
			},
			targetTime:         time.Date(2024, time.Month(2), 26, 9, 0, 0, 0, time.Local),  // Within range on Monday
			expectedWindowEdge: time.Date(2024, time.Month(2), 28, 10, 0, 0, 0, time.Local), // End of range on Wednesday
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "WeeklyRange recurrence with end date - target within range after end date",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "10:00",
				WeekdayFrom:    1,
				WeekdayTo:      3,
				Frequency:      WeeklyRange,
				TimeTo:         time.Date(2024, time.Month(2), 27, 10, 0, 0, 0, time.Local), // End date is Tuesday
			},
			targetTime:         time.Date(2024, time.Month(2), 28, 8, 30, 0, 0, time.Local),  // Wednesday after end date
			expectedWindowEdge: time.Date(2024, time.Month(2), 27, 10, 00, 0, 0, time.Local), // No recurrence as end date has passed
			expectedIsBetween:  false,
			expectedIsExpired:  true,
			expectErr:          false,
		},
		{
			description: "WeeklyRange recurrence with end date - target outside range before end date",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "10:00",
				WeekdayFrom:    1,
				WeekdayTo:      3,
				Frequency:      WeeklyRange,
				TimeTo:         time.Date(2024, time.Month(3), 1, 0, 0, 0, 0, time.Local),
			},
			targetTime:         time.Date(2024, time.Month(2), 25, 9, 0, 0, 0, time.Local), // Sunday before range starts
			expectedWindowEdge: time.Date(2024, time.Month(2), 26, 8, 0, 0, 0, time.Local), // Next start on Monday
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "WeeklyRange recurrence with start and end dates - target within range and period",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "10:00",
				WeekdayFrom:    1,
				WeekdayTo:      3,
				Frequency:      WeeklyRange,
				TimeFrom:       time.Date(2024, time.Month(2), 24, 0, 0, 0, 0, time.Local), // Saturday
				TimeTo:         time.Date(2024, time.Month(3), 1, 0, 0, 0, 0, time.Local),  // Friday
			},
			targetTime:         time.Date(2024, time.Month(2), 26, 8, 30, 0, 0, time.Local), // Within range on Monday
			expectedWindowEdge: time.Date(2024, time.Month(2), 28, 10, 0, 0, 0, time.Local), // End of range on Wednesday
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "WeeklyRange recurrence with start and end dates - target after end date",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "10:00",
				WeekdayFrom:    1,
				WeekdayTo:      3,
				Frequency:      WeeklyRange,
				TimeFrom:       time.Date(2024, time.Month(2), 24, 0, 0, 0, 0, time.Local),
				TimeTo:         time.Date(2024, time.Month(2), 27, 10, 0, 0, 0, time.Local), // Tuesday
			},
			targetTime:         time.Date(2024, time.Month(2), 28, 9, 0, 0, 0, time.Local),  // Wednesday after end date
			expectedWindowEdge: time.Date(2024, time.Month(2), 27, 10, 0, 0, 0, time.Local), // No recurrence as period has expired
			expectedIsBetween:  false,
			expectedIsExpired:  true,
			expectErr:          false,
		},
		{
			description: "WeeklyRange recurrence with start and end dates - target before start date",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "10:00",
				WeekdayFrom:    1,
				WeekdayTo:      3,
				Frequency:      WeeklyRange,
				TimeFrom:       time.Date(2024, time.Month(2), 26, 0, 0, 0, 0, time.Local), // Start date is Monday
				TimeTo:         time.Date(2024, time.Month(3), 1, 0, 0, 0, 0, time.Local),  // End date is Friday
			},
			targetTime:         time.Date(2024, time.Month(2), 25, 9, 0, 0, 0, time.Local), // Sunday before start date
			expectedWindowEdge: time.Date(2024, time.Month(2), 26, 8, 0, 0, 0, time.Local), // Next start on Monday
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "Target time between the time range with HourMinuteFrom > HourMinuteTo and DayTo > DayFrom (April)",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "08:00",
				DayFrom:        4,
				DayTo:          28,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2024, time.Month(4), 27, 10, 0, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2024, time.Month(4), 28, 8, 0, 0, 0, time.Local),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Target time between the time range with HourMinuteFrom > HourMinuteTo and DayFrom > DayTo (December)",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "11:00",
				DayFrom:        3,
				DayTo:          27,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2024, time.Month(12), 28, 10, 0, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2025, time.Month(1), 3, 9, 0, 0, 0, time.Local),
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "Target time within the time range with DayFrom < DayTo, target time for next month",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "10:00",
				DayFrom:        4,
				DayTo:          26,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2025, time.Month(1), 2, 10, 0, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2025, time.Month(1), 4, 8, 0, 0, 0, time.Local),
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Target time outside the time range with DayFrom < DayTo, target time for next month",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "08:00",
				DayFrom:        -3,
				DayTo:          -1,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2025, time.Month(1), 30, 10, 0, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2025, time.Month(1), 31, 8, 0, 0, 0, time.Local),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "Target time at the edge of the time range",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "10:00",
				DayFrom:        25,
				DayTo:          -3,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2024, time.Month(2), 24, 4, 0, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2024, time.Month(2), 25, 9, 0, 0, 0, time.Local),
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "Target time at the edge of the time range with HourMinuteFrom > HourMinuteTo",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "10:00",
				DayFrom:        4,
				DayTo:          -3,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2024, time.Month(2), 26, 9, 1, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2024, time.Month(2), 27, 10, 0, 0, 0, time.Local),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Target time before the time range with HourMinuteFrom > HourMinuteTo",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "10:00",
				DayFrom:        13,
				DayTo:          -2,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2024, time.Month(2), 14, 9, 30, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2024, time.Month(2), 28, 10, 0, 0, 0, time.Local),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "Target time with negative DayFrom and DayTo, time within range",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "10:00",
				DayFrom:        -3,
				DayTo:          -1,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2024, time.Month(1), 30, 10, 0, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2024, time.Month(1), 31, 10, 0, 0, 0, time.Local),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Target time outside range with negative DayFrom and DayTo",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "10:00",
				DayFrom:        -2,
				DayTo:          -1,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2025, time.Month(1), 30, 10, 0, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2025, time.Month(1), 31, 10, 0, 0, 0, time.Local),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "Target time after the time range with End Date",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "08:00",
				DayFrom:        4,
				DayTo:          28,
				Frequency:      Monthly,
				TimeTo:         time.Date(2024, time.Month(4), 16, 8, 0, 0, 0, time.Local),
			},
			targetTime:         time.Date(2024, time.Month(4), 27, 10, 0, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2024, time.Month(4), 16, 8, 0, 0, 0, time.Local),
			expectedIsBetween:  false,
			expectedIsExpired:  true,
			expectErr:          false,
		},
		{
			description: "Target time between the time range with HourMinuteFrom > HourMinuteTo and DayFrom > DayTo (December)",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "11:00",
				DayFrom:        3,
				DayTo:          27,
				Frequency:      Monthly,
				TimeTo:         time.Date(2024, time.Month(4), 26, 10, 0, 0, 0, time.Local),
			},
			targetTime:         time.Date(2024, time.Month(3), 28, 10, 0, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2024, time.Month(4), 3, 9, 0, 0, 0, time.Local),
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "Target time within the time range with DayFrom < DayTo, target time for next month",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "10:00",
				DayFrom:        4,
				DayTo:          26,
				Frequency:      Monthly,
				//TimeFrom: time.Date(2024, time.Month(3), 1, 10, 0, 0, 0, time.Local),
				TimeTo: time.Date(2024, time.Month(4), 28, 10, 0, 0, 0, time.Local),
			},
			targetTime:         time.Date(2024, time.Month(4), 27, 8, 30, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2024, time.Month(4), 26, 10, 0, 0, 0, time.Local),
			expectedIsBetween:  false,
			expectedIsExpired:  true,
			expectErr:          false,
		},
		{
			description: "Target time within the time range with DayFrom < DayTo, target time for next month",
			timeRange: TimeRange{
				HourMinuteFrom: "08:00",
				HourMinuteTo:   "10:00",
				DayFrom:        4,
				DayTo:          26,
				Frequency:      Monthly,
				TimeFrom:       time.Date(2024, time.Month(3), 1, 10, 0, 0, 0, time.Local),
				TimeTo:         time.Date(2024, time.Month(4), 28, 10, 0, 0, 0, time.Local),
			},
			targetTime:         time.Date(2024, time.Month(3), 2, 8, 30, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2024, time.Month(3), 4, 8, 0, 0, 0, time.Local),
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "Target time outside the time range with DayFrom < DayTo, target time for next month",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "08:00",
				DayFrom:        -3,
				DayTo:          -1,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2025, time.Month(1), 30, 10, 0, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2025, time.Month(1), 31, 8, 0, 0, 0, time.Local),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Target time at the edge of the time range",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "10:00",
				DayFrom:        25,
				DayTo:          -3,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2024, time.Month(2), 24, 4, 0, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2024, time.Month(2), 25, 9, 0, 0, 0, time.Local),
			expectedIsBetween:  false,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "Target time at the edge of the time range with HourMinuteFrom > HourMinuteTo",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "10:00",
				DayFrom:        3,
				DayTo:          -2,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2024, time.Month(3), 4, 10, 1, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2024, time.Month(3), 30, 10, 0, 0, 0, time.Local),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Target time before the time range with HourMinuteFrom > HourMinuteTo",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "10:00",
				DayFrom:        13,
				DayTo:          -2,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2024, time.Month(2), 14, 9, 30, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2024, time.Month(2), 28, 10, 0, 0, 0, time.Local),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},

		{
			description: "Target time with negative DayFrom and DayTo, time within range",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "10:00",
				DayFrom:        -3,
				DayTo:          -1,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2024, time.Month(1), 30, 10, 0, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2024, time.Month(1), 31, 10, 0, 0, 0, time.Local),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
		{
			description: "Target time outside range with negative DayFrom and DayTo",
			timeRange: TimeRange{
				HourMinuteFrom: "09:00",
				HourMinuteTo:   "10:00",
				DayFrom:        -2,
				DayTo:          -1,
				Frequency:      Monthly,
			},
			targetTime:         time.Date(2025, time.Month(1), 30, 10, 0, 0, 0, time.Local),
			expectedWindowEdge: time.Date(2025, time.Month(1), 31, 10, 0, 0, 0, time.Local),
			expectedIsBetween:  true,
			expectedIsExpired:  false,
			expectErr:          false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			nextWindowEdge, isTimeBetween, isExpired, err := tc.timeRange.GetTimeRangeWindow(tc.targetTime)

			if (err != nil) != tc.expectErr {
				t.Fatalf("Unexpected error: %v, expected error: %v", err, tc.expectErr)
			}

			if nextWindowEdge != tc.expectedWindowEdge || isTimeBetween != tc.expectedIsBetween || isExpired != tc.expectedIsExpired {
				t.Errorf("Test case failed: %s\nExpected nextWindowEdge: %v, got: %v\nExpected isTimeBetween: %t, got: %t\nExpected isExpired: %t, got: %t",
					tc.description, tc.expectedWindowEdge, nextWindowEdge, tc.expectedIsBetween, isTimeBetween, tc.expectedIsExpired, isExpired)
			}
		})
	}
}
