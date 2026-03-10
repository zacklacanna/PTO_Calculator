package tli

import (
	tea "charm.land/bubbletea/v2"
)

type startupModel struct{}

func InitTUI() {
	m := startupModel{}
	tea.NewProgram(m).Run()
}

// options for TLI
// 1.) Add new trip
// 2.) List saved trips
// 3.) Display Current PTO data & Total accrural / rate
// 4.) Change config settings
// 5.) Change holdiday dates
// 6.) Quit
