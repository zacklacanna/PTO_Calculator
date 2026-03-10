package pto

import (
	"pto_calculator/config"
	"time"
)

// need to add logic with PTO rate increasing after a year
// for now:
// only allow trips planned within the calender year
// setup prompt in TUI to update PTO accrural rate at the start of new year and save to config

// Add logic to calculate if PTO gained during the trip
// This function does not factor in holidays, off fridays, just raw PTO & found trips
func CalculatePtoAtDate(startDate time.Time) {

	cfg := config.GetSettings()
	trips := config.GetSavedTrips()

	startingBalance := cfg.CurrentDays
	for _, trips := range trips {

	}

}
