package config

import (
	"fmt"
	"os"
	"pto_calculator/pto"

	"gopkg.in/yaml.v3"
)

type Config struct {
	CurrentDays   int     `json:"currentdays"`
	Rate          float64 `json:"rate"`
	Max           int     `json:"maxdays`
	HasOffFridays bool    `json:"hasOffFridays"`
	NextFridayOff string  `json:"nextfridayoff"`
	UserName      string  `json:"user"`
}

type SavedTrips struct {
	trips []pto.TripRequest `json:"trips"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Could not find config file!")
	}

	var cfg Config
	yaml.Unmarshal(data, &cfg)
	return &cfg, nil
}

func Save(cfg *Config, path string) {

	data, err := yaml.Marshal(cfg)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = os.WriteFile(path, data, 0o644)

}
