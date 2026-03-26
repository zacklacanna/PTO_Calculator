package pto

import (
	"fmt"
	"pto_calculator/config"
)

func AddNewTrip(tripReq *config.Trip) error {

	//pull current config settings
	var config = config.GetSettings()
	if config.InitialBalance < 0 {
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

	foundTrip, ok := currentTrips.Trips[trip.Name]
	if !ok {
		return config.Trip{}, fmt.Errorf("Could not find Trip %s", trip.Name)
	}

	return foundTrip, nil
}

func checkValidNewTrip(tripReq *config.Trip) (config.Trip, error) {

	_, err := GetTrip(tripReq)
	if err == nil {
		return config.Trip{}, fmt.Errorf("A trip already exists with the name: %s", tripReq.Name)
	}

	foundOverlapTrip, ok := hasOverlappingTrip(tripReq, config.GetSavedTrips())
	if !ok {
		return config.Trip{}, fmt.Errorf("Could not create trip as it overlaps with %s", foundOverlapTrip.Name)
	}

	runningBalance, err := CalculatePtoAtDate(tripReq)
	if err != nil {
		return config.Trip{}, fmt.Errorf("Not enough PTO at start date to book this trip!")
	}

	if runningBalance >= 0 {
		return *tripReq, nil
	} else {
		return config.Trip{}, fmt.Errorf("Could not create trip as would exceed PTO Balance!")
	}
}

func hasOverlappingTrip(newTrip *config.Trip, savedTrips *config.SavedTrips) (config.Trip, bool) {
	for _, trip := range savedTrips.Trips {
		if !newTrip.EndDate.Before(trip.StartDate) && !trip.EndDate.Before(newTrip.StartDate) {
			return trip, false
		}
	}
	return config.Trip{}, true
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
