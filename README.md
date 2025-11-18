# DICESONG

A beautiful terminal-based music player built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea). Browse and play your music collection directly from the terminal with an elegant, keyboard-driven interface.

## Features

- **File Browser**: Navigate through your music library with an intuitive file browser
- **Audio Playback**: Play MP3 and WAV files with smooth audio streaming
- **Playback Controls**: Play, pause, skip, and repeat tracks
- **Playback Modes**: 
  - Repeat mode - Loop the current track
  - Shuffle mode - Randomize playback order
- **Progress Tracking**: Real-time progress bar with timestamps
- **Persistent State**: Remembers your playback settings between sessions
- **Responsive UI**: Adapts to different terminal sizes
- **Everblush Color Scheme**: Beautiful, easy-on-the-eyes color palette

## Prerequisites

- Go 1.24.5 or higher
- Audio support (ALSA on Linux, CoreAudio on macOS, DirectSound on Windows)

## Installation

### Using Make (Linux/macOS)

```bash
# Build and install to ~/.local/bin
make install

# Or just build
make build

# Uninstall
make uninstall

# Clean build artifacts
make clean
```

### Using Install Script (Linux/macOS)

```bash
# Install
./install.sh install

# Build only
./install.sh build

# Uninstall
./install.sh uninstall

# Clean
./install.sh clean
```

### Using Install Script (Windows)

```cmd
REM Install
install.bat install

REM Build only
install.bat build

REM Uninstall
install.bat uninstall

REM Clean
install.bat clean
```

### Manual Installation

```bash
# Build the binary
go build -o dicesong

# Move to your PATH (optional)
mv dicesong ~/.local/bin/
```

## Usage

Simply run the program to start browsing and playing music from your `~/Music` directory:

```bash
dicesong
```

For help information:

```bash
dicesong -h
# or
dicesong --help
```

## Keyboard Shortcuts

### Navigation
- `↑` / `k` - Move cursor up
- `↓` / `j` - Move cursor down
- `→` / `l` - Enter folder
- `←` / `h` - Go back to parent folder
- `Enter` - Play selected song / Enter folder

### Playback Controls
- `p` - Play / Pause
- `n` - Next track
- `b` - Previous track (back)

### Playback Modes
- `r` - Toggle repeat mode
- `s` - Toggle shuffle mode

### General
- `q` - Quit application
- `Ctrl+C` - Force quit

## Configuration

- **Music Directory**: `~/Music` (default)
- **State File**: `./state.json` (stores repeat/shuffle settings and current song)

## Project Structure

```
dicesong/
├── player/         # Audio playback engine
│   └── player.go
├── state/          # State persistence
│   └── state.go
├── tui/            # Terminal UI (Bubble Tea)
│   └── model.go
├── build/          # Build output directory
├── main.go         # Application entry point
├── go.mod          # Go module definition
├── Makefile        # Build automation (Make)
├── install.sh      # Installation script (Unix)
└── install.bat     # Installation script (Windows)
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal layouts
- [Beep](https://github.com/faiface/beep) - Audio playback library
- [ID3v2](https://github.com/bogem/id3v2) - MP3 metadata parsing

## Development

### Building from Source

```bash
# Get dependencies
go mod download

# Build
go build -o dicesong

# Run
./dicesong
```

### Running Directly

```bash
go run main.go
```

## License

This project is open source and available under the MIT License.

## Credits

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm
- [Beep](https://github.com/faiface/beep) audio library
- Everblush color scheme

## Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the issues page.

## Author

Created by [Gylmynnn](https://github.com/Gylmynnn)
