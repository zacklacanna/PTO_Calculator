package pto

import (
	"fmt"
	"pto_calculator/config"
)

func AddNewTrip(tripReq *config.Trip) error {

	//pull current config settings
	var config = config.GetSettings()
	if config.CurrentDays < 0 {
		// Tea flow for setting default settings
		return fmt.Errorf("Invalid config files")
	}

	//parse trip settings
	// name doesnt already exist
	// no overlapping dates with another saved trip

	_, err := checkValidNewTrip(tripReq)
	if err != nil {
		return err
	}

	// calculate Pto at time of start of trip

	// calculate total days that will need to be used between start & end
	// filter out:
	// - weekends
	// - off fridays
	// - holidays

	// if can approve return trip and save to saved trips
	// false return error

	return nil

}

func GetTrip(trip *config.Trip) (config.Trip, error) {

	currentTrips := config.GetSavedTrips()

	trip, ok := currentTrips.Trips[name]
	if !ok {
		return config.Trip{}, fmt.Errorf("Could not find Trip %s", name)
	}

	return trip, nil
}

func checkValidNewTrip(tripReq *config.Trip) (config.Trip, error) {

	var foundTrip config.Trip
	currentTrips := config.GetSavedTrips()

	_, err := GetTrip(tripReq)
	if err == nil {
		return config.Trip{}, fmt.Errorf("A trip already exists with the name: %s", tripReq.Name)
	}

	for _, trip := range currentTrips.Trips {

		// Check if trip start overlapps
		if tripReq.StartDate.Before(trip.EndDate) && trip.StartDate.Before(tripReq.EndDate) {
			return config.Trip{}, fmt.Errorf("This trip would overlap with %s", trip.Name)
		}
	}

	return foundTrip, nil

}

func RemoveTrip(name string) error {

	trips := config.GetSavedTrips().Trips

	if trips == nil {
		return fmt.Errorf("no saved trips found")
	}

	if _, ok := trips[name]; !ok {
		return fmt.Errorf("trip %s not found", name)
	}

	delete(trips, name)

	if err := config.Save(config.TripPath, trips); err != nil {
		return err
	}

	return nil
}
