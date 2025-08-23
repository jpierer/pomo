package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"time"

	_ "embed"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const (
	workView int = iota
	pauseView
	settingView
	quitView
)

var (
	primaryColor = lg.Color("#F1F5F9")
	secondColor  = lg.Color("#81edf6ff")
	blurColor    = lg.Color("#767676")

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

	digits = map[rune][]string{
		'0': {
			"████████",
			"███  ███",
			"███  ███",
			"███  ███",
			"████████",
		},
		'1': {
			"   ███  ",
			"   ███  ",
			"   ███  ",
			"   ███  ",
			"   ███  ",
		},
		'2': {
			"████████",
			"     ███",
			"████████",
			"███     ",
			"████████",
		},
		'3': {
			"████████",
			"     ███",
			"████████",
			"     ███",
			"████████",
		},
		'4': {
			"███  ███",
			"███  ███",
			"████████",
			"     ███",
			"     ███",
		},
		'5': {
			"████████",
			"███     ",
			"████████",
			"     ███",
			"████████",
		},
		'6': {
			"████████",
			"███     ",
			"████████",
			"███  ███",
			"████████",
		},
		'7': {
			"████████",
			"     ███",
			"     ███",
			"     ███",
			"     ███",
		},
		'8': {
			"████████",
			"███  ███",
			"████████",
			"███  ███",
			"████████",
		},
		'9': {
			"████████",
			"███  ███",
			"████████",
			"     ███",
			"████████",
		},
		':': {
			"     ",
			" ███ ",
			"     ",
			" ███ ",
			"     ",
		},
	}
)

//go:embed blink.mp3
var soundFile []byte

type model struct {
	width             int
	height            int
	workMinutes       int
	pauseMinutes      int
	workRemaining     time.Duration
	pauseRemaining    time.Duration
	state             int
	beforeState       int
	stopped           bool
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
		autoIterateStates: false,
		quitSelected:      0,
		settingInputs:     settingInputs,
		settingsIndex:     0,
	}
	m.modeTitle = GetRandomModeTitle(*m)
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
				}
			case settingView:
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
				PlaySound()
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
				PlaySound()
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

	content := RenderView(m) + "\n" + RenderHelp(m)

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

func RenderView(m model) string {
	switch m.state {
	case workView:
		return RenderTime(m)
	case pauseView:
		return RenderTime(m)
	case settingView:
		return RenderSettings(m)
	case quitView:
		return RenderQuit(m)
	}
	return ""
}

func RenderTime(m model) string {

	var t time.Duration
	if m.state == workView {
		t = m.workRemaining
	} else {
		t = m.pauseRemaining
	}

	min := int(t.Minutes())
	sec := int(t.Seconds()) % 60

	// Format time as MM:SS
	timeStr := fmt.Sprintf("%02d:%02d", min, sec)

	// Create timer display
	lines := make([]string, 5)
	for i := 0; i < 5; i++ {
		for _, c := range timeStr {
			if art, ok := digits[c]; ok {
				lines[i] += art[i] + " "
			}
		}
	}

	timerDisplay := lines[0] + "\n" + lines[1] + "\n" + lines[2] + "\n" + lines[3] + "\n" + lines[4]

	titleStyle := lg.NewStyle().
		Foreground(secondColor).
		Bold(true).
		Align(lg.Center)

	modeDisplay := titleStyle.Render(m.modeTitle)

	fullDisplay := timerDisplay + "\n\n" + modeDisplay + "\n\n"

	timerStyle := lg.NewStyle().
		Foreground(primaryColor).
		Align(lg.Center)

	return timerStyle.Render(fullDisplay)
}

func RenderSettings(m model) string {
	labels := []string{"Work", "Pause", "Auto Mode"}

	focusStyle := lg.NewStyle().
		Border(lg.NormalBorder()).
		BorderForeground(secondColor).
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
		Foreground(primaryColor).
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
		Foreground(secondColor)

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
		Foreground(secondColor).
		Bold(true).
		Align(lg.Center).
		Margin(0, 0, 1, 0)

	title := titleStyle.Render("Pomo-Settings")

	settingsStyle := lg.NewStyle().
		Foreground(primaryColor).
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
		Foreground(primaryColor).
		Bold(true).
		Align(lg.Center)

	noStyle := lg.NewStyle().Foreground(blurColor)
	yesStyle := lg.NewStyle().Foreground(blurColor)

	if m.quitSelected == 0 {
		noStyle = lg.NewStyle().Foreground(primaryColor).Bold(true)
	} else {
		yesStyle = lg.NewStyle().Foreground(primaryColor).Bold(true)
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

func PlaySound() {
	// Run sound playback in a goroutine to avoid blocking the UI
	go func() {
		reader := bytes.NewReader(soundFile)

		streamer, format, err := mp3.Decode(io.NopCloser(reader))
		if err != nil {
			log.Printf("Error decoding MP3: %v", err)
			return
		}
		defer streamer.Close()

		// Note: speaker.Init can only be called once, so we handle the case where it's already initialized
		if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
			// Speaker might already be initialized, which is fine
			log.Printf("Speaker init warning (might already be initialized): %v", err)
		}

		speaker.Play(streamer)
	}()
}
