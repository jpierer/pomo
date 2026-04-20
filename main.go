package main

import (
	"encoding/json"
	"flag"
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
	standView
	sitView
	settingView
	quitView
)

const (
	modeWorkPause int = iota
	modeDesk
)

var (
	lightColor    = lg.Color("#F1F5F9")
	workModeColor = lg.Color("#81edf6ff")
	pauseColor    = lg.Color("#38e979ff")
	standColor    = lg.Color("#fbbf24ff")
	sitColor      = lg.Color("#d8b4feff")
	darkColor     = lg.Color("#333")
	blurColor     = lg.Color("#767676")

	workTitles = []string{
		"--- WORK ---",
	}

	pauseTitles = []string{
		"--- PAUSE ---",
	}

	standTitles = []string{
		"^^^ STAND UP ^^^",
	}

	sitTitles = []string{
		"vvv SIT DOWN vvv",
	}
)

type settings struct {
	WorkMinutes  int
	PauseMinutes int
	StandMinutes int
	SitMinutes   int
	AutoMode     bool
}

type model struct {
	width             int
	height            int
	workRemaining     time.Duration
	pauseRemaining    time.Duration
	standRemaining    time.Duration
	sitRemaining      time.Duration
	state             int
	beforeState       int
	stopped           bool
	settings          settings
	workMinutes       int
	pauseMinutes      int
	standMinutes      int
	sitMinutes        int
	autoIterateStates bool
	modeTitle         string
	quitSelected      int
	settingInputs     []textinput.Model
	settingsIndex     int
	modeType          int
}

func main() {
	mode := flag.String("mode", "work", "Mode: 'work' for Work/Pause or 'desk' for Stand/Sit")
	flag.Parse()

	modeType := modeWorkPause
	if *mode == "desk" {
		modeType = modeDesk
	}

	p := tea.NewProgram(NewModel(modeType), tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		log.Fatalf("Error to run pomo %+v", err)
	}

}

func NewModel(modeType int) model {

	workMinutes := 25
	pauseMinutes := 5
	standMinutes := 30
	sitMinutes := 30
	autoMode := false

	settings, err := LoadSettings()
	if err == nil {
		workMinutes = settings.WorkMinutes
		pauseMinutes = settings.PauseMinutes
		autoMode = settings.AutoMode
		// Only override if non-zero values exist (for backwards compatibility)
		if settings.StandMinutes > 0 {
			standMinutes = settings.StandMinutes
		}
		if settings.SitMinutes > 0 {
			sitMinutes = settings.SitMinutes
		}
	}

	// Determine number of input fields based on mode
	numFields := 3 // Work, Pause, AutoMode for WorkPause mode
	if modeType == modeDesk {
		numFields = 3 // Stand, Sit, AutoMode
	}

	settingInputs := make([]textinput.Model, numFields)
	var values []string

	if modeType == modeWorkPause {
		values = []string{fmt.Sprint(workMinutes), fmt.Sprint(pauseMinutes), "[ ]"}
	} else {
		values = []string{fmt.Sprint(standMinutes), fmt.Sprint(sitMinutes), "[ ]"}
	}

	for i := range settingInputs {
		settingInputs[i] = textinput.New()
		if i < len(values) {
			settingInputs[i].SetValue(values[i])
		}
		settingInputs[i].Prompt = ""
		settingInputs[i].CharLimit = 3

		if i == 0 {
			settingInputs[i].Focus()
		} else {
			settingInputs[i].Blur()
		}
	}

	initialState := workView
	if modeType == modeDesk {
		initialState = standView
	}

	m := &model{
		workMinutes:       workMinutes,
		workRemaining:     time.Duration(workMinutes) * time.Minute,
		pauseMinutes:      pauseMinutes,
		pauseRemaining:    time.Duration(pauseMinutes) * time.Minute,
		standMinutes:      standMinutes,
		standRemaining:    time.Duration(standMinutes) * time.Minute,
		sitMinutes:        sitMinutes,
		sitRemaining:      time.Duration(sitMinutes) * time.Minute,
		stopped:           true,
		state:             initialState,
		modeTitle:         "",
		autoIterateStates: autoMode,
		quitSelected:      0,
		settingInputs:     settingInputs,
		settingsIndex:     0,
		modeType:          modeType,
	}

	m.modeTitle = GetRandomModeTitle(*m)

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
			// Open confirm (guard against double-q)
			if m.state != quitView {
				m.SwitchState(quitView)
			}

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
				if m.modeType == modeWorkPause {
					if m.settingsIndex == 0 && m.workMinutes < 60 {
						m.workMinutes++
						m.settingInputs[0].SetValue(fmt.Sprint(m.workMinutes))
						m.workRemaining = time.Duration(m.workMinutes) * time.Minute
					} else if m.settingsIndex == 1 && m.pauseMinutes < 60 {
						m.pauseMinutes++
						m.settingInputs[1].SetValue(fmt.Sprint(m.pauseMinutes))
						m.pauseRemaining = time.Duration(m.pauseMinutes) * time.Minute
					}
				} else {
					if m.settingsIndex == 0 && m.standMinutes < 60 {
						m.standMinutes++
						m.settingInputs[0].SetValue(fmt.Sprint(m.standMinutes))
						m.standRemaining = time.Duration(m.standMinutes) * time.Minute
					} else if m.settingsIndex == 1 && m.sitMinutes < 60 {
						m.sitMinutes++
						m.settingInputs[1].SetValue(fmt.Sprint(m.sitMinutes))
						m.sitRemaining = time.Duration(m.sitMinutes) * time.Minute
					}
				}
			}

		case "down":
			if m.state == settingView {
				// Decrease value (only for numeric fields)
				if m.modeType == modeWorkPause {
					if m.settingsIndex == 0 && m.workMinutes > 1 {
						m.workMinutes--
						m.settingInputs[0].SetValue(fmt.Sprint(m.workMinutes))
						m.workRemaining = time.Duration(m.workMinutes) * time.Minute
					} else if m.settingsIndex == 1 && m.pauseMinutes > 1 {
						m.pauseMinutes--
						m.settingInputs[1].SetValue(fmt.Sprint(m.pauseMinutes))
						m.pauseRemaining = time.Duration(m.pauseMinutes) * time.Minute
					}
				} else {
					if m.settingsIndex == 0 && m.standMinutes > 1 {
						m.standMinutes--
						m.settingInputs[0].SetValue(fmt.Sprint(m.standMinutes))
						m.standRemaining = time.Duration(m.standMinutes) * time.Minute
					} else if m.settingsIndex == 1 && m.sitMinutes > 1 {
						m.sitMinutes--
						m.settingInputs[1].SetValue(fmt.Sprint(m.sitMinutes))
						m.sitRemaining = time.Duration(m.sitMinutes) * time.Minute
					}
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
					StandMinutes: m.standMinutes,
					SitMinutes:   m.sitMinutes,
					AutoMode:     m.autoIterateStates,
				}
				SaveSettings(m.settings)
				if m.modeType == modeWorkPause {
					m.SwitchState(workView)
				} else {
					m.SwitchState(standView)
				}
			}

		// Cancel quit (ESC or n)
		case "esc", "n":
			if m.state == quitView {
				m.state = m.beforeState
				m.modeTitle = GetRandomModeTitle(m)
			}

		// Open Setting
		case "s":
			if m.state != settingView && m.state != quitView {
				m.stopped = true
				m.SwitchState(settingView)
			}

		// Mode Toggle - T
		case "t":
			if m.state == quitView {
				break
			}
			m.stopped = true
			if m.modeType == modeWorkPause {
				// Toggle between workView and pauseView
				if m.state == workView {
					m.SwitchState(pauseView)
				} else if m.state == pauseView {
					m.SwitchState(workView)
				} else {
					m.SwitchState(workView)
				}
			} else {
				// Toggle between standView and sitView
				if m.state == standView {
					m.SwitchState(sitView)
				} else if m.state == sitView {
					m.SwitchState(standView)
				} else {
					m.SwitchState(standView)
				}
			}

		// Toggle Timer (space)
		case " ":
			if m.state == workView || m.state == pauseView || m.state == standView || m.state == sitView {
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
			if m.state == workView || m.state == pauseView || m.state == standView || m.state == sitView {
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
		if m.state == standView {
			if !m.stopped {
				m.standRemaining -= time.Second
			}
			if m.standRemaining < 0 {
				Notify("Time to sit down!")
				m.ResetAndStop()
				m.SwitchState(sitView)

				// If auto iterate is enabled, start the next timer
				if m.autoIterateStates {
					m.stopped = false
				}

			}
		}
		if m.state == sitView {
			if !m.stopped {
				m.sitRemaining -= time.Second
			}
			if m.sitRemaining < 0 {
				Notify("Time to stand up!")
				m.ResetAndStop()
				m.SwitchState(standView)

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
	m.standRemaining = time.Duration(m.standMinutes) * time.Minute
	m.sitRemaining = time.Duration(m.sitMinutes) * time.Minute
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
	case standView:
		return RenderDisplay(m)
	case sitView:
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
	var totalDuration time.Duration

	if m.state == workView {
		timerRemaining = m.workRemaining
		timerColor = workModeColor
		totalDuration = time.Duration(m.workMinutes) * time.Minute
	} else if m.state == pauseView {
		timerRemaining = m.pauseRemaining
		timerColor = pauseColor
		totalDuration = time.Duration(m.pauseMinutes) * time.Minute
	} else if m.state == standView {
		timerRemaining = m.standRemaining
		timerColor = standColor
		totalDuration = time.Duration(m.standMinutes) * time.Minute
	} else {
		timerRemaining = m.sitRemaining
		timerColor = sitColor
		totalDuration = time.Duration(m.sitMinutes) * time.Minute
	}

	// Calculate progress percentage
	if totalDuration.Seconds() > 0 {
		progressPercent = 1.0 - (timerRemaining.Seconds() / totalDuration.Seconds())
	} else {
		progressPercent = 0
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
	var labels []string
	if m.modeType == modeWorkPause {
		labels = []string{"Work", "Pause", "Auto Mode"}
	} else {
		labels = []string{"Stand", "Sit", "Auto Mode"}
	}

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

	var title string
	if m.modeType == modeWorkPause {
		title = titleStyle.Render("Pomo-Settings")
	} else {
		title = titleStyle.Render("Desk-Mode Settings")
	}

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
		helpText = "[SPACE] TimerToggle, [T]oggle Mode, [S]ettings, [R]eset, [Q]uit"
	case pauseView:
		helpText = "[SPACE] TimerToggle, [T]oggle Mode, [S]ettings, [R]eset, [Q]uit"
	case standView:
		helpText = "[SPACE] TimerToggle, [T]oggle Mode, [S]ettings, [R]eset, [Q]uit"
	case sitView:
		helpText = "[SPACE] TimerToggle, [T]oggle Mode, [S]ettings, [R]eset, [Q]uit"
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
	case standView:
		return "- " + standTitles[rand.Intn(len(standTitles))] + " -"
	case sitView:
		return "- " + sitTitles[rand.Intn(len(sitTitles))] + " -"
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
