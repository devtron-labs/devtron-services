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
	"strconv"
	"time"
)

func isToHourMinuteBeforeWindowEnd(hourMinute string, targetTime time.Time) bool {

	currentHourMinute, _ := time.Parse(hourMinuteFormat, targetTime.Format(hourMinuteFormat))
	parsedHourTo, _ := time.Parse(hourMinuteFormat, hourMinute)

	return currentHourMinute.Before(parsedHourTo)
}

func getLastDayOfMonth(targetYear int, targetMonth time.Month) int {
	firstDayOfNextMonth := time.Date(targetYear, targetMonth+1, 1, 0, 0, 0, 0, time.UTC)
	lastDayOfMonth := firstDayOfNextMonth.Add(-time.Hour * 24).Day()
	return lastDayOfMonth
}

func constructDateTime(hourMinute string, days int) time.Time {
	now := time.Now()
	dateTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	hour, minute := parseHourMinute(hourMinute)
	fromHour, _ := strconv.Atoi(hour)
	fromMinute, _ := strconv.Atoi(minute)
	dateTime = dateTime.Add(time.Duration(fromHour+24*days)*time.Hour + time.Duration(fromMinute)*time.Minute)
	return dateTime
}

func isToBeforeFrom(from, to string) bool {
	parseHourFrom, _ := time.Parse(hourMinuteFormat, from)
	parsedHourTo, _ := time.Parse(hourMinuteFormat, to)
	return parsedHourTo.Before(parseHourFrom) || parsedHourTo.Equal(parseHourFrom)
}

func isTimeInBetween(timeCurrent, periodStart, periodEnd time.Time) bool {
	return (timeCurrent.After(periodStart) && timeCurrent.Before(periodEnd)) || timeCurrent.Equal(periodStart)
}

// Validate the date from and date to handle same day limits
func isDateFromBeforeTo(timeFrom, timeTo time.Time) bool {
	yearFrom, monthFrom, dayFrom := timeFrom.Date()
	yearTo, monthTo, dayTo := timeTo.Date()
	// Create new dates without the time component for comparison
	dateFrom := time.Date(yearFrom, monthFrom, dayFrom, 0, 0, 0, 0, time.UTC)
	dateTo := time.Date(yearTo, monthTo, dayTo, 0, 0, 0, 0, time.UTC)
	// Return true if dateFrom is strictly less than dateTo
	return dateFrom.Before(dateTo)
}
