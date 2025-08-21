package main

import (
	"fmt"
	"log"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
			if m.state == settingView {
				m.SwitchState(workView)
				return m, nil
			}

			// Open confirm
			m.SwitchState(quitView)

		// Confirm quit
		case "y":
			if m.state == quitView {
				return m, tea.Quit
			}

		// Cancel quit
		case "n":
			if m.state == quitView {
				m.state = m.beforeState
			}

		// Open Setting
		case "s":
			if m.state != settingView {
				m.stopped = true
				m.SwitchState(settingView)
			}

		// Save Setting
		case "enter":
			if m.state == settingView {
				// todo save settings
				m.SwitchState(workView)
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

	s := ""
	s += Render(m)
	s += RenderHelp(m)

	return s
}

func (m *model) SwitchState(newState uint) {
	m.beforeState = m.state
	m.state = newState
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return t
	})
}

func (m *model) ResetAndStop() {
	m.workRemaining = time.Duration(m.workMinutes) * time.Second //  TODO auf Minutes umstellen
	m.pauseRemaining = time.Duration(m.pauseMinutes) * time.Second
	m.stopped = true
}

func Render(m model) string {
	switch m.state {
	case workView:
		return RenderTime(m.workRemaining)
	case pauseView:
		return RenderTime(m.pauseRemaining)
	case settingView:
		return RenderSettings(m)
	case quitView:
		return "Really quit? [Y] Yes  [N] No"
	}
	return ""
}

func RenderTime(t time.Duration) string {

	digits := map[rune][]string{
		'0': {
			" ███ ",
			"█   █",
			"█   █",
			"█   █",
			" ███ ",
		},
		'1': {
			"  █  ",
			" ██  ",
			"  █  ",
			"  █  ",
			" ███ ",
		},
		'2': {
			" ███ ",
			"    █",
			" ███ ",
			"█    ",
			"█████",
		},
		'3': {
			"████ ",
			"    █",
			" ███ ",
			"    █",
			"████ ",
		},
		'4': {
			"█  █ ",
			"█  █ ",
			"█████",
			"   █ ",
			"   █ ",
		},
		'5': {
			"█████",
			"█    ",
			"████ ",
			"    █",
			"████ ",
		},
		'6': {
			" ███ ",
			"█    ",
			"████ ",
			"█   █",
			" ███ ",
		},
		'7': {
			"█████",
			"    █",
			"   █ ",
			"  █  ",
			"  █  ",
		},
		'8': {
			" ███ ",
			"█   █",
			" ███ ",
			"█   █",
			" ███ ",
		},
		'9': {
			" ███ ",
			"█   █",
			" ████",
			"    █",
			" ███ ",
		},
		':': {
			"     ",
			"  █  ",
			"     ",
			"  █  ",
			"     ",
		},
	}

	min := int(t.Minutes())
	sec := int(t.Seconds()) % 60

	// Format time as MM:SS
	timeStr := fmt.Sprintf("%02d:%02d", min, sec)

	// Map ASCII art to time format
	lines := make([]string, 5) // height of ascii digits
	for i := 0; i < 5; i++ {
		for _, c := range timeStr {
			if art, ok := digits[c]; ok {
				lines[i] += art[i] + "  "
			}
		}
	}
	return "\n" + lines[0] + "\n" + lines[1] + "\n" + lines[2] + "\n" + lines[3] + "\n" + lines[4] + "\n"
}

func RenderSettings(m model) string {
	return "Settings" // Work, Pause, AutoIterate
}

func RenderHelp(m model) string {
	s := "\n\n"
	switch m.state {
	case workView:
		s += "[SPACE] Timer Start/Pause  [P] PauseMode  [S] Settings  [R] Reset  [Q] Quit"
	case pauseView:
		s += "[SPACE] Timer Start/Pause  [W] WorkMode  [S] Settings  [R] Reset  [Q] Quit"
	case settingView:
		s += "[ENTER] Save  [Q] Cancel"
	}
	return s
}
