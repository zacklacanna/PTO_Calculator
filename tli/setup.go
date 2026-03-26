package tli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"pto_calculator/config"
)

type setupField struct {
	Key         string
	Label       string
	Placeholder string
	Hint        string
	Value       string
}

func defaultSetupFields() []setupField {
	return []setupField{
		{Key: "user", Label: "User name", Placeholder: "Zack", Hint: "Name shown on the dashboard."},
		{Key: "initialBalance", Label: "Starting hours", Placeholder: "120", Hint: "Current PTO balance in hours."},
		{Key: "rate", Label: "Accrual rate", Placeholder: "4.62", Hint: "Hours earned each pay period."},
		{Key: "maxdays", Label: "Max days", Placeholder: "25", Hint: "Maximum PTO days allowed to sit in the bank."},
		{Key: "dailyHours", Label: "Daily hours", Placeholder: "8", Hint: "Hours deducted for one PTO workday."},
		{Key: "firstDay", Label: "First work day", Placeholder: "2026-01-05", Hint: "Use YYYY-MM-DD."},
		{Key: "hasOffFridays", Label: "Off Fridays", Placeholder: "yes", Hint: "Type yes or no."},
		{Key: "whichFridayOff", Label: "Friday cycle", Placeholder: "0", Hint: "Only used if you have off Fridays. Use 0 or 1."},
	}
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

func parseFloat(value string, label string) (float64, error) {
	if value == "" {
		return 0, fmt.Errorf("%s is required", label)
	}

	out, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%s must be a number", label)
	}

	if out < 0 {
		return 0, fmt.Errorf("%s cannot be negative", label)
	}

	return out, nil
}

func parseInt(value string, label string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("%s is required", label)
	}

	out, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a whole number", label)
	}

	if out < 0 {
		return 0, fmt.Errorf("%s cannot be negative", label)
	}

	return out, nil
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "y", "yes", "true":
		return true, nil
	case "n", "no", "false":
		return false, nil
	default:
		return false, fmt.Errorf("off fridays must be yes or no")
	}
}

func bannerArt() string {
	return strings.TrimSpace(`
 ____  _______ ___
|  _ \|_   _|_ _|
| |_) | | |  | |
|  __/  | |  | |
|_|     |_| |___|
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

func centerText(width int, text string) string {
	if width <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		padding := (width - len(line)) / 2
		if padding > 0 {
			lines[i] = strings.Repeat(" ", padding) + line
		}
	}

	return strings.Join(lines, "\n")
}

func centerBlock(width int, text string) string {
	if width <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	maxLen := 0
	for _, line := range lines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	padding := (width - maxLen) / 2
	if padding < 0 {
		padding = 0
	}

	prefix := strings.Repeat(" ", padding)
	for i, line := range lines {
		lines[i] = prefix + line
	}

	return strings.Join(lines, "\n")
}
