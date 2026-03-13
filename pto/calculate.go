package pto

import (
	"fmt"
	"pto_calculator/config"
	"time"
)

// need to add logic with PTO rate increasing after a year
// for now:
// only allow trips planned within the calender year
// setup prompt in TUI to update PTO accrural rate at the start of new year and save to config

// Add logic to calculate if PTO gained during the trip
// This function does not factor in holidays, off fridays, just raw PTO & found trips
func CalculatePtoAtDate(startDate time.Time) (float64, error) {

	cfg := config.GetSettings()
	trips := config.GetSavedTrips()
	holidays := config.GetHolidays()

	// Running balance
	runningBalance := cfg.InitialBalance

	// Calculate PTO gained from first day of job until now
	if !startDate.Before(cfg.FirstDay) {
		daysSinceStart := int(startDate.Sub(cfg.FirstDay).Hours() / 24)
		twoWeekBlocks := daysSinceStart / 14
		runningBalance += float64(twoWeekBlocks) * cfg.Rate
	}

	// Iterate through each trip and subtract from logic
	for _, trip := range trips.Trips {

		tripHours := 0.0
		//Iterate through days of trip and subtract if it is not
		// - Weekend
		// - Off friday
		// - Holiday

		for d := normalizeDate(trip.StartDate); !d.After(normalizeDate(trip.EndDate)); d = d.AddDate(0, 0, 1) {

			if IsOffFriday(d, cfg) || IsHoliday(d, holidays) || !IsWeekday(d) {
				continue
			}
			tripHours += float64(cfg.DailyHours)
		}

		if runningBalance-tripHours < 0 {
			// handle case where trips are exceeding PTO
			return -1, fmt.Errorf("Trips exceed allowed PTO")
		}

		runningBalance -= tripHours
	}

	return runningBalance, nil

}

func normalizeDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func IsHoliday(date time.Time, holidays *config.SavedHolidays) bool {

	date = normalizeDate(date)
	_, exists := holidays.Holidays[date]
	return exists
}

func IsWeekday(date time.Time) bool {

	weekday := date.Weekday()
	return weekday != time.Friday && weekday != time.Saturday && weekday != time.Sunday

}

func IsOffFriday(date time.Time, cfg *config.Config) bool {

	if !cfg.HasOffFridays {
		return false
	}

	if date.Weekday() != time.Friday {
		return false
	}

	startOfYear := time.Date(date.Year(), time.January, 1, 0, 0, 0, 0, date.Location())

	daysSinceYearStart := int(date.Sub(startOfYear).Hours() / 24)
	weekIndex := daysSinceYearStart / 7

	// 0 = odd weeks, 1 = even weeks
	if cfg.WhichFridayOff == 0 {
		return weekIndex%2 == 1
	}

	return weekIndex%2 == 0
}
