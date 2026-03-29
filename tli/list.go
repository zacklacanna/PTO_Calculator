package tli

import (
	"fmt"
	"strings"

	"pto_calculator/config"
)

func renderTripList(trips []config.Trip) string {
	lines := []string{
		muted("Saved trips are ordered by start date."),
		"",
	}

	if len(trips) == 0 {
		lines = append(lines, muted("No trips saved yet."))
	} else {
		for i, trip := range trips {
			lines = append(lines, strong(trip.Name))
			lines = append(lines, fmt.Sprintf("  %s to %s", trip.StartDate.Format("2006-01-02"), trip.EndDate.Format("2006-01-02")))
			if i != len(trips)-1 {
				lines = append(lines, "")
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, muted("Press esc to return to the main menu."))
	return box("Trips", lines)
}

func renderHolidayList(holidays []config.Holiday) string {
	lines := []string{
		muted("Holiday days do not consume PTO."),
		"",
	}

	if len(holidays) == 0 {
		lines = append(lines, muted("No holidays saved yet."))
	} else {
		for i, holiday := range holidays {
			lines = append(lines, strong(holiday.Name))
			lines = append(lines, "  "+holiday.Date.Format("2006-01-02"))
			if i != len(holidays)-1 {
				lines = append(lines, "")
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, muted("Press esc to return to the main menu."))
	return box("Holidays", lines)
}

func renderSingleFieldForm(title string, intro string, field removeField, active bool, errText string, footer string) string {
	prefix := muted("  ")
	label := muted(field.Label)
	value := field.Value
	if active {
		prefix = accent("› ")
		label = highlight(field.Label)
		value = strong(value + " ")
	}

	lines := []string{
		muted(intro),
		"",
		fmt.Sprintf("%s%-12s %s", prefix, label, value),
		muted(field.Hint),
	}

	if strings.TrimSpace(errText) != "" {
		lines = append(lines, "")
		lines = append(lines, danger("Error: "+errText))
	}

	lines = append(lines, "")
	lines = append(lines, muted(footer))

	return box(title, lines)
}
