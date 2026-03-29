package tli

import (
	"fmt"
	"strings"
	"time"

	"pto_calculator/config"
	"pto_calculator/pto"
)

func renderTripList(trips []config.Trip, selected int) string {
	lines := []string{
		muted("Saved trips are ordered by start date."),
		"",
	}

	if len(trips) == 0 {
		lines = append(lines, muted("No trips saved yet."))
	} else {
		for i, trip := range trips {
			prefix := muted("  ")
			name := strong(trip.Name)
			if i == selected {
				prefix = accent("› ")
				name = highlight(trip.Name)
			}

			lines = append(lines, prefix+name)
			lines = append(lines, fmt.Sprintf("  %s to %s", trip.StartDate.Format("2006-01-02"), trip.EndDate.Format("2006-01-02")))
			if i != len(trips)-1 {
				lines = append(lines, "")
			}
		}
	}

	lines = append(lines, "")
	lines = append(lines, muted("Use up/down or j/k to move through trips."))
	lines = append(lines, muted("Press esc to return to the main menu."))
	return box("Trips", lines)
}

func renderTripDetail(trip config.Trip) string {
	if err := pto.LoadHolidays(); err != nil {
		return box("Trip Details", []string{
			danger("Error: " + err.Error()),
			"",
			muted("Press esc to return to the trip list."),
		})
	}

	cfg := config.GetSettings()
	holidays := config.GetHolidays()
	tripHours, err := pto.CalcultePtoOfTrip(&trip, cfg, holidays)
	if err != nil {
		return box("Trip Details", []string{
			danger("Error: " + err.Error()),
			"",
			muted("Press esc to return to the trip list."),
		})
	}

	startBalance, err := pto.CalculatePtoOnDate(trip.StartDate)
	startBalanceText := "Unavailable"
	endBalanceText := "Unavailable"
	if err == nil {
		ptoAfterTrip := startBalance
		ptoAtTripStart := startBalance + tripHours
		startBalanceText = fmt.Sprintf("%.1f hours", ptoAtTripStart)
		endBalanceText = fmt.Sprintf("%.1f hours", ptoAfterTrip)
	}

	startDate := startOfDay(trip.StartDate)
	endDate := startOfDay(trip.EndDate)
	calendarDays := int(endDate.Sub(startDate).Hours()/24) + 1
	containsOffFriday := tripHasOffFriday(trip, cfg)

	lines := []string{
		strong(trip.Name),
		"",
		fmt.Sprintf("%s %s", muted("Start date:"), trip.StartDate.Format("2006-01-02")),
		fmt.Sprintf("%s %s", muted("End date:"), trip.EndDate.Format("2006-01-02")),
		fmt.Sprintf("%s %d", muted("Calendar days:"), calendarDays),
		fmt.Sprintf("%s %.1f", muted("PTO hours used:"), tripHours),
		fmt.Sprintf("%s %d", muted("Daily PTO hours:"), cfg.DailyHours),
		fmt.Sprintf("%s %s", muted("PTO at trip start:"), startBalanceText),
		fmt.Sprintf("%s %s", muted("PTO after trip:"), endBalanceText),
		fmt.Sprintf("%s %t", muted("Includes off Friday:"), containsOffFriday),
		fmt.Sprintf("%s %s", muted("Starts on:"), trip.StartDate.Weekday().String()),
		fmt.Sprintf("%s %s", muted("Ends on:"), trip.EndDate.Weekday().String()),
		"",
		muted("Press esc to return to the trip list."),
	}

	return box("Trip Details", lines)
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func tripHasOffFriday(trip config.Trip, cfg *config.Config) bool {
	for d := startOfDay(trip.StartDate); !d.After(startOfDay(trip.EndDate)); d = d.AddDate(0, 0, 1) {
		if pto.IsOffFriday(d, cfg) {
			return true
		}
	}

	return false
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
