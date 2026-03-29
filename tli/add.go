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

type tripPreview struct {
	Trip                config.Trip
	StartBalance        float64
	EndBalance          float64
	PTOHoursUsed        float64
	CalendarDays        int
	StartWeekday        string
	EndWeekday          string
	OverlapTripName     string
	ValidationMessage   string
	ValidationIsFailure bool
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

func adjustTripDateField(fields []tripField, index int, days int) {
	if index < 0 || index >= len(fields) {
		return
	}
	if fields[index].Key != "startDate" && fields[index].Key != "endDate" {
		return
	}

	currentDate, err := time.Parse("2006-01-02", strings.TrimSpace(fields[index].Value))
	if err != nil {
		return
	}

	fields[index].Value = currentDate.AddDate(0, 0, days).Format("2006-01-02")
	fields[index].Edited = true
}

func buildTripPreview(fields []tripField) (*tripPreview, error) {
	trip, err := buildTrip(fields)
	if err != nil {
		return nil, err
	}

	if err := pto.LoadHolidays(); err != nil {
		return nil, err
	}

	cfg := config.GetSettings()
	holidays := config.GetHolidays()

	tripHours, err := pto.CalcultePtoOfTrip(&trip, cfg, holidays)
	if err != nil {
		return nil, err
	}

	startBalance, err := pto.CalculatePtoOnDate(trip.StartDate)
	if err != nil {
		return nil, err
	}

	startDate := time.Date(trip.StartDate.Year(), trip.StartDate.Month(), trip.StartDate.Day(), 0, 0, 0, 0, trip.StartDate.Location())
	endDate := time.Date(trip.EndDate.Year(), trip.EndDate.Month(), trip.EndDate.Day(), 0, 0, 0, 0, trip.EndDate.Location())

	preview := &tripPreview{
		Trip:         trip,
		StartBalance: startBalance,
		EndBalance:   startBalance - tripHours,
		PTOHoursUsed: tripHours,
		CalendarDays: int(endDate.Sub(startDate).Hours()/24) + 1,
		StartWeekday: trip.StartDate.Weekday().String(),
		EndWeekday:   trip.EndDate.Weekday().String(),
	}

	overlapTrip, overlaps, err := pto.FindOverlappingTrip(&trip)
	if err != nil {
		return nil, err
	}
	if overlaps {
		preview.OverlapTripName = overlapTrip.Name
		preview.ValidationMessage = "Trip overlaps with " + overlapTrip.Name + "."
		preview.ValidationIsFailure = true
		return preview, nil
	}

	if err := pto.ValidateProjectedTrips(&trip); err != nil {
		preview.ValidationMessage = err.Error()
		preview.ValidationIsFailure = true
		return preview, nil
	}

	if preview.EndBalance < 0 {
		preview.ValidationMessage = "Trip exceeds the projected PTO available at the start date."
		preview.ValidationIsFailure = true
	} else {
		preview.ValidationMessage = "Trip fits the projected PTO balance."
	}

	return preview, nil
}

func renderTripPreviewBox(fields []tripField, width int) string {
	preview, err := buildTripPreview(fields)
	if err != nil {
		return fitBox("Trip Preview", []string{
			muted("Live trip estimate"),
			"",
			danger("Preview unavailable: " + err.Error()),
			"",
			muted("Tip: use left/right to nudge the selected date by one day."),
			muted("Use page up/page down to move a week at a time."),
		}, width)
	}

	statusLine := success(preview.ValidationMessage)
	if preview.ValidationIsFailure {
		statusLine = danger(preview.ValidationMessage)
	}

	return fitBox("Trip Preview", []string{
		muted("Live trip estimate"),
		"",
		fmt.Sprintf("%s %s", muted("Start:"), preview.Trip.StartDate.Format("2006-01-02")+" ("+preview.StartWeekday+")"),
		fmt.Sprintf("%s %s", muted("End:"), preview.Trip.EndDate.Format("2006-01-02")+" ("+preview.EndWeekday+")"),
		fmt.Sprintf("%s %d", muted("Calendar days:"), preview.CalendarDays),
		fmt.Sprintf("%s %.1f", muted("PTO hours used:"), preview.PTOHoursUsed),
		fmt.Sprintf("%s %.1f hours", muted("PTO on start date:"), preview.StartBalance),
		fmt.Sprintf("%s %.1f hours", muted("PTO after trip:"), preview.EndBalance),
		func() string {
			if preview.OverlapTripName == "" {
				return fmt.Sprintf("%s %s", muted("Overlapping trip:"), muted("none"))
			}
			return fmt.Sprintf("%s %s", muted("Overlapping trip:"), danger(preview.OverlapTripName))
		}(),
		"",
		statusLine,
		"",
		muted("Tip: use left/right to nudge the selected date by one day."),
		muted("Use page up/page down to move a week at a time."),
	}, width)
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
