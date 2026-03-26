package tli

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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
		{Key: "initialBalance", Label: "Starting hours", Default: "120", Hint: "Current PTO balance in hours."},
		{Key: "rate", Label: "Accrual rate", Default: "4.62", Hint: "Hours earned each pay period."},
		{Key: "maxdays", Label: "Max days", Default: "25", Hint: "Maximum PTO days allowed to sit in the bank."},
		{Key: "dailyHours", Label: "Daily hours", Default: "8", Hint: "Hours deducted for one PTO workday."},
		{Key: "firstDay", Label: "First work day", Default: "2026-01-05", Hint: "Use YYYY-MM-DD."},
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

func centerText(width int, text string) string {
	if width <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		padding := (width - visibleWidth(line)) / 2
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
		if visibleWidth(line) > maxLen {
			maxLen = visibleWidth(line)
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

func ansi(code string, text string) string {
	return "\033[" + code + "m" + text + "\033[0m"
}

func muted(text string) string {
	return ansi("38;5;245", text)
}

func accent(text string) string {
	return ansi("38;5;81", text)
}

func highlight(text string) string {
	return ansi("38;5;229", text)
}

func success(text string) string {
	return ansi("38;5;120", text)
}

func danger(text string) string {
	return ansi("38;5;210", text)
}

func strong(text string) string {
	return ansi("1", text)
}

func box(title string, lines []string) string {
	width := visibleWidth(title) + 4
	for _, line := range lines {
		if visibleWidth(line) > width-4 {
			width = visibleWidth(line) + 4
		}
	}

	top := "┌" + strings.Repeat("─", width-2) + "┐"
	header := "│ " + strong(title) + strings.Repeat(" ", width-visibleWidth(title)-3) + "│"
	body := make([]string, 0, len(lines)+3)
	body = append(body, top, header, "├"+strings.Repeat("─", width-2)+"┤")
	for _, line := range lines {
		padding := width - visibleWidth(line) - 3
		if padding < 0 {
			padding = 0
		}
		body = append(body, "│ "+line+strings.Repeat(" ", padding)+"│")
	}
	body = append(body, "└"+strings.Repeat("─", width-2)+"┘")
	return strings.Join(body, "\n")
}

func visibleWidth(text string) int {
	width := 0
	for i := 0; i < len(text); {
		if text[i] == '\x1b' && i+1 < len(text) && text[i+1] == '[' {
			i += 2
			for i < len(text) && text[i] != 'm' {
				i++
			}
			if i < len(text) {
				i++
			}
			continue
		}

		_, size := utf8.DecodeRuneInString(text[i:])
		if size == 0 {
			break
		}
		width++
		i += size
	}

	return width
}
