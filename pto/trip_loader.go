package pto

import (
	"os"
	"time"

	"pto_calculator/config"
)

func LoadTrips() error {
	info, err := os.Stat(config.TripPath)
	if err != nil {
		if os.IsNotExist(err) {
			config.GetSavedTrips().Trips = map[string]config.Trip{}
			return nil
		}
		return err
	}

	if info.Size() == 0 {
		config.GetSavedTrips().Trips = map[string]config.Trip{}
		return nil
	}

	var trips config.SavedTrips
	if err := config.Load(config.TripPath, &trips); err != nil {
		return err
	}
	if trips.Trips == nil {
		trips.Trips = map[string]config.Trip{}
	}

	*config.GetSavedTrips() = trips
	return nil
}

func LoadHolidays() error {
	info, err := os.Stat(config.HolidaysPath)
	if err != nil {
		if os.IsNotExist(err) {
			config.GetHolidays().Holidays = map[time.Time]config.Holiday{}
			return nil
		}
		return err
	}

	if info.Size() == 0 {
		config.GetHolidays().Holidays = map[time.Time]config.Holiday{}
		return nil
	}

	var holidays config.SavedHolidays
	if err := config.Load(config.HolidaysPath, &holidays); err != nil {
		return err
	}
	if holidays.Holidays == nil {
		holidays.Holidays = map[time.Time]config.Holiday{}
	}

	*config.GetHolidays() = holidays
	return nil
}
