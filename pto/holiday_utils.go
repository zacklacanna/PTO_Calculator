package pto

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"pto_calculator/config"
)

func AddHoliday(holiday *config.Holiday) error {
	if err := LoadHolidays(); err != nil {
		return err
	}

	if holiday.Name == "" {
		return fmt.Errorf("holiday name is required")
	}

	holiday.Date = normalizeDate(holiday.Date)
	if holiday.Date.IsZero() {
		return fmt.Errorf("holiday date is required")
	}

	if existing, ok := config.GetHolidays().Holidays[holiday.Date]; ok {
		return fmt.Errorf("holiday %s already exists on %s", existing.Name, holiday.Date.Format("2006-01-02"))
	}

	holidays := config.GetHolidays()
	if holidays.Holidays == nil {
		holidays.Holidays = map[time.Time]config.Holiday{}
	}
	holidays.Holidays[holiday.Date] = *holiday

	return config.Save(config.HolidaysPath, holidays)
}

func RemoveHoliday(date time.Time) error {
	if err := LoadHolidays(); err != nil {
		return err
	}

	date = normalizeDate(date)
	holidays := config.GetHolidays()

	if holidays.Holidays == nil {
		return fmt.Errorf("no saved holidays found")
	}

	if _, ok := holidays.Holidays[date]; !ok {
		return fmt.Errorf("holiday on %s not found", date.Format("2006-01-02"))
	}

	delete(holidays.Holidays, date)
	return config.Save(config.HolidaysPath, holidays)
}

func ListHolidays() ([]config.Holiday, error) {
	if err := LoadHolidays(); err != nil {
		return nil, err
	}

	holidays := make([]config.Holiday, 0, len(config.GetHolidays().Holidays))
	for _, holiday := range config.GetHolidays().Holidays {
		holidays = append(holidays, holiday)
	}

	slices.SortFunc(holidays, func(a config.Holiday, b config.Holiday) int {
		if cmp := a.Date.Compare(b.Date); cmp != 0 {
			return cmp
		}
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})

	return holidays, nil
}

func CheckHoliday(name string, startDate time.Time, endDate time.Time) (config.Holiday, bool) {
	if err := LoadHolidays(); err != nil {
		return config.Holiday{}, false
	}

	name = strings.TrimSpace(strings.ToLower(name))
	for date, holiday := range config.GetHolidays().Holidays {
		if name != "" && strings.ToLower(holiday.Name) != name {
			continue
		}
		if !normalizeDate(date).Before(normalizeDate(startDate)) && !normalizeDate(date).After(normalizeDate(endDate)) {
			return holiday, true
		}
	}

	return config.Holiday{}, false
}
