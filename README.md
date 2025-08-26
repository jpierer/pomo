# Pomo

A beautiful terminal-based Pomodoro Timer written in Go with a sleek TUI interface and progress visualization.

## Screenshots

![Work Mode](img/work.png)
_Work mode_

![Pause Mode](img/pause.png)
_Pause mode_

![Settings](img/settings.png)
_Settings_

## What is the Pomodoro Technique?

The Pomodoro Technique is a time management method that uses a timer to break work into intervals, traditionally 25 minutes in length, separated by short pauses. These intervals are known as "pomodoros", named after the tomato-shaped kitchen timer that Francesco Cirillo used as a university student.

The technique helps improve focus and productivity by creating structured work sessions with regular pauses. After completing four pomodoros, you take a longer pause of 15-30 minutes.

Learn more about the [Pomodoro Technique on Wikipedia](https://en.wikipedia.org/wiki/Pomodoro_Technique).

## Installation

### Requirements

- Go 1.21 or higher
- Terminal with Unicode support

### Install

```bash
# Clone the repository
git clone https://github.com/jpierer/pomo.git

# Navigate to the project directory
cd pomo

# Build the application
go build

# Run the app
./pomo
```

Or install directly with Go:

```bash
go install github.com/jpierer/pomo@latest
```

## Features

- **Beautiful TUI Interface** - Clean, modern terminal interface using Bubble Tea
- **Visual Progress Bar** - Real-time progress visualization with color-coded modes
- **Customizable Timers** - Set custom work and pause durations (1-60 minutes)
- **Auto-Iterate Mode** - Automatically switch between work and pause sessions
- **Desktop Notifications** - System notifications when sessions complete
- **Persistent Settings** - Your preferences are saved automatically
- **Dynamic Mode Titles** - Motivational titles that change with each session
- **Lightweight** - Fast and minimal resource usage
- **Cross-Platform** - Works on macOS, Linux, and Windows
- **Keyboard Shortcuts** - Full keyboard navigation and controls

### Controls

- `SPACE` - Start/Stop timer
- `S` - Open settings
- `R` - Reset current timer
- `W` - Switch to work mode
- `P` - Switch to pause mode
- `Q` - Quit application
- `↑/↓` - Adjust values in settings
- `←/→` - Navigate between settings fields
- `ENTER` - Save settings and return to timer

## Technical Details

Built with modern Go libraries:

- **Bubble Tea** - Terminal UI framework
- **Lipgloss** - Styling and layout
- **Bubbles** - UI components (textinput, progress)
- **Beeep** - Cross-platform desktop notifications

## Support Me

Give a ⭐ if this project was helpful in any way!

## License

The code is released under the [MIT LICENSE](LICENSE).
