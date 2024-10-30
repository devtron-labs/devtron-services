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
	"fmt"
	"github.com/robfig/cron/v3"
	"time"
)

func (tr TimeRange) GetTimeRangeWindow(targetTime time.Time) (nextWindowEdge time.Time, isTimeBetween, isExpired bool, err error) {
	err = tr.ValidateTimeRange()
	if err != nil {
		return nextWindowEdge, false, false, err
	}
	windowStart, windowEnd, err := tr.getWindowForTargetTime(targetTime)
	if err != nil {
		return nextWindowEdge, isTimeBetween, isExpired, err
	}
	if isTimeInBetween(targetTime, windowStart, windowEnd) {
		return windowEnd, true, false, nil
	}
	if targetTime.After(windowEnd) {
		return windowStart, false, true, nil
	}
	return windowStart, false, false, nil
}
func (tr TimeRange) getWindowForTargetTime(targetTime time.Time) (time.Time, time.Time, error) {

	if tr.Frequency == Fixed {
		windowStart, windowEnd := tr.getWindowForFixedTime(targetTime)
		return windowStart, windowEnd, nil
	}
	return tr.getWindowStartAndEndTime(targetTime)
}

// here target time is required to handle exceptions in monthly
// frequency where current time determines the cron and duration
func (tr TimeRange) getCronScheduleAndDuration(targetTime time.Time) (cron.Schedule, time.Duration, time.Duration, error) {

	evaluator := tr.getTimeRangeExpressionEvaluator(targetTime)
	cronExp := evaluator.getCron()
	parser := cron.NewParser(CRON)
	schedule, err := parser.Parse(cronExp)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("error parsing cron expression %s %v", cronExp, err)
	}
	duration := evaluator.getDuration()
	prevDuration := evaluator.getDurationOfPreviousWindow(duration)
	return schedule, duration, prevDuration, nil
}

func (tr TimeRange) getWindowStartAndEndTime(targetTime time.Time) (time.Time, time.Time, error) {

	var windowEnd time.Time
	schedule, duration, prevDuration, err := tr.getCronScheduleAndDuration(targetTime)
	if err != nil {
		return windowEnd, windowEnd, err
	}

	timeMinusDuration := tr.currentTimeMinusWindowDuration(targetTime, prevDuration)
	windowStart := schedule.Next(timeMinusDuration)
	windowEnd = windowStart.Add(duration)

	windowStart, windowEnd = tr.applyStartEndBoundary(windowStart, windowEnd)
	return windowStart, windowEnd, nil
}

/*
There are three possible cases for windowStart and windowEnd
relative to the defined time range (TimeFrom to TimeTo):

1. windowStart < TimeFrom
   - The window has not yet started.

2. windowStart > TimeTo
   - The occurrence has expired because the window starts after the end date.

3. windowEnd > TimeTo
   - The occurrence has expired because the window ends after the end date.

In each case, windowStart and windowEnd represent the next occurrence calculated based on the current time.
*/

func (tr TimeRange) applyStartEndBoundary(windowStart time.Time, windowEnd time.Time) (time.Time, time.Time) {
	if !tr.TimeFrom.IsZero() && windowStart.Before(tr.TimeFrom) {
		windowStart = tr.TimeFrom
	}
	if !tr.TimeTo.IsZero() && windowStart.After(tr.TimeTo) {
		windowStart = tr.TimeTo
	}
	if !tr.TimeTo.IsZero() && windowEnd.After(tr.TimeTo) {
		windowEnd = tr.TimeTo
	}
	return windowStart, windowEnd
}

func (tr TimeRange) currentTimeMinusWindowDuration(targetTime time.Time, duration time.Duration) time.Time {
	return targetTime.Add(-1 * duration)
}

// getWindowForFixedTime calculates the time window for a fixed frequency.
// Before comparing times, it is essential to ensure they are in the same time zone.
// The input targetTime is already converted to the correct time zone,(means the time zone in which they have saved mentioned explicitly)
// while tr.TimeFrom and tr.TimeTo are in GMT. Therefore, to compare them accurately, (in case of Fixed it happens)
// we convert the time range to match the time zone of targetTime.
func (tr TimeRange) getWindowForFixedTime(targetTime time.Time) (time.Time, time.Time) {
	timeFromInLocation := tr.TimeFrom.In(targetTime.Location())
	timeToInLocation := tr.TimeTo.In(targetTime.Location())
	return timeFromInLocation, timeToInLocation
}
func (tr TimeRange) SanitizeTimeFromAndTo(loc *time.Location) TimeRange {
	if !tr.TimeFrom.IsZero() {
		// Parse HourMinuteFrom (in HH:MM format)
		hourMinuteFrom, err := time.Parse("15:04", tr.HourMinuteFrom)
		if err != nil {
			fmt.Println("Error parsing HourMinuteFrom:", err)
			return tr // Return unchanged in case of an error
		}

		// Update TimeFrom with the parsed hour and minute while keeping date and location
		tr.TimeFrom = time.Date(
			tr.TimeFrom.Year(), tr.TimeFrom.Month(), tr.TimeFrom.Day(),
			hourMinuteFrom.Hour(), hourMinuteFrom.Minute(), 0, 0, loc,
		)

	}
	if !tr.TimeTo.IsZero() {
		// Parse HourMinuteTo (in HH:MM format)
		hourMinuteTo, err := time.Parse("15:04", tr.HourMinuteTo)
		if err != nil {
			fmt.Println("Error parsing HourMinuteTo:", err)
			return tr // Return unchanged in case of an error
		}
		// Update TimeTo with the parsed hour and minute while keeping date and location
		tr.TimeTo = time.Date(
			tr.TimeTo.Year(), tr.TimeTo.Month(), tr.TimeTo.Day(),
			hourMinuteTo.Hour(), hourMinuteTo.Minute(), 0, 0, loc,
		)

	}
	return tr
}
