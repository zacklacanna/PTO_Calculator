package tli

import (
	"fmt"
	"strings"
	"time"

	"pto_calculator/config"
)

type setupField struct {
	Key     string
	Label   string
	Default string
	Hint    string
	Value   string
	Edited  bool
}

func defaultSetupFields() []setupField {
	fields := []setupField{
		{Key: "user", Label: "User name", Default: "Zack", Hint: "Name shown on the dashboard."},
		{Key: "initialBalance", Label: "Starting hours", Default: "40", Hint: "Current PTO balance in hours."},
		{Key: "rate", Label: "Accrual rate", Default: "3.52", Hint: "Hours earned each pay period."},
		{Key: "maxdays", Label: "Max Hours", Default: "240", Hint: "Maximum hours days allowed to sit in the bank."},
		{Key: "dailyHours", Label: "Daily hours", Default: "8", Hint: "Hours deducted for one PTO workday."},
		{Key: "firstDay", Label: "First work day", Default: "2026-02-09", Hint: "Use YYYY-MM-DD."},
		{Key: "hasOffFridays", Label: "Off Fridays", Default: "yes", Hint: "Type yes or no."},
		{Key: "whichFridayOff", Label: "Friday cycle", Default: "0", Hint: "Only used if you have off Fridays. Use 0 or 1."},
	}

	for i := range fields {
		fields[i].Value = fields[i].Default
	}

	return fields
}

func saveConfig(cfg config.Config) error {
	*config.GetSettings() = cfg
	return config.Save(config.ConfigPath, &cfg)
}

func buildConfig(fields []setupField) (config.Config, error) {
	values := map[string]string{}
	for _, field := range fields {
		values[field.Key] = strings.TrimSpace(field.Value)
	}

	userName := values["user"]
	if userName == "" {
		return config.Config{}, fmt.Errorf("user name is required")
	}

	initialBalance, err := parseFloat(values["initialBalance"], "starting hours")
	if err != nil {
		return config.Config{}, err
	}

	rate, err := parseFloat(values["rate"], "accrual rate")
	if err != nil {
		return config.Config{}, err
	}

	maxDays, err := parseInt(values["maxdays"], "max days")
	if err != nil {
		return config.Config{}, err
	}

	dailyHours, err := parseInt(values["dailyHours"], "daily hours")
	if err != nil {
		return config.Config{}, err
	}

	firstDay, err := time.Parse("2006-01-02", values["firstDay"])
	if err != nil {
		return config.Config{}, fmt.Errorf("first work day must use YYYY-MM-DD")
	}

	hasOffFridays, err := parseBool(values["hasOffFridays"])
	if err != nil {
		return config.Config{}, err
	}

	whichFridayOff := -1
	if hasOffFridays {
		whichFridayOff, err = parseInt(values["whichFridayOff"], "friday cycle")
		if err != nil {
			return config.Config{}, err
		}
		if whichFridayOff != 0 && whichFridayOff != 1 {
			return config.Config{}, fmt.Errorf("friday cycle must be 0 or 1")
		}
	}

	return config.Config{
		InitialBalance: initialBalance,
		Rate:           rate,
		Max:            maxDays,
		HasOffFridays:  hasOffFridays,
		WhichFridayOff: whichFridayOff,
		UserName:       userName,
		FirstDay:       firstDay,
		DailyHours:     dailyHours,
	}, nil
}

func shouldShowFridayCycle(fields []setupField) bool {
	for _, field := range fields {
		if field.Key == "hasOffFridays" {
			enabled, err := parseBool(strings.TrimSpace(field.Value))
			return err == nil && enabled
		}
	}

	return false
}

func visibleFieldIndexes(fields []setupField) []int {
	showFridayCycle := shouldShowFridayCycle(fields)
	indexes := make([]int, 0, len(fields))
	for i, field := range fields {
		if field.Key == "whichFridayOff" && !showFridayCycle {
			continue
		}
		indexes = append(indexes, i)
	}

	return indexes
}

func bannerArt() string {
	return strings.TrimSpace(`
██████╗ ████████╗ ██████╗
██╔══██╗╚══██╔══╝██╔═══██╗
██████╔╝   ██║   ██║   ██║
██╔═══╝    ██║   ██║   ██║
██║        ██║   ╚██████╔╝
╚═╝        ╚═╝    ╚═════╝
`)
}

func welcomeCopy() string {
	return strings.Join([]string{
		"Plan time off without guessing your balance.",
		"",
		"No config was found, so this first screen will build one.",
		"Press enter to start your PTO setup.",
		"",
		"Controls: enter to continue, q to quit",
	}, "\n")
}
