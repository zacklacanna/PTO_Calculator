package pto

import (
	"time"
)

type Trip struct {
	name      string    `json:"name"`
	startDate time.Time `json:"startDate"`
	endDate   time.Time `json:"startDate"`
}

type Holiday struct {
	name string    `json:"name"`
	date time.Time `json:"date"`
}

func AddNewTrip(tripReq *Trip) error {

	//pull current config settings

	//parse trip settings
	// name doesnt already exist
	// no overlapping dates with another saved trip

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

func RemoveTrip(name string) error {

	return nil
}

func ListTrips() []Trip {

}
