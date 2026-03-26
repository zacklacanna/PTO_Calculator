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
		m.status = "Fill in each field. Tab moves forward, shift+tab moves back."
		m.errText = ""
	}
	return m, nil
}

func (m startupModel) updateForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "shift+tab":
		if m.currentField > 0 {
			m.currentField--
		}
		m.errText = ""
		return m, nil
	case "down", "tab":
		if m.currentField < len(m.fields)-1 {
			m.currentField++
		}
		m.errText = ""
		return m, nil
	case "enter":
		if m.currentField == len(m.fields)-1 {
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

		m.currentField++
		m.errText = ""
		return m, nil
	case "backspace":
		field := &m.fields[m.currentField]
		if len(field.Value) > 0 {
			field.Value = field.Value[:len(field.Value)-1]
		}
		m.errText = ""
		return m, nil
	}

	if text := msg.Key().Text; text != "" {
		m.fields[m.currentField].Value += text
		m.errText = ""
	}

	return m, nil
}

func (m startupModel) View() tea.View {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(centerText(m.width, bannerArt()))
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
	var b strings.Builder

	b.WriteString("First-time setup\n")
	b.WriteString("Build your PTO profile once. You can edit it later.\n\n")

	for i, field := range m.fields {
		cursor := "  "
		if i == m.currentField {
			cursor = "> "
		}

		value := field.Value
		if value == "" {
			value = field.Placeholder
		}

		b.WriteString(fmt.Sprintf("%s%-16s %s\n", cursor, field.Label+":", value))
		if i == m.currentField {
			b.WriteString(fmt.Sprintf("  %s\n\n", field.Hint))
		}
	}

	if m.errText != "" {
		b.WriteString("Error: " + m.errText + "\n\n")
	}

	b.WriteString("Controls: type to edit, tab to move, enter to continue, q to quit")
	return b.String()
}

func (m startupModel) renderReady() string {
	name := config.GetSettings().UserName
	if name == "" {
		name = "there"
	}

	return fmt.Sprintf(
		"Setup complete\n\nWelcome, %s.\n%s\n\nPress enter to close this screen.",
		name,
		m.status,
	)
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
