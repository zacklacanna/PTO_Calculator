package pto

import (
	"fmt"
	"math"
	"pto_calculator/config"
	"time"
)

// need to add logic with PTO rate increasing after a year
// for now:
// only allow trips planned within the calender year
// setup prompt in TUI to update PTO accrural rate at the start of new year and save to config

// Add logic to calculate if PTO gained during the trip
// This function does not factor in holidays, off fridays, just raw PTO & found trips
func CalculatePtoAtDate(tripReq *config.Trip) (float64, error) {
	if err := LoadTrips(); err != nil {
		return -1, err
	}
	if err := LoadHolidays(); err != nil {
		return -1, err
	}

	cfg := config.GetSettings()
	trips := config.GetSavedTrips()
	holidays := config.GetHolidays()
	startDate := normalizeDate(tripReq.StartDate)

	// Running balance
	runningBalance := cfg.InitialBalance

	// Calculate PTO gained from first day of job until now
	if !tripReq.StartDate.Before(cfg.FirstDay) {
		daysSinceStart := int(tripReq.StartDate.Sub(cfg.FirstDay).Hours() / 24)
		twoWeekBlocks := daysSinceStart / 14
		runningBalance += float64(twoWeekBlocks) * cfg.Rate
	}

	// Iterate through each trip and subtract from logic
	for _, trip := range trips.Trips {
		if normalizeDate(trip.StartDate).After(startDate) {
			continue
		}

		tripHours, err := CalcultePtoOfTrip(&trip, cfg, holidays)
		if err != nil {
			return -1, err
		}

		if runningBalance-tripHours < 0 {
			return -1, fmt.Errorf("trip %s would exceed the PTO balance", trip.Name)
		}

		runningBalance -= tripHours
	}

	newTripCheck, err := CalcultePtoOfTrip(tripReq, cfg, holidays)
	if err != nil {
		return -1, err
	}

	if runningBalance-newTripCheck < 0 {
		return -1, fmt.Errorf("Could not add new trip, PTO would be exceeded")
	}

	// Ensure we dont surpass the maximum threshold
	return math.Max(float64(cfg.Max), runningBalance), nil
}

// Returns PTO usage of trip given
func CalcultePtoOfTrip(trip *config.Trip,
	cfg *config.Config,
	holidays *config.SavedHolidays) (float64, error) {

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

	return tripHours, nil

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
	return weekday != time.Saturday && weekday != time.Sunday

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
