package tli

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"pto_calculator/config"
)

type screenMode int

const (
	modeWelcome screenMode = iota
	modeForm
	modeReady
)

type startupModel struct {
	mode         screenMode
	width        int
	height       int
	fields       []setupField
	currentField int
	visible      []int
	status       string
	errText      string
	configReady  bool
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
	m := startupModel{
		fields: defaultSetupFields(),
		visible: visibleFieldIndexes(defaultSetupFields()),
		status: "Press enter to start setup.",
	}

	loaded, err := loadExistingConfig()
	if err != nil {
		return m, err
	}

	if loaded {
		m.mode = modeReady
		m.configReady = true
		m.status = fmt.Sprintf("Config loaded for %s. Main dashboard is next.", config.GetSettings().UserName)
		return m, nil
	}

	m.mode = modeWelcome
	m.visible = visibleFieldIndexes(m.fields)
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
		case modeForm:
			return m.updateForm(msg)
		case modeReady:
			if msg.String() == "enter" {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m startupModel) updateWelcome(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "enter" {
		m.mode = modeForm
		m.visible = visibleFieldIndexes(m.fields)
		m.status = "Fill in each field. Tab moves forward, shift+tab moves back."
		m.errText = ""
	}
	return m, nil
}

func (m startupModel) updateForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
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

			m.mode = modeReady
			m.configReady = true
			m.status = fmt.Sprintf("Config saved for %s. Main dashboard is next.", cfg.UserName)
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
		if field.Key == "hasOffFridays" {
			m.visible = visibleFieldIndexes(m.fields)
			if !shouldShowFridayCycle(m.fields) && m.fields[m.currentField].Key == "hasOffFridays" {
				for _, index := range m.visible {
					if m.fields[index].Key != "whichFridayOff" {
						continue
					}
				}
				if m.currentVisibleField() >= len(m.visible) {
					m.currentField = m.visible[len(m.visible)-1]
				}
			}
		}
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
	case modeForm:
		b.WriteString(centerBlock(m.width, m.renderForm()))
	case modeReady:
		b.WriteString(centerBlock(m.width, m.renderReady()))
	}

	view := tea.NewView(b.String())
	view.AltScreen = true
	return view
}

func (m startupModel) renderForm() string {
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

func (m startupModel) renderReady() string {
	name := config.GetSettings().UserName
	if name == "" {
		name = "there"
	}

	return box("Setup Complete", []string{
		success("Your PTO profile is ready."),
		"",
		"Welcome, " + strong(name) + ".",
		muted(m.status),
		"",
		muted("Press enter to close this screen."),
	})
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
