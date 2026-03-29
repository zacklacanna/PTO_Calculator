package tli

import (
	"fmt"
	"os"
	"strings"

	"pto_calculator/config"

	tea "charm.land/bubbletea/v2"
)

type screenMode int

const (
	modeWelcome screenMode = iota
	modeSetupForm
	modeMenu
	modeAddTrip
)

type startupModel struct {
	mode              screenMode
	width             int
	height            int
	fields            []setupField
	currentField      int
	visible           []int
	menuOptions       []menuOption
	currentMenuOption int
	addFields         []tripField
	currentAddField   int
	status            string
	errText           string
	configReady       bool
	lastAddedTripName string
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
		fields:      setupFields,
		visible:     visibleFieldIndexes(setupFields),
		menuOptions: defaultMenuOptions(),
		addFields:   defaultTripFields(),
		status:      "Press enter to start setup.",
	}

	loaded, err := loadExistingConfig()
	if err != nil {
		return m, err
	}

	if loaded {
		m.mode = modeMenu
		m.configReady = true
		m.status = fmt.Sprintf("Config loaded for %s.", config.GetSettings().UserName)
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
			return m.updateAddForm(msg)
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

			m.mode = modeMenu
			m.configReady = true
			m.status = fmt.Sprintf("Config saved for %s.", cfg.UserName)
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
		if m.currentVisibleField() >= len(m.visible) {
			m.currentField = m.visible[len(m.visible)-1]
		}
		m.errText = ""
	}

	return m, nil
}

func (m startupModel) updateMenu(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.currentMenuOption > 0 {
			m.currentMenuOption--
		}
		return m, nil
	case "down", "j", "tab":
		if m.currentMenuOption < len(m.menuOptions)-1 {
			m.currentMenuOption++
		}
		return m, nil
	case "a":
		m.currentMenuOption = 0
		fallthrough
	case "enter":
		switch m.menuOptions[m.currentMenuOption].Action {
		case "add-trip":
			m.mode = modeAddTrip
			m.addFields = defaultTripFields()
			m.currentAddField = 0
			m.errText = ""
		default:
			m.status = m.menuOptions[m.currentMenuOption].Title + " is the next screen to build."
		}
		return m, nil
	}

	return m, nil
}

func (m startupModel) updateAddForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
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

			m.mode = modeMenu
			m.lastAddedTripName = trip.Name
			m.status = fmt.Sprintf("Trip %s saved from %s to %s.", trip.Name, trip.StartDate.Format("2006-01-02"), trip.EndDate.Format("2006-01-02"))
			m.errText = ""
			m.addFields = defaultTripFields()
			m.currentAddField = 0
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
		b.WriteString(centerBlock(m.width, m.renderAddForm()))
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

	lines := []string{
		success("Welcome, " + name + "."),
		muted(m.status),
		"",
	}

	for i, option := range m.menuOptions {
		prefix := muted("  ")
		title := strong(option.Title)
		description := muted(option.Description)

		if i == m.currentMenuOption {
			prefix = accent("› ")
			title = highlight(option.Title)
		}

		lines = append(lines, fmt.Sprintf("%s%s", prefix, title))
		lines = append(lines, fmt.Sprintf("  %s", description))
		if i != len(m.menuOptions)-1 {
			lines = append(lines, "")
		}
	}

	lines = append(lines, "")
	lines = append(lines, muted("Use up/down or j/k to choose. Enter opens the selected screen."))
	lines = append(lines, muted("Press q to quit."))

	if m.lastAddedTripName != "" {
		lines = append(lines, "")
		lines = append(lines, success("Last added trip: "+m.lastAddedTripName))
	}

	return box("Main Menu", lines)
}

func (m startupModel) renderAddForm() string {
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
