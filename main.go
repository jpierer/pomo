package main

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
)

func main() {
	p := tea.NewProgram(NewModel())
	_, err := p.Run()
	if err != nil {
		log.Fatalf("Error to run pomo %+v", err)
	}

}

const (
	workView uint = iota
	pauseView
	settingView
	quitView
)

type model struct {
	workMinutes       uint
	pauseMinutes      uint
	workRemaining     time.Duration
	pauseRemaining    time.Duration
	state             uint
	beforeState       uint
	stopped           bool
	autoIterateStates bool
	quitSelected      int // 0 = No, 1 = Yes
}

func NewModel() model {
	workMinutes := uint(10)
	pauseMinutes := uint(5)
	return model{
		workMinutes:       workMinutes,
		workRemaining:     time.Duration(workMinutes) * time.Second, // todo auf Minutes umstellen
		pauseMinutes:      pauseMinutes,
		pauseRemaining:    time.Duration(pauseMinutes) * time.Second,
		stopped:           true,
		autoIterateStates: false,
		quitSelected:      0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	// on key press
	case tea.KeyMsg:

		switch msg.String() {

		// Quit View
		case "ctrl+c", "q":
			// Open confirm
			m.quitSelected = 0 // Reset to "No"
			m.SwitchState(quitView)

		// Navigation in quit view
		case "left":
			if m.state == quitView {
				m.quitSelected = 0 // No
			}

		case "right":
			if m.state == quitView {
				m.quitSelected = 1 // Yes
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
				// todo save settings
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
	content := RenderView(m) + "\n" + RenderHelp(m)

	viewStyle := lg.NewStyle().
		Padding(2, 4, 0, 4).
		Border(lg.RoundedBorder()).
		BorderForeground(lg.Color("#ecd10aff")).
		Align(lg.Center).
		Width(80)

	return viewStyle.Render(content)
}

func (m *model) SwitchState(newState uint) {
	m.beforeState = m.state
	m.state = newState
}

func (m *model) ResetAndStop() {
	m.workRemaining = time.Duration(m.workMinutes) * time.Second //  TODO auf Minutes umstellen
	m.pauseRemaining = time.Duration(m.pauseMinutes) * time.Second
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

	digits := map[rune][]string{
		'0': {
			" ██████ ",
			"██    ██",
			"██    ██",
			"██    ██",
			" ██████ ",
		},
		'1': {
			"   ██   ",
			"  ███   ",
			"   ██   ",
			"   ██   ",
			" ██████ ",
		},
		'2': {
			" ██████ ",
			"      ██",
			" ██████ ",
			"██      ",
			"████████",
		},
		'3': {
			" ██████ ",
			"      ██",
			"  █████ ",
			"      ██",
			" ██████ ",
		},
		'4': {
			"██    ██",
			"██    ██",
			"████████",
			"      ██",
			"      ██",
		},
		'5': {
			"████████",
			"██      ",
			"███████ ",
			"      ██",
			"███████ ",
		},
		'6': {
			" ██████ ",
			"██      ",
			"███████ ",
			"██    ██",
			" ██████ ",
		},
		'7': {
			"████████",
			"      ██",
			"     ██ ",
			"    ██  ",
			"   ██   ",
		},
		'8': {
			" ██████ ",
			"██    ██",
			" ██████ ",
			"██    ██",
			" ██████ ",
		},
		'9': {
			" ██████ ",
			"██    ██",
			" ███████",
			"      ██",
			" ██████ ",
		},
		':': {
			"        ",
			"   ██   ",
			"        ",
			"   ██   ",
			"        ",
		},
	}

	var t time.Duration
	var modeTitle string
	if m.state == workView {
		t = m.workRemaining
		modeTitle = " - Work Mode - "
	} else {
		t = m.pauseRemaining
		modeTitle = " - Pause Mode - "
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
				lines[i] += art[i] + "  "
			}
		}
	}

	timerDisplay := lines[0] + "\n" + lines[1] + "\n" + lines[2] + "\n" + lines[3] + "\n" + lines[4]

	titleStyle := lg.NewStyle().
		Foreground(lg.Color("#FFFFFF")).
		Bold(true).
		Align(lg.Center)

	modeDisplay := titleStyle.Render(modeTitle)

	fullDisplay := timerDisplay + "\n\n" + modeDisplay + "\n\n"

	timerStyle := lg.NewStyle().
		Foreground(lg.Color("#FFFFFF")).
		Align(lg.Center)

	return timerStyle.Render(fullDisplay)
}

func RenderSettings(m model) string {
	settingsStyle := lg.NewStyle().
		Foreground(lg.Color("#666666")).
		Faint(true).
		Padding(1, 2).
		Align(lg.Center)

	return settingsStyle.Render("Settings - TODO\n\n")
}

func RenderHelp(m model) string {
	helpStyle := lg.NewStyle().
		Foreground(lg.Color("#888888")).
		Faint(true).
		Padding(0, 0).
		Align(lg.Center)

	var helpText string
	switch m.state {
	case workView:
		helpText = "[SPACE] Start/Pause  [P] PauseMode  [S] Settings  [R] Reset  [Q] Quit"
	case pauseView:
		helpText = "[SPACE] Start/Pause  [W] WorkMode  [S] Settings  [R] Reset  [Q] Quit"
	case settingView:
		helpText = "[ENTER] Save & Exit"
	}

	return helpStyle.Render(helpText)
}

func RenderQuit(m model) string {
	quitStyle := lg.NewStyle().
		Foreground(lg.Color("#fff")).
		Bold(true).
		Align(lg.Center)

	noStyle := lg.NewStyle().Foreground(lg.Color("#888888"))
	yesStyle := lg.NewStyle().Foreground(lg.Color("#888888"))

	if m.quitSelected == 0 {
		noStyle = lg.NewStyle().Foreground(lg.Color("#fff")).Bold(true)
	} else {
		yesStyle = lg.NewStyle().Foreground(lg.Color("#fff")).Bold(true)
	}

	noOption := noStyle.Render("[ No ]")
	yesOption := yesStyle.Render("[ Yes ]")

	content := "Really quit?\n\n" + noOption + "    " + yesOption + "\n\n" +
		lg.NewStyle().Foreground(lg.Color("#666666")).Faint(true).Render("← → to select, ENTER to confirm")

	return quitStyle.Render(content)
}
