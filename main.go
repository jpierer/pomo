package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/gen2brain/beeep"
)

const (
	workView int = iota
	pauseView
	settingView
	quitView
)

var (
	lightColor    = lg.Color("#F1F5F9")
	workModeColor = lg.Color("#81edf6ff")
	pauseColor    = lg.Color("#38e979ff")
	darkColor     = lg.Color("#333")
	blurColor     = lg.Color("#767676")

	workTitles = []string{
		"beast mode",
		"flow state",
		"focus zone",
		"deep work",
		"laser focus",
		"hustle hard",
		"zone mode",
		"turbo mode",
	}

	pauseTitles = []string{
		"chill zone",
		"recharge",
		"breathe",
		"refresh",
		"stretch it out",
		"coffee time",
		"take five",
		"cool down",
	}
)

type settings struct {
	WorkMinutes  int
	PauseMinutes int
	AutoMode     bool
}

type model struct {
	width             int
	height            int
	workRemaining     time.Duration
	pauseRemaining    time.Duration
	state             int
	beforeState       int
	stopped           bool
	settings          settings
	workMinutes       int
	pauseMinutes      int
	autoIterateStates bool
	modeTitle         string
	quitSelected      int
	settingInputs     []textinput.Model
	settingsIndex     int
}

func main() {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		log.Fatalf("Error to run pomo %+v", err)
	}

}

func NewModel() model {

	workMinutes := 25
	pauseMinutes := 5
	autoMode := false

	settings, err := LoadSettings()
	if err == nil {
		workMinutes = settings.WorkMinutes
		pauseMinutes = settings.PauseMinutes
		autoMode = settings.AutoMode
	}

	// Create settings input fields
	settingInputs := make([]textinput.Model, 3)
	values := []string{fmt.Sprint(workMinutes), fmt.Sprint(pauseMinutes), "[ ]"}

	for i := range settingInputs {
		settingInputs[i] = textinput.New()
		settingInputs[i].SetValue(values[i])
		settingInputs[i].Prompt = ""
		settingInputs[i].CharLimit = 3

		if i == 0 {
			settingInputs[i].Focus()
		} else {
			settingInputs[i].Blur()
		}
	}

	m := &model{
		workMinutes:       workMinutes,
		workRemaining:     time.Duration(workMinutes) * time.Minute,
		pauseMinutes:      pauseMinutes,
		pauseRemaining:    time.Duration(pauseMinutes) * time.Minute,
		stopped:           true,
		state:             workView,
		modeTitle:         "",
		autoIterateStates: autoMode,
		quitSelected:      0,
		settingInputs:     settingInputs,
		settingsIndex:     0,
	}

	m.modeTitle = GetRandomModeTitle(*m)

	settingInputs[2].SetValue("[ ]")
	if m.autoIterateStates {
		settingInputs[2].SetValue("[x]")
	}

	return *m
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("Pomo")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// on key press
	case tea.KeyMsg:

		switch msg.String() {

		// Quit View
		case "ctrl+c", "q":
			// Open confirm
			m.state = quitView

		// Navigation in quit view
		case "left":
			switch m.state {
			case quitView:
				m.quitSelected = 0 // No
			case settingView:
				// Navigate between settings fields
				m.settingInputs[m.settingsIndex].Blur()
				if m.settingsIndex > 0 {
					m.settingsIndex--
				}
				m.settingInputs[m.settingsIndex].Focus()
			}

		case "right":
			switch m.state {
			case quitView:
				m.quitSelected = 1 // Yes
			case settingView:
				// Navigate between settings fields
				m.settingInputs[m.settingsIndex].Blur()
				if m.settingsIndex < len(m.settingInputs)-1 {
					m.settingsIndex++
				}
				m.settingInputs[m.settingsIndex].Focus()
			}

		case "up":
			if m.state == settingView {
				// Increase value (only for numeric fields)
				if m.settingsIndex == 0 && m.workMinutes < 60 {
					m.workMinutes++
					m.settingInputs[0].SetValue(fmt.Sprint(m.workMinutes))
					m.workRemaining = time.Duration(m.workMinutes) * time.Minute
				} else if m.settingsIndex == 1 && m.pauseMinutes < 60 {
					m.pauseMinutes++
					m.settingInputs[1].SetValue(fmt.Sprint(m.pauseMinutes))
					m.pauseRemaining = time.Duration(m.pauseMinutes) * time.Minute
				}
			}

		case "down":
			if m.state == settingView {
				// Decrease value (only for numeric fields)
				if m.settingsIndex == 0 && m.workMinutes > 1 {
					m.workMinutes--
					m.settingInputs[0].SetValue(fmt.Sprint(m.workMinutes))
					m.workRemaining = time.Duration(m.workMinutes) * time.Minute
				} else if m.settingsIndex == 1 && m.pauseMinutes > 1 {
					m.pauseMinutes--
					m.settingInputs[1].SetValue(fmt.Sprint(m.pauseMinutes))
					m.pauseRemaining = time.Duration(m.pauseMinutes) * time.Minute
				}
			}

		// Confirm
		case "enter":
			switch m.state {
			case quitView:
				if m.quitSelected == 1 {
					return m, tea.Quit
				} else {
					m.state = m.beforeState
					m.modeTitle = GetRandomModeTitle(m)
				}
			case settingView:
				// Save settings
				m.settings = settings{
					WorkMinutes:  m.workMinutes,
					PauseMinutes: m.pauseMinutes,
					AutoMode:     m.autoIterateStates,
				}
				SaveSettings(m.settings)
				m.SwitchState(workView)
			}

		// Cancel quit (ESC or n)
		case "esc", "n":
			if m.state == quitView {
				m.state = m.beforeState
			}

		// Open Setting
		case "s":
			if m.state != settingView {
				m.stopped = true
				m.SwitchState(settingView)
			}

		//Work View
		case "w":
			m.stopped = true
			m.SwitchState(workView)

		// Pause View
		case "p":
			m.stopped = true
			m.SwitchState(pauseView)

		// Toggle Timer
		case " ":
			if m.state == workView || m.state == pauseView {
				wasStopped := m.stopped
				m.stopped = !m.stopped
				if wasStopped && !m.stopped {
					return m, tickCmd()
				}
			} else if m.state == settingView && m.settingsIndex == 2 {
				// Toggle AutoTimer checkbox
				m.autoIterateStates = !m.autoIterateStates
				if m.autoIterateStates {
					m.settingInputs[2].SetValue("[x]")
				} else {
					m.settingInputs[2].SetValue("[ ]")
				}
			}

		// Reset
		case "r":
			if m.state == workView || m.state == pauseView {
				m.ResetAndStop()
			}

		}

	// on tick
	case time.Time:
		if m.state == workView {
			if !m.stopped {
				m.workRemaining -= time.Second
			}
			if m.workRemaining < 0 {
				Notify("Time is up, take a break!")
				m.ResetAndStop()
				m.SwitchState(pauseView)

				// If auto iterate is enabled, start the next timer
				if m.autoIterateStates {
					m.stopped = false
				}

			}
		}
		if m.state == pauseView {
			if !m.stopped {
				m.pauseRemaining -= time.Second
			}
			if m.pauseRemaining < 0 {
				Notify("Time is up, get back to work!")
				m.ResetAndStop()
				m.SwitchState(workView)

				// If auto iterate is enabled, start the next timer
				if m.autoIterateStates {
					m.stopped = false
				}

			}
		}

		if !m.stopped {
			return m, tickCmd()
		}
	}

	return m, nil
}

func (m model) View() string {

	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	content := RenderView(&m) + "\n" + RenderHelp(m)

	viewStyle := lg.NewStyle().
		Padding(1, 2, 0, 2).
		Align(lg.Center).
		Width(60)

	return lg.Place(
		m.width,
		m.height,
		lg.Center,
		lg.Center,
		viewStyle.Render(content),
	)

}

func (m *model) SwitchState(newState int) {
	m.beforeState = m.state
	m.state = newState
	m.modeTitle = GetRandomModeTitle(*m)
}

func (m *model) ResetAndStop() {
	m.workRemaining = time.Duration(m.workMinutes) * time.Minute
	m.pauseRemaining = time.Duration(m.pauseMinutes) * time.Minute
	m.stopped = true
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return t
	})
}

func RenderView(m *model) string {
	switch m.state {
	case workView:
		return RenderDisplay(m)
	case pauseView:
		return RenderDisplay(m)
	case settingView:
		return RenderSettings(*m)
	case quitView:
		return RenderQuit(*m)
	}
	return ""
}

func RenderDisplay(m *model) string {

	currentTime := time.Now()
	currentTimeMin := int(currentTime.Hour())
	currentTimeSec := int(currentTime.Minute()) % 60

	// Format time as MM:SS
	currentTimeString := fmt.Sprintf("%02d:%02d", currentTimeMin, currentTimeSec)

	var timerRemaining time.Duration
	var timerColor lg.Color
	var progressPercent float64

	if m.state == workView {
		timerRemaining = m.workRemaining
		timerColor = workModeColor
		totalDuration := time.Duration(m.workMinutes) * time.Minute
		progressPercent = 1.0 - (m.workRemaining.Seconds() / totalDuration.Seconds())
	} else {
		timerRemaining = m.pauseRemaining
		timerColor = pauseColor
		totalDuration := time.Duration(m.pauseMinutes) * time.Minute
		progressPercent = 1.0 - (m.pauseRemaining.Seconds() / totalDuration.Seconds())
	}

	// Ensure progress is between 0 and 1
	if progressPercent < 0 {
		progressPercent = 0
	}
	if progressPercent > 1 {
		progressPercent = 1
	}

	tempProgress := progress.New(progress.WithDefaultGradient())
	tempProgress.Width = 52

	min := int(timerRemaining.Minutes())
	sec := int(timerRemaining.Seconds()) % 60

	// Format time as MM:SS
	timerString := fmt.Sprintf("%02d:%02d", min, sec)

	titleStyle := lg.NewStyle().
		Foreground(darkColor).
		Background(timerColor).
		Bold(true).
		Padding(0, 1)

	modeDisplay := titleStyle.Render(m.modeTitle)

	fullDisplay := modeDisplay + "\n\nTime: " + currentTimeString + " - Remaining: " + timerString + "\n\n" + tempProgress.ViewAs(progressPercent) + "\n\n"

	timerStyle := lg.NewStyle().
		Foreground(lightColor)

	return timerStyle.Render(fullDisplay)
}

func RenderSettings(m model) string {
	labels := []string{"Work", "Pause", "Auto Mode"}

	focusStyle := lg.NewStyle().
		Border(lg.NormalBorder()).
		BorderForeground(pauseColor).
		Width(12).
		Align(lg.Center).
		Padding(0, 1)

	blurStyle := lg.NewStyle().
		Border(lg.NormalBorder()).
		BorderForeground(blurColor).
		Width(12).
		Align(lg.Center).
		Padding(0, 1)

	labelStyle := lg.NewStyle().
		Width(12).
		Align(lg.Center).
		Foreground(lightColor).
		Bold(true).
		Padding(0, 1)

	checkboxBlurStyle := lg.NewStyle().
		Width(12).
		Align(lg.Center).
		Padding(1)

	checkboxFocusStyle := lg.NewStyle().
		Width(12).
		Align(lg.Center).
		Padding(1).
		Foreground(pauseColor)

	// Create each field as a column
	var columns []string
	for i, input := range m.settingInputs {
		label := labelStyle.Render(labels[i])
		var field string

		if i == 2 { // AutoTimer checkbox
			if input.Focused() {
				field = checkboxFocusStyle.Render(input.View())
			} else {
				field = checkboxBlurStyle.Render(input.View())
			}
		} else {
			if input.Focused() {
				field = focusStyle.Render(input.View())
			} else {
				field = blurStyle.Render(input.View())
			}
		}
		column := label + "\n" + field
		columns = append(columns, column)
	}

	// Join columns horizontally with spacing
	content := lg.JoinHorizontal(lg.Top, columns[0], "  ", columns[1], "  ", columns[2])

	titleStyle := lg.NewStyle().
		Foreground(pauseColor).
		Bold(true).
		Align(lg.Center).
		Margin(0, 0, 1, 0)

	title := titleStyle.Render("Pomo-Settings")

	settingsStyle := lg.NewStyle().
		Foreground(lightColor).
		Padding(2, 0).
		Align(lg.Center)

	return settingsStyle.Render(title + "\n\n" + content)
}

func RenderHelp(m model) string {
	helpStyle := lg.NewStyle().
		Foreground(blurColor).
		Faint(true).
		Padding(0, 0).
		Align(lg.Center)

	var helpText string
	switch m.state {
	case workView:
		helpText = "[SPACE] Toggle, [P]ause, [S]ettings, [R]eset, [Q]uit"
	case pauseView:
		helpText = "[SPACE] Toggle, [W]ork, [S]ettings, [R]eset, [Q]uit"
	case settingView:
		helpText = "[← →] Field  [↑ ↓] +/- min  [SPACE] Toggle  [ENTER] Save"
	}

	return helpStyle.Render(helpText)
}

func RenderQuit(m model) string {
	quitStyle := lg.NewStyle().
		Foreground(lightColor).
		Bold(true).
		Align(lg.Center)

	noStyle := lg.NewStyle().Foreground(blurColor)
	yesStyle := lg.NewStyle().Foreground(blurColor)

	if m.quitSelected == 0 {
		noStyle = lg.NewStyle().Foreground(pauseColor).Bold(true)
	} else {
		yesStyle = lg.NewStyle().Foreground(pauseColor).Bold(true)
	}

	noOption := noStyle.Render("[ No ]")
	yesOption := yesStyle.Render("[ Yes ]")

	content := "Really quit?\n\n" + noOption + "    " + yesOption + "\n\n" +
		lg.NewStyle().Foreground(blurColor).Faint(true).Render("← → to select, ENTER to confirm")

	return quitStyle.Render(content)
}

func GetRandomModeTitle(m model) string {
	switch m.state {
	case workView:
		return "- " + workTitles[rand.Intn(len(workTitles))] + " -"
	case pauseView:
		return "- " + pauseTitles[rand.Intn(len(pauseTitles))] + " -"
	}
	return ""
}

func Notify(message string) {
	err := beeep.Notify("Pomo", message, "")
	if err != nil {
		log.Println("Error showing notification:", err)
	}
}

func SaveSettings(settings settings) {
	// Save the current settings to a file in the ~/.pomo/settings.json.
	filePath := filepath.Join(os.Getenv("HOME"), ".pomo", "settings.json")

	// Create the directory if it doesn't exist
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		log.Println("Error creating settings directory:", err)
		return
	}

	data, err := json.Marshal(settings)
	if err != nil {
		log.Println("Error marshalling settings:", err)
		return
	}
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		log.Println("Error writing settings to file:", err)
	}
}

func LoadSettings() (settings, error) {
	var s settings
	filePath := filepath.Join(os.Getenv("HOME"), ".pomo", "settings.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Println("Error reading settings file:", err)
		return s, err
	}
	err = json.Unmarshal(data, &s)
	if err != nil {
		log.Println("Error unmarshalling settings:", err)
		return s, err
	}
	return s, nil
}
