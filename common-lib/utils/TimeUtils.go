package utils

import (
	"fmt"
	"time"
)

type TimeRangeRequest struct {
	From       *time.Time   `json:"from" schema:"from"`
	To         *time.Time   `json:"to" schema:"to"`
	TimeWindow *TimeWindows `json:"timeWindow" schema:"timeWindow" validate:"omitempty,oneof=today yesterday week month quarter lastWeek lastMonth"`
}

func NewTimeRangeRequest(from *time.Time, to *time.Time) *TimeRangeRequest {
	return &TimeRangeRequest{
		From: from,
		To:   to,
	}
}

func NewTimeWindowRequest(timeWindow TimeWindows) *TimeRangeRequest {
	return &TimeRangeRequest{
		TimeWindow: &timeWindow,
	}
}

// TimeWindows is a string type that represents different time windows
type TimeWindows string

func (timeRange TimeWindows) String() string {
	return string(timeRange)
}

// Define constants for different time windows
const (
	Today     TimeWindows = "today"
	Yesterday TimeWindows = "yesterday"
	Week      TimeWindows = "week"
	Month     TimeWindows = "month"
	Quarter   TimeWindows = "quarter"
	LastWeek  TimeWindows = "lastWeek"
	LastMonth TimeWindows = "lastMonth"
)

func (timeRange *TimeRangeRequest) ParseTimeRange() (*TimeRangeRequest, error) {
	if timeRange == nil {
		return NewTimeRangeRequest(&time.Time{}, &time.Time{}), fmt.Errorf("invalid time range request. either from/to or timeWindow must be provided")
	}
	now := time.Now()
	// If timeWindow is provided, it takes preference over from/to
	if timeRange.TimeWindow != nil {
		switch *timeRange.TimeWindow {
		case Today:
			start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			return NewTimeRangeRequest(&start, &now), nil
		case Yesterday:
			start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Add(-24 * time.Hour)
			end := start.Add(24 * time.Hour)
			return NewTimeRangeRequest(&start, &end), nil
		case Week:
			// Current week (Monday to Sunday)
			weekday := int(now.Weekday())
			if weekday == 0 { // Sunday
				weekday = 7
			}
			start := now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
			return NewTimeRangeRequest(&start, &now), nil
		case Month:
			start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			return NewTimeRangeRequest(&start, &now), nil
		case Quarter:
			quarter := ((int(now.Month()) - 1) / 3) + 1
			quarterStart := time.Month((quarter-1)*3 + 1)
			start := time.Date(now.Year(), quarterStart, 1, 0, 0, 0, 0, now.Location())
			return NewTimeRangeRequest(&start, &now), nil
		case LastWeek:
			weekday := int(now.Weekday())
			if weekday == 0 { // Sunday
				weekday = 7
			}
			thisWeekStart := now.AddDate(0, 0, -(weekday - 1)).Truncate(24 * time.Hour)
			lastWeekStart := thisWeekStart.AddDate(0, 0, -7)
			lastWeekEnd := thisWeekStart.Add(-time.Second)
			return NewTimeRangeRequest(&lastWeekStart, &lastWeekEnd), nil
		case LastMonth:
			thisMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			lastMonthStart := thisMonthStart.AddDate(0, -1, 0)
			lastMonthEnd := thisMonthStart.Add(-time.Second)
			return NewTimeRangeRequest(&lastMonthStart, &lastMonthEnd), nil
		default:
			return NewTimeRangeRequest(&time.Time{}, &time.Time{}), fmt.Errorf("unsupported time window: %q", *timeRange.TimeWindow)
		}
	}

	// Use from/to dates if provided
	if timeRange.From != nil && timeRange.To != nil {
		if timeRange.From.After(*timeRange.To) {
			return NewTimeRangeRequest(&time.Time{}, &time.Time{}), fmt.Errorf("from date cannot be after to date")
		}
		return NewTimeRangeRequest(timeRange.From, timeRange.To), nil
	} else {
		return NewTimeRangeRequest(&time.Time{}, &time.Time{}), fmt.Errorf("from and to dates are required if time window is not provided")
	}
}
