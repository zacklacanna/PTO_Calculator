package tli

import (
	"fmt"
	"strings"
	"time"

	"pto_calculator/config"
	"pto_calculator/pto"
)

type tripField struct {
	Key     string
	Label   string
	Default string
	Hint    string
	Value   string
	Edited  bool
}

type holidayField struct {
	Key     string
	Label   string
	Default string
	Hint    string
	Value   string
	Edited  bool
}

func defaultTripFields() []tripField {
	startDate := time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	endDate := time.Now().AddDate(0, 0, 34).Format("2006-01-02")

	fields := []tripField{
		{Key: "name", Label: "Trip name", Default: "Summer trip", Hint: "Use a unique trip name."},
		{Key: "startDate", Label: "Start date", Default: startDate, Hint: "Use YYYY-MM-DD."},
		{Key: "endDate", Label: "End date", Default: endDate, Hint: "Use YYYY-MM-DD."},
	}

	for i := range fields {
		fields[i].Value = fields[i].Default
	}

	return fields
}

func defaultHolidayFields() []holidayField {
	date := time.Now().AddDate(0, 0, 7).Format("2006-01-02")

	fields := []holidayField{
		{Key: "name", Label: "Holiday name", Default: "Company holiday", Hint: "Use a short descriptive name."},
		{Key: "date", Label: "Holiday date", Default: date, Hint: "Use YYYY-MM-DD."},
	}

	for i := range fields {
		fields[i].Value = fields[i].Default
	}

	return fields
}

func buildTrip(fields []tripField) (config.Trip, error) {
	values := map[string]string{}
	for _, field := range fields {
		values[field.Key] = strings.TrimSpace(field.Value)
	}

	name := values["name"]
	if name == "" {
		return config.Trip{}, fmt.Errorf("trip name is required")
	}

	startDate, err := time.Parse("2006-01-02", values["startDate"])
	if err != nil {
		return config.Trip{}, fmt.Errorf("start date must use YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", values["endDate"])
	if err != nil {
		return config.Trip{}, fmt.Errorf("end date must use YYYY-MM-DD")
	}

	if endDate.Before(startDate) {
		return config.Trip{}, fmt.Errorf("end date cannot be before start date")
	}

	return config.Trip{
		Name:      name,
		StartDate: startDate,
		EndDate:   endDate,
	}, nil
}

func saveTrip(fields []tripField) (config.Trip, error) {
	trip, err := buildTrip(fields)
	if err != nil {
		return config.Trip{}, err
	}

	if err := pto.AddNewTrip(&trip); err != nil {
		return config.Trip{}, err
	}

	return trip, nil
}

func buildHoliday(fields []holidayField) (config.Holiday, error) {
	values := map[string]string{}
	for _, field := range fields {
		values[field.Key] = strings.TrimSpace(field.Value)
	}

	name := values["name"]
	if name == "" {
		return config.Holiday{}, fmt.Errorf("holiday name is required")
	}

	date, err := time.Parse("2006-01-02", values["date"])
	if err != nil {
		return config.Holiday{}, fmt.Errorf("holiday date must use YYYY-MM-DD")
	}

	return config.Holiday{
		Name: name,
		Date: date,
	}, nil
}

func saveHoliday(fields []holidayField) (config.Holiday, error) {
	holiday, err := buildHoliday(fields)
	if err != nil {
		return config.Holiday{}, err
	}

	if err := pto.AddHoliday(&holiday); err != nil {
		return config.Holiday{}, err
	}

	return holiday, nil
}
