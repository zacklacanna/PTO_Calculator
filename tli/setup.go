package tli

import (
	tea "charm.land/bubbletea/v2"
)

func HandleStartup(*t tea.View) {
	// check if data is empty
	SetupData()

}

func CheckData() {
	//check if next friday off has past
}

func SetupData() {

	var t tea.View

	// ask user for name
	//
	// ask if user has off every other friday
	//yes: ask if it is this frida
	//save 0 or 1 for even or odd week, use to calculate for all future trips
	//no: save -1 and false for fridayoffs

	// check if date has

	// add functionality for saving holidays that are free days
}
