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

## Modes

**Pomo** supports two distinct working modes:

### Work/Pause Mode (default)
The classic Pomodoro technique. Work in focused intervals followed by restorative breaks. Default is 25 minutes work + 5 minutes pause, fully customizable.

### Desk Mode
A health-focused mode designed to encourage movement. Alternate between standing and sitting at your desk. Default is 30 minutes each, fully customizable. Perfect for remote workers and desk-bound professionals.

## Usage

### Start the application

**Work/Pause Mode (default):**
```bash
./pomo
```

**Desk Mode (Stand/Sit):**
```bash
./pomo --mode desk
```

The app saves your settings automatically, so your preferred durations and auto-iterate preference persist across sessions.

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

# Run the app (Work/Pause mode)
./pomo

# Or use Desk mode
./pomo --mode desk
```

Or install directly with Go:

```bash
go install github.com/jpierer/pomo@latest
```

## Features

- **Beautiful TUI Interface** - Clean, modern terminal interface using Bubble Tea
- **Visual Progress Bar** - Real-time progress visualization with color-coded modes
- **Two Operating Modes**:
  - **Work/Pause Mode** - Classic Pomodoro technique (25 min work + 5 min pause, customizable)
  - **Desk Mode** - Health-conscious Stand/Sit reminders (30 min each, customizable)
- **Customizable Timers** - Set custom durations for all timer types (1-60 minutes)
- **Auto-Iterate Mode** - Automatically switch between sessions
- **Desktop Notifications** - System notifications when sessions complete
- **Persistent Settings** - Your preferences are saved automatically
- **Lightweight** - Fast and minimal resource usage
- **Cross-Platform** - Works on macOS, Linux, and Windows
- **Keyboard Shortcuts** - Full keyboard navigation and controls

### Controls

#### Work/Pause Mode (default)
- `SPACE` - Timer Start/Stop
- `T` - Toggle between Work and Pause modes
- `S` - Open settings
- `R` - Reset current timer
- `Q` - Quit application

#### Desk Mode (Stand/Sit)
- `SPACE` - Timer Start/Stop
- `T` - Toggle between Stand and Sit modes
- `S` - Open settings
- `R` - Reset current timer
- `Q` - Quit application

#### Settings
- `← / →` - Navigate between settings fields
- `↑ / ↓` - Adjust values (+/- minutes)
- `SPACE` - Toggle auto-iterate checkbox
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
