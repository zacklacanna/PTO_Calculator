package tli

import (
	"fmt"
	"strings"
	"time"

	"pto_calculator/pto"
)

type removeField struct {
	Label   string
	Default string
	Hint    string
	Value   string
	Edited  bool
}

func defaultRemoveTripField() removeField {
	return removeField{
		Label:   "Trip name",
		Default: "Summer trip",
		Hint:    "Enter the exact trip name to remove.",
		Value:   "Summer trip",
	}
}

func defaultRemoveHolidayField() removeField {
	date := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	return removeField{
		Label:   "Holiday date",
		Default: date,
		Hint:    "Enter the holiday date to remove in YYYY-MM-DD format.",
		Value:   date,
	}
}

func removeTripByName(field removeField) error {
	name := strings.TrimSpace(field.Value)
	if name == "" {
		return fmt.Errorf("trip name is required")
	}

	return pto.RemoveTrip(name)
}

func removeHolidayByDate(field removeField) error {
	date, err := time.Parse("2006-01-02", strings.TrimSpace(field.Value))
	if err != nil {
		return fmt.Errorf("holiday date must use YYYY-MM-DD")
	}

	return pto.RemoveHoliday(date)
}
