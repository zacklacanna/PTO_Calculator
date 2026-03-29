package tli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"pto_calculator/config"
	"pto_calculator/pto"

	tea "charm.land/bubbletea/v2"
)

type screenMode int

const (
	modeWelcome screenMode = iota
	modeSetupForm
	modeMenu
	modeAddTrip
	modeListTrips
	modeRemoveTrip
	modeListHolidays
	modeAddHoliday
	modeRemoveHoliday
)

type startupModel struct {
	mode                screenMode
	width               int
	height              int
	fields              []setupField
	currentField        int
	visible             []int
	menuOptions         []menuOption
	currentMenuOption   int
	addFields           []tripField
	currentAddField     int
	addHolidayFields    []holidayField
	currentHolidayField int
	removeTrip          removeField
	removeHoliday       removeField
	tripList            []config.Trip
	holidayList         []config.Holiday
	status              string
	errText             string
	configReady         bool
	lastAddedTripName   string
	lastAddedHoliday    string
	currentPTOText      string
}

type menuOption struct {
	Title       string
	Description string
	Action      string
}

func InitTUI() error {
	model, err := newStartupModel()
	if err != nil {
		return err
	}

	_, err = tea.NewProgram(model).Run()
	return err
}

func newStartupModel() (startupModel, error) {
	setupFields := defaultSetupFields()
	m := startupModel{
		fields:           setupFields,
		visible:          visibleFieldIndexes(setupFields),
		menuOptions:      defaultMenuOptions(),
		addFields:        defaultTripFields(),
		addHolidayFields: defaultHolidayFields(),
		removeTrip:       defaultRemoveTripField(),
		removeHoliday:    defaultRemoveHolidayField(),
		status:           "Press enter to start setup.",
		currentPTOText:   muted("Unavailable"),
	}

	loaded, err := loadExistingConfig()
	if err != nil {
		return m, err
	}

	if loaded {
		m.mode = modeMenu
		m.configReady = true
		m.status = fmt.Sprintf("Config loaded for %s.", config.GetSettings().UserName)
		m.refreshMenuSummary()
		return m, nil
	}

	m.mode = modeWelcome
	return m, nil
}

func (m startupModel) Init() tea.Cmd {
	return nil
}

func (m startupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		switch m.mode {
		case modeWelcome:
			return m.updateWelcome(msg)
		case modeSetupForm:
			return m.updateSetupForm(msg)
		case modeMenu:
			return m.updateMenu(msg)
		case modeAddTrip:
			return m.updateAddTripForm(msg)
		case modeListTrips, modeListHolidays:
			return m.updateListScreen(msg)
		case modeRemoveTrip:
			return m.updateRemoveTripForm(msg)
		case modeAddHoliday:
			return m.updateAddHolidayForm(msg)
		case modeRemoveHoliday:
			return m.updateRemoveHolidayForm(msg)
		}
	}

	return m, nil
}

func (m startupModel) updateWelcome(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		m.mode = modeSetupForm
		m.visible = visibleFieldIndexes(m.fields)
		m.status = "Fill in each field. Tab moves forward, shift+tab moves back."
		m.errText = ""
	}
	return m, nil
}

func (m startupModel) updateSetupForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.visible = visibleFieldIndexes(m.fields)
	current := m.currentVisibleField()

	switch msg.String() {
	case "up", "shift+tab":
		if current > 0 {
			m.currentField = m.visible[current-1]
		}
		m.errText = ""
		return m, nil
	case "down", "tab":
		if current < len(m.visible)-1 {
			m.currentField = m.visible[current+1]
		}
		m.errText = ""
		return m, nil
	case "enter":
		if current == len(m.visible)-1 {
			cfg, err := buildConfig(m.fields)
			if err != nil {
				m.errText = err.Error()
				return m, nil
			}

			if err := saveConfig(cfg); err != nil {
				m.errText = err.Error()
				return m, nil
			}

			m.configReady = true
			m.status = fmt.Sprintf("Config saved for %s.", cfg.UserName)
			m.refreshMenuSummary()
			m.mode = modeMenu
			m.errText = ""
			return m, nil
		}

		m.currentField = m.visible[current+1]
		m.errText = ""
		return m, nil
	case "backspace":
		field := &m.fields[m.currentField]
		if len(field.Value) > 0 {
			field.Value = field.Value[:len(field.Value)-1]
			field.Edited = true
		}
		m.errText = ""
		return m, nil
	}

	if text := msg.Key().Text; text != "" {
		field := &m.fields[m.currentField]
		if !field.Edited {
			field.Value = ""
		}
		field.Value += text
		field.Edited = true
		m.visible = visibleFieldIndexes(m.fields)
		if len(m.visible) > 0 && m.currentVisibleField() >= len(m.visible) {
			m.currentField = m.visible[len(m.visible)-1]
		}
		m.errText = ""
	}

	return m, nil
}

func (m startupModel) updateMenu(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.currentMenuOption%3 > 0 {
			m.currentMenuOption--
		}
		return m, nil
	case "down", "j", "tab":
		if m.currentMenuOption%3 < 2 && m.currentMenuOption+1 < len(m.menuOptions) {
			m.currentMenuOption++
		}
		return m, nil
	case "right", "l":
		if m.currentMenuOption < 3 && m.currentMenuOption+3 < len(m.menuOptions) {
			m.currentMenuOption += 3
		}
		return m, nil
	case "left", "h":
		if m.currentMenuOption >= 3 {
			m.currentMenuOption -= 3
		}
		return m, nil
	case "enter":
		switch m.menuOptions[m.currentMenuOption].Action {
		case "add-trip":
			m.mode = modeAddTrip
			m.addFields = defaultTripFields()
			m.currentAddField = 0
		case "list-trips":
			trips, err := pto.ListTrips()
			if err != nil {
				m.status = err.Error()
				return m, nil
			}
			m.tripList = trips
			m.mode = modeListTrips
		case "remove-trip":
			m.removeTrip = defaultRemoveTripField()
			m.mode = modeRemoveTrip
		case "list-holidays":
			holidays, err := pto.ListHolidays()
			if err != nil {
				m.status = err.Error()
				return m, nil
			}
			m.holidayList = holidays
			m.mode = modeListHolidays
		case "add-holiday":
			m.addHolidayFields = defaultHolidayFields()
			m.currentHolidayField = 0
			m.mode = modeAddHoliday
		case "remove-holiday":
			m.removeHoliday = defaultRemoveHolidayField()
			m.mode = modeRemoveHoliday
		}
		m.errText = ""
		return m, nil
	}

	return m, nil
}

func (m startupModel) updateAddTripForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeMenu
		m.errText = ""
		return m, nil
	case "up", "shift+tab":
		if m.currentAddField > 0 {
			m.currentAddField--
		}
		m.errText = ""
		return m, nil
	case "down", "tab":
		if m.currentAddField < len(m.addFields)-1 {
			m.currentAddField++
		}
		m.errText = ""
		return m, nil
	case "enter":
		if m.currentAddField == len(m.addFields)-1 {
			trip, err := saveTrip(m.addFields)
			if err != nil {
				m.errText = err.Error()
				return m, nil
			}

			m.lastAddedTripName = trip.Name
			m.status = fmt.Sprintf("Trip %s saved from %s to %s.", trip.Name, trip.StartDate.Format("2006-01-02"), trip.EndDate.Format("2006-01-02"))
			m.addFields = defaultTripFields()
			m.currentAddField = 0
			m.errText = ""
			m.refreshMenuSummary()
			m.mode = modeMenu
			return m, nil
		}

		m.currentAddField++
		m.errText = ""
		return m, nil
	case "backspace":
		field := &m.addFields[m.currentAddField]
		if len(field.Value) > 0 {
			field.Value = field.Value[:len(field.Value)-1]
			field.Edited = true
		}
		m.errText = ""
		return m, nil
	}

	if text := msg.Key().Text; text != "" {
		field := &m.addFields[m.currentAddField]
		if !field.Edited {
			field.Value = ""
		}
		field.Value += text
		field.Edited = true
		m.errText = ""
	}

	return m, nil
}

func (m startupModel) updateAddHolidayForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeMenu
		m.errText = ""
		return m, nil
	case "up", "shift+tab":
		if m.currentHolidayField > 0 {
			m.currentHolidayField--
		}
		m.errText = ""
		return m, nil
	case "down", "tab":
		if m.currentHolidayField < len(m.addHolidayFields)-1 {
			m.currentHolidayField++
		}
		m.errText = ""
		return m, nil
	case "enter":
		if m.currentHolidayField == len(m.addHolidayFields)-1 {
			holiday, err := saveHoliday(m.addHolidayFields)
			if err != nil {
				m.errText = err.Error()
				return m, nil
			}

			m.lastAddedHoliday = holiday.Name
			m.status = fmt.Sprintf("Holiday %s saved on %s.", holiday.Name, holiday.Date.Format("2006-01-02"))
			m.addHolidayFields = defaultHolidayFields()
			m.currentHolidayField = 0
			m.errText = ""
			m.refreshMenuSummary()
			m.mode = modeMenu
			return m, nil
		}

		m.currentHolidayField++
		m.errText = ""
		return m, nil
	case "backspace":
		field := &m.addHolidayFields[m.currentHolidayField]
		if len(field.Value) > 0 {
			field.Value = field.Value[:len(field.Value)-1]
			field.Edited = true
		}
		m.errText = ""
		return m, nil
	}

	if text := msg.Key().Text; text != "" {
		field := &m.addHolidayFields[m.currentHolidayField]
		if !field.Edited {
			field.Value = ""
		}
		field.Value += text
		field.Edited = true
		m.errText = ""
	}

	return m, nil
}

func (m startupModel) updateRemoveTripForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeMenu
		m.errText = ""
		return m, nil
	case "enter":
		if err := removeTripByName(m.removeTrip); err != nil {
			m.errText = err.Error()
			return m, nil
		}
		m.status = fmt.Sprintf("Trip %s removed.", strings.TrimSpace(m.removeTrip.Value))
		m.removeTrip = defaultRemoveTripField()
		m.errText = ""
		m.refreshMenuSummary()
		m.mode = modeMenu
		return m, nil
	case "backspace":
		if len(m.removeTrip.Value) > 0 {
			m.removeTrip.Value = m.removeTrip.Value[:len(m.removeTrip.Value)-1]
			m.removeTrip.Edited = true
		}
		m.errText = ""
		return m, nil
	}

	if text := msg.Key().Text; text != "" {
		if !m.removeTrip.Edited {
			m.removeTrip.Value = ""
		}
		m.removeTrip.Value += text
		m.removeTrip.Edited = true
		m.errText = ""
	}

	return m, nil
}

func (m startupModel) updateRemoveHolidayForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeMenu
		m.errText = ""
		return m, nil
	case "enter":
		if err := removeHolidayByDate(m.removeHoliday); err != nil {
			m.errText = err.Error()
			return m, nil
		}
		m.status = fmt.Sprintf("Holiday on %s removed.", strings.TrimSpace(m.removeHoliday.Value))
		m.removeHoliday = defaultRemoveHolidayField()
		m.errText = ""
		m.refreshMenuSummary()
		m.mode = modeMenu
		return m, nil
	case "backspace":
		if len(m.removeHoliday.Value) > 0 {
			m.removeHoliday.Value = m.removeHoliday.Value[:len(m.removeHoliday.Value)-1]
			m.removeHoliday.Edited = true
		}
		m.errText = ""
		return m, nil
	}

	if text := msg.Key().Text != ""; text {
		keyText := msg.Key().Text
		if !m.removeHoliday.Edited {
			m.removeHoliday.Value = ""
		}
		m.removeHoliday.Value += keyText
		m.removeHoliday.Edited = true
		m.errText = ""
	}

	return m, nil
}

func (m startupModel) updateListScreen(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		m.mode = modeMenu
		m.errText = ""
	}
	return m, nil
}

func (m startupModel) View() tea.View {
	var b strings.Builder

	b.WriteString("\n\n")
	b.WriteString(centerText(m.width, accent(bannerArt())))
	b.WriteString("\n\n")

	switch m.mode {
	case modeWelcome:
		b.WriteString(centerBlock(m.width, welcomeCopy()))
	case modeSetupForm:
		b.WriteString(centerBlock(m.width, m.renderSetupForm()))
	case modeMenu:
		b.WriteString(centerBlock(m.width, m.renderMenu()))
	case modeAddTrip:
		b.WriteString(centerBlock(m.width, m.renderAddTripForm()))
	case modeListTrips:
		b.WriteString(centerBlock(m.width, renderTripList(m.tripList)))
	case modeRemoveTrip:
		b.WriteString(centerBlock(m.width, m.renderRemoveTripForm()))
	case modeListHolidays:
		b.WriteString(centerBlock(m.width, renderHolidayList(m.holidayList)))
	case modeAddHoliday:
		b.WriteString(centerBlock(m.width, m.renderAddHolidayForm()))
	case modeRemoveHoliday:
		b.WriteString(centerBlock(m.width, m.renderRemoveHolidayForm()))
	}

	view := tea.NewView(b.String())
	view.AltScreen = true
	return view
}

func (m startupModel) renderSetupForm() string {
	m.visible = visibleFieldIndexes(m.fields)
	lines := []string{
		muted("Build your PTO profile once. You can refine it later."),
		"",
	}

	for position, index := range m.visible {
		field := m.fields[index]
		prefix := muted("  ")
		label := muted(field.Label)
		value := field.Value

		if index == m.currentField {
			prefix = accent("› ")
			label = highlight(field.Label)
			value = strong(value + " ")
		}

		lines = append(lines, fmt.Sprintf("%s%-16s %s", prefix, label, value))
		if index == m.currentField {
			lines = append(lines, muted(field.Hint))
		}
		if position != len(m.visible)-1 {
			lines = append(lines, "")
		}
	}

	if m.errText != "" {
		lines = append(lines, "")
		lines = append(lines, danger("Error: "+m.errText))
	}

	lines = append(lines, "")
	lines = append(lines, muted("Type to replace a default. Tab moves forward. Shift+Tab moves back."))
	lines = append(lines, muted("Press enter on the last field to save."))

	return box("First-Time Setup", lines)
}

func (m startupModel) renderMenu() string {
	name := config.GetSettings().UserName
	if name == "" {
		name = "there"
	}

	summaryLines := []string{
		success("Welcome, " + name + "."),
		muted(m.status),
		"",
		strong("Current PTO today: ") + highlight(m.currentPTOText),
	}

	if m.lastAddedTripName != "" {
		summaryLines = append(summaryLines, "")
		summaryLines = append(summaryLines, success("Last added trip: "+m.lastAddedTripName))
	}
	if m.lastAddedHoliday != "" {
		summaryLines = append(summaryLines, success("Last added holiday: "+m.lastAddedHoliday))
	}

	tripLines := []string{
		muted("Trip actions"),
		"",
	}
	holidayLines := []string{
		muted("Holiday actions"),
		"",
	}

	for i, option := range m.menuOptions[:3] {
		absoluteIndex := i
		prefix := muted("  ")
		title := strong(option.Title)
		description := muted(option.Description)

		if absoluteIndex == m.currentMenuOption {
			prefix = accent("› ")
			title = highlight(option.Title)
		}

		tripLines = append(tripLines, fmt.Sprintf("%s%s", prefix, title))
		tripLines = append(tripLines, fmt.Sprintf("  %s", description))
		if i != len(m.menuOptions[:3])-1 {
			tripLines = append(tripLines, "")
		}
	}

	for i, option := range m.menuOptions[3:] {
		absoluteIndex := i + 3
		prefix := muted("  ")
		title := strong(option.Title)
		description := muted(option.Description)

		if absoluteIndex == m.currentMenuOption {
			prefix = accent("› ")
			title = highlight(option.Title)
		}

		holidayLines = append(holidayLines, fmt.Sprintf("%s%s", prefix, title))
		holidayLines = append(holidayLines, fmt.Sprintf("  %s", description))
		if i != len(m.menuOptions[3:])-1 {
			holidayLines = append(holidayLines, "")
		}
	}

	summaryLines = append(summaryLines, "")
	summaryLines = append(summaryLines, muted("Use arrows or h/j/k/l to move. Enter opens the selected screen."))
	summaryLines = append(summaryLines, muted("Press q to quit."))

	summary := box("Main Menu", summaryLines)
	tripPanel := box("Trips", tripLines)
	holidayPanel := box("Holidays", holidayLines)
	panels := joinColumns(tripPanel, holidayPanel, 4)
	summary = centerBlock(blockWidth(panels), summary)

	return strings.Join([]string{
		summary,
		"",
		panels,
	}, "\n")
}

func (m startupModel) renderAddTripForm() string {
	lines := []string{
		muted("Create a new trip and run it through your PTO validation logic."),
		"",
	}

	for i, field := range m.addFields {
		prefix := muted("  ")
		label := muted(field.Label)
		value := field.Value

		if i == m.currentAddField {
			prefix = accent("› ")
			label = highlight(field.Label)
			value = strong(value + " ")
		}

		lines = append(lines, fmt.Sprintf("%s%-12s %s", prefix, label, value))
		if i == m.currentAddField {
			lines = append(lines, muted(field.Hint))
		}
		if i != len(m.addFields)-1 {
			lines = append(lines, "")
		}
	}

	if m.errText != "" {
		lines = append(lines, "")
		lines = append(lines, danger("Error: "+m.errText))
	}

	lines = append(lines, "")
	lines = append(lines, muted("Type to replace a default. Enter saves on the last field."))
	lines = append(lines, muted("Esc returns to the main menu."))

	return box("Add Trip", lines)
}

func (m startupModel) renderAddHolidayForm() string {
	lines := []string{
		muted("Add a holiday that should not consume PTO."),
		"",
	}

	for i, field := range m.addHolidayFields {
		prefix := muted("  ")
		label := muted(field.Label)
		value := field.Value

		if i == m.currentHolidayField {
			prefix = accent("› ")
			label = highlight(field.Label)
			value = strong(value + " ")
		}

		lines = append(lines, fmt.Sprintf("%s%-12s %s", prefix, label, value))
		if i == m.currentHolidayField {
			lines = append(lines, muted(field.Hint))
		}
		if i != len(m.addHolidayFields)-1 {
			lines = append(lines, "")
		}
	}

	if m.errText != "" {
		lines = append(lines, "")
		lines = append(lines, danger("Error: "+m.errText))
	}

	lines = append(lines, "")
	lines = append(lines, muted("Type to replace a default. Enter saves on the last field."))
	lines = append(lines, muted("Esc returns to the main menu."))

	return box("Add Holiday", lines)
}

func (m startupModel) renderRemoveTripForm() string {
	return renderSingleFieldForm(
		"Remove Trip",
		"Remove a saved trip by name.",
		m.removeTrip,
		true,
		m.errText,
		"Press enter to remove the trip. Esc returns to the main menu.",
	)
}

func (m startupModel) renderRemoveHolidayForm() string {
	return renderSingleFieldForm(
		"Remove Holiday",
		"Remove a saved holiday by date.",
		m.removeHoliday,
		true,
		m.errText,
		"Press enter to remove the holiday. Esc returns to the main menu.",
	)
}

func (m startupModel) currentVisibleField() int {
	for i, index := range m.visible {
		if index == m.currentField {
			return i
		}
	}
	if len(m.visible) == 0 {
		return 0
	}
	m.currentField = m.visible[0]
	return 0
}

func (m *startupModel) refreshMenuSummary() {
	balance, err := pto.CalculatePtoOnDate(time.Now())
	if err != nil {
		m.currentPTOText = "Unavailable"
		return
	}

	m.currentPTOText = fmt.Sprintf("%.1f hours", balance)
}

func loadExistingConfig() (bool, error) {
	info, err := os.Stat(config.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if info.Size() == 0 {
		return false, nil
	}

	var cfg config.Config
	if err := config.Load(config.ConfigPath, &cfg); err != nil {
		return false, nil
	}

	*config.GetSettings() = cfg
	return true, nil
}

func defaultMenuOptions() []menuOption {
	return []menuOption{
		{Title: "Add Trip", Description: "Create a new trip and validate it against PTO rules.", Action: "add-trip"},
		{Title: "List Trips", Description: "View all saved trips in one place.", Action: "list-trips"},
		{Title: "Remove Trip", Description: "Delete a trip you no longer need.", Action: "remove-trip"},
		{Title: "List Holidays", Description: "Review holiday days that should not use PTO.", Action: "list-holidays"},
		{Title: "Add Holiday", Description: "Add a company holiday or personal no-PTO day.", Action: "add-holiday"},
		{Title: "Remove Holiday", Description: "Remove an existing holiday entry.", Action: "remove-holiday"},
	}
}
