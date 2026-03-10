package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CurrentDays   int     `yaml:"currentdays"`
	Rate          float64 `yaml:"rate"`
	Max           int     `yaml:"maxdays"`
	HasOffFridays bool    `yaml:"hasOffFridays"`
	NextFridayOff string  `yaml:"nextfridayoff"`
	UserName      string  `yaml:"user"`
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

const (
	ConfigPath = "config.yml"
	TripPath   = "trips.yml"
)

var settings Config
var savedTrips SavedTrips

func GetSettings() *Config {
	return &settings
}

func GetSavedTrips() *SavedTrips {
	return &savedTrips
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
