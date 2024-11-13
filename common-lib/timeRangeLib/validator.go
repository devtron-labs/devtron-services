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
	"errors"
	"golang.org/x/exp/slices"
	"strconv"
	"strings"
)

func (tr TimeRange) ValidateTimeRange() error {
	if !slices.Contains(AllowedFrequencies, tr.Frequency) {
		return errors.New(string(InvalidFrequencyType))
	}
	if tr.Frequency != Fixed {
		err := validateHourMinute(tr.HourMinuteFrom)
		if err != nil {
			return err
		}
		err = validateHourMinute(tr.HourMinuteTo)
		if err != nil {
			return err
		}

		// Ensure HourMinuteFrom < HourMinuteTo for Daily and Weekly frequencies
		if tr.Frequency == Daily || tr.Frequency == Weekly {
			if isToBeforeFrom(tr.HourMinuteFrom, tr.HourMinuteTo) {
				return errors.New(string(InvalidHourMinuteForWeeklyAndDaily))
			}
		}
	}

	if !tr.TimeFrom.IsZero() && !tr.TimeTo.IsZero() {
		if tr.Frequency == Fixed {
			if tr.TimeFrom.After(tr.TimeTo) {
				return errors.New(string(TimeFromLessThanTimeTo))
			}
			if tr.TimeFrom.Equal(tr.TimeTo) {
				return errors.New(string(TimeFromEqualToTimeTo))
			}
		} else {
			// Validate that the date range is valid (same date safe for recurrence type windows)
			if !isDateFromBeforeTo(tr.TimeFrom, tr.TimeTo) {
				return errors.New(string(DayFromOrToNotValid))
			}
		}
	}

	switch tr.Frequency {
	case Daily:
		if tr.HourMinuteFrom == "" || tr.HourMinuteTo == "" {
			return errors.New(string(HourMinuteNotPresent))
		}
	case Fixed:
		if tr.TimeFrom.IsZero() || tr.TimeTo.IsZero() {
			return errors.New(string(TimeFromOrToNotPresent))
		}
	case Weekly:
		if len(tr.Weekdays) == 0 {
			return errors.New(string(WeekDaysNotPresent))
		}
		for _, weekday := range tr.Weekdays {
			if weekday < 0 || weekday > 6 {
				return errors.New(string(WeekDayOutsideRange))
			}
		}
	case WeeklyRange:
		if tr.WeekdayFrom == 0 || tr.WeekdayTo == 0 {
			return errors.New(string(WeekDayFromOrToNotPresent))
		}
		if (tr.WeekdayFrom < 0 || tr.WeekdayFrom > 6) || (tr.WeekdayTo < 0 || tr.WeekdayTo > 6) {
			return errors.New(string(WeekDayOutsideRange))
		}
	case Monthly:
		if tr.DayFrom == 0 || tr.DayTo == 0 {
			return errors.New(string(DayFromOrToNotPresent))
		}
		if tr.DayFrom == tr.DayTo && isToBeforeFrom(tr.HourMinuteFrom, tr.HourMinuteTo) {
			return errors.New(string(ToBeforeFrom))
		}
		// DayFrom > DayTo => check for the correct value
		if tr.DayFrom > 0 && tr.DayTo > 0 && tr.DayFrom > tr.DayTo {
			return errors.New(string(DayFromOrToNotValid))
		}
		if tr.DayFrom < 0 && tr.DayTo > 0 {
			return errors.New(string(DayFromOrToNotValid))
		}

		// this is to prevent overlapping windows crossing to next month for both negatives
		if tr.DayFrom < 0 && tr.DayTo < 0 && tr.DayFrom > tr.DayTo {
			return errors.New(string(BothLessThanZeroAndFromGreaterThanTo))
		}
		// this is an edge case where with negative 'to' date results into a date before the 'from' date
		// example: 26,-4 will pe prevented because for February it will become invalid
		// also currently max negative supported is third last day of the month
		if (tr.DayTo < 0 && tr.DayFrom > 0 && tr.DayFrom > 29+tr.DayTo) || tr.DayTo < -3 || tr.DayFrom < -3 {
			return errors.New(string(DayFromOrToNotValid))
		}
	}
	return nil
}

func validateHourMinute(hourMinute string) error {
	parts := strings.Split(hourMinute, ":")
	if len(parts) != 2 {
		return errors.New("HourMinute is not valid, should be strictly of format HH:MM")
	}
	hh, err := strconv.Atoi(parts[0])
	if err != nil || hh > 23 || hh < 0 {
		return errors.New("Hour is not valid" + parts[0])
	}

	mm, err := strconv.Atoi(parts[1])
	if err != nil || mm > 59 || mm < 0 {
		return errors.New("Minute is not valid" + parts[1])
	}
	return nil
}
