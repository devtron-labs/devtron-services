package scheduler

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

func (tr TimeRange) ValidateTimeRange() error {
	if !slices.Contains(AllowedFrequencies, tr.Frequency) {
		return errors.New("invalid Frequency type")
	}
	if tr.Frequency != FIXED {
		colonCount := strings.Count(tr.HourMinuteFrom, ":")
		if colonCount != 1 {
			return errors.New("invalid format: must contain exactly one colon")
		}
		if tr.HourMinuteFrom == tr.HourMinuteTo {
			return errors.New("HourMinuteFrom must not be equal to HourMinuteTo")
		}
	}
	switch tr.Frequency {
	case FIXED:
		if tr.TimeFrom.IsZero() || tr.TimeTo.IsZero() {
			return errors.New("TimeFrom and TimeTo must be present for FIXED frequency")
		}
		if tr.TimeFrom.After(tr.TimeTo) {
			return errors.New("TimeFrom must be less than TimeTo for FIXED frequency")
		}
		if tr.TimeFrom == tr.TimeTo {
			return errors.New("TimeFrom must not be equal to TimeTo for FIXED frequency")
		}
	case DAILY:
		if tr.HourMinuteFrom == "" || tr.HourMinuteTo == "" {
			return errors.New("HourMinuteFrom and HourMinuteTo must be present for DAILY frequency")
		}
	case WEEKLY:
		if len(tr.Weekdays) == 0 || tr.HourMinuteFrom == "" || tr.HourMinuteTo == "" {
			return errors.New("weekdays, HourMinuteFrom, and HourMinuteTo must be present for WEEKLY frequency")
		}
	case WEEKLY_RANGE:
		if tr.WeekdayFrom == 0 || tr.HourMinuteFrom == "" || tr.HourMinuteTo == "" {
			return errors.New("WeekdayFrom, HourMinuteFrom, and HourMinuteTo must be present for WEEKLY_RANGE frequency")
		}
	case MONTHLY:
		if tr.DayFrom == 0 || tr.DayTo == 0 || tr.HourMinuteFrom == "" || tr.HourMinuteTo == "" {
			return errors.New("DayFrom, DayTo, HourMinuteFrom, and HourMinuteTo must be present for MONTHLY frequency")
		}
	default:
		return fmt.Errorf("unknown frequency: %s", tr.Frequency)
	}
	return nil
}
