package bean

import "testing"

func TestIsValidDateInput_ValidDates(t *testing.T) {
	validDates := []string{
		"Mon Jan 2 15:04:05 2006",
		"Mon Jan 2 15:04:05 MST 2006",
		"Mon Jan 02 15:04:05 -0700 2006",
		"2006-01-02",
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05-07:00",
		"2007-01-02T15+0700",
		"2007-01-02T15:04+0700",
		"2007-01-02T15:04:05+0700",
		"2007-01-02T15:04:05.999999999+0700",
	}
	for _, date := range validDates {
		if !isValidDateInput(date) {
			t.Errorf("expected date %s to be valid", date)
		}
	}
}

func TestIsValidDateInput_InvalidDates(t *testing.T) {
	invalidDates := []string{
		"invalid date",
		"2021-13-01",
		"2021-12-32",
		"2021-12-01 25:00",
		"2021-12-01 23:60",
		"2021-12-01T25:00Z0700",
		"2021-12-01T23:60Z0700",
	}
	for _, date := range invalidDates {
		if isValidDateInput(date) {
			t.Errorf("expected date %s to be invalid", date)
		}
	}
}
