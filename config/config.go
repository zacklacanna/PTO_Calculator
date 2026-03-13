package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	InitialBalance float64   `yaml:"initialBalance"`
	Rate           float64   `yaml:"rate"`
	Max            int       `yaml:"maxdays"`
	HasOffFridays  bool      `yaml:"hasOffFridays"`
	WhichFridayOff int       `yaml:"whichFridayOff"`
	UserName       string    `yaml:"user"`
	FirstDay       time.Time `yaml:"firstDay"`
	DailyHours     int       `yaml:"dailyHours"`
}

type Trip struct {
	Name      string    `yaml:"name"`
	StartDate time.Time `yaml:"startDate"`
	EndDate   time.Time `yaml:"endDate"`
}

type Holiday struct {
	Name string    `yaml:"name"`
	Date time.Time `yaml:"date"`
}

type SavedTrips struct {
	Trips map[string]Trip `yaml:"trips"`
}

type SavedHolidays struct {
	Holidays map[time.Time]Holiday `yaml:"holidays"`
}

const (
	ConfigPath   = "config.yml"
	TripPath     = "trips.yml"
	HolidaysPath = "holidays.yml"
)

var settings Config
var savedTrips SavedTrips
var savedHolidays SavedHolidays

func GetSettings() *Config {
	return &settings
}

func GetSavedTrips() *SavedTrips {
	return &savedTrips
}

func GetHolidays() *SavedHolidays {
	return &savedHolidays
}

func LoadDefaults() {

	// generate all the inital files

}

func Load(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Could not find config file!")
	}

	return yaml.Unmarshal(data, out)
}

func Save(path string, out any) error {

	data, err := yaml.Marshal(out)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, 0o644)
	if err != nil {
		return err
	}

	return nil
}
