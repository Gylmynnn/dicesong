package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Gylmynnn/dicesong/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func printHelp() {
	fmt.Print(`
╔════════════════════════════════════════════════════════════════╗
║                    ♪ DICESONG - TUI Music Player               ║
╚════════════════════════════════════════════════════════════════╝

USAGE:
  dicesong [OPTIONS]

OPTIONS:
  -h, --help    Show this help message

KEYBOARD SHORTCUTS:

  Navigation:
    ↑ / k       Move cursor up
    ↓ / j       Move cursor down
    → / l       Enter folder
    ← / h       Go back to parent folder
    Enter       Play selected song / Enter folder

  Playback Controls:
    p           Play / Pause
    n           Next track
    b           Previous track (back)

  Playback Modes:
    r           Toggle repeat mode
    s           Toggle shuffle mode

  General:
    q           Quit application
    Ctrl+C      Force quit

FEATURES:
  • Browse and play MP3 and WAV files
  • Shuffle and repeat modes
  • Progress bar with timestamps
  • Persistent state (remembers last settings)
  • Responsive design for different terminal sizes

Music directory: ~/Music
State file: ./state.json

`)
}

func main() {
	help := flag.Bool("h", false, "Show help message")
	flag.BoolVar(help, "help", false, "Show help message")
	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	m := tui.InitialModel()
	go tui.PlaybackManager(m.PlayRequest, m.DoneChan, m.LoadedChan)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Failed to launch:", err)
		os.Exit(1)
	}
}
