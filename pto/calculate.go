package pto

import (
	"fmt"
	"math"
	"pto_calculator/config"
	"slices"
	"time"
)

// need to add logic with PTO rate increasing after a year
// for now:
// only allow trips planned within the calender year
// setup prompt in TUI to update PTO accrural rate at the start of new year and save to config

// Add logic to calculate if PTO gained during the trip
// This function does not factor in holidays, off fridays, just raw PTO & found trips
func CalculatePtoAtDate(tripReq *config.Trip) (float64, error) {
	runningBalance, err := CalculatePtoOnDate(tripReq.StartDate)
	if err != nil {
		return -1, err
	}

	cfg := config.GetSettings()
	holidays := config.GetHolidays()

	newTripCheck, err := CalcultePtoOfTrip(tripReq, cfg, holidays)
	if err != nil {
		return -1, err
	}

	if runningBalance-newTripCheck < 0 {
		return -1, fmt.Errorf("Could not add new trip, PTO would be exceeded")
	}

	return runningBalance, nil
}

func CalculatePtoOnDate(date time.Time) (float64, error) {
	if err := LoadTrips(); err != nil {
		return -1, err
	}
	if err := LoadHolidays(); err != nil {
		return -1, err
	}

	cfg := config.GetSettings()
	trips := config.GetSavedTrips()
	holidays := config.GetHolidays()
	targetDate := normalizeDate(date)

	sortedTrips := make([]config.Trip, 0, len(trips.Trips))
	for _, trip := range trips.Trips {
		sortedTrips = append(sortedTrips, trip)
	}
	slices.SortFunc(sortedTrips, func(a config.Trip, b config.Trip) int {
		return a.StartDate.Compare(b.StartDate)
	})

	return simulateProjectedBalance(sortedTrips, cfg, holidays, &targetDate)
}

func calculateAccruedPTO(targetDate time.Time, cfg *config.Config) float64 {
	firstAccrualDate, ok := nextAccrualDate(cfg.FirstDay, cfg)
	if !ok || targetDate.Before(firstAccrualDate) {
		return 0
	}

	accruals := 0
	for accrualDate := firstAccrualDate; !accrualDate.After(targetDate); accrualDate = accrualDate.AddDate(0, 0, 14) {
		accruals++
	}

	return float64(accruals) * cfg.Rate
}

func ValidateProjectedTrips(candidate *config.Trip) error {
	if err := LoadTrips(); err != nil {
		return err
	}
	if err := LoadHolidays(); err != nil {
		return err
	}

	cfg := config.GetSettings()
	holidays := config.GetHolidays()

	trips := make([]config.Trip, 0, len(config.GetSavedTrips().Trips)+1)
	for _, trip := range config.GetSavedTrips().Trips {
		trips = append(trips, trip)
	}
	trips = append(trips, *candidate)
	slices.SortFunc(trips, func(a config.Trip, b config.Trip) int {
		return a.StartDate.Compare(b.StartDate)
	})

	_, err := simulateProjectedBalance(trips, cfg, holidays, nil)
	return err
}

func simulateProjectedBalance(
	trips []config.Trip,
	cfg *config.Config,
	holidays *config.SavedHolidays,
	targetDate *time.Time,
) (float64, error) {
	runningBalance := math.Min(float64(cfg.Max), cfg.InitialBalance)
	nextAccrualDate, hasAccrual := nextAccrualDate(cfg.FirstDay, cfg)

	applyAccruals := func(until time.Time) {
		for hasAccrual && !nextAccrualDate.After(until) {
			runningBalance = math.Min(float64(cfg.Max), runningBalance+cfg.Rate)
			nextAccrualDate = nextAccrualDate.AddDate(0, 0, 14)
		}
	}

	for _, trip := range trips {
		tripStart := normalizeDate(trip.StartDate)
		if targetDate != nil && tripStart.After(*targetDate) {
			break
		}

		applyAccruals(tripStart)

		tripHours, err := CalcultePtoOfTrip(&trip, cfg, holidays)
		if err != nil {
			return -1, err
		}

		if runningBalance-tripHours < 0 {
			return -1, fmt.Errorf("trip %s would exceed the PTO balance", trip.Name)
		}

		runningBalance -= tripHours
	}

	if targetDate != nil {
		applyAccruals(*targetDate)
	}

	return runningBalance, nil
}

func nextAccrualDate(firstDay time.Time, cfg *config.Config) (time.Time, bool) {
	start := normalizeDate(firstDay)
	for d := start; !d.After(start.AddDate(0, 0, 14)); d = d.AddDate(0, 0, 1) {
		if isAccrualFriday(d, cfg) {
			return d, true
		}
	}

	return time.Time{}, false
}

func isAccrualFriday(date time.Time, cfg *config.Config) bool {
	if date.Weekday() != time.Friday {
		return false
	}

	if !cfg.HasOffFridays {
		// Without an alternating Friday schedule, default to the first Friday
		// after start date and then every 14 days from there.
		return true
	}

	// Accrual follows the same alternating Friday set selected in config.
	return IsOffFriday(date, cfg)
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
