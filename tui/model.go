package tui

import (
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Gylmynnn/dicesong/player"
	"github.com/Gylmynnn/dicesong/state"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	everblushBg0    = lipgloss.Color("#141b1e")
	everblushRed    = lipgloss.Color("#e57474")
	everblushGreen  = lipgloss.Color("#8ccf7e")
	everblushYellow = lipgloss.Color("#e5c76b")
	// everblushBlue   = lipgloss.Color("#67b0e8")
	// everblushCyan   = lipgloss.Color("#6cbfbf")
	everblushFg     = lipgloss.Color("#dadada")
	everblushGray   = lipgloss.Color("#5c6a72")
)

type (
	tickMsg         struct{}
	songFinishedMsg struct{}
	songLoadedMsg   struct{ success bool }
)

type fsEntry struct {
	name  string
	path  string
	isDir bool
}

func listenForFinished(c chan bool) tea.Cmd {
	return func() tea.Msg {
		<-c
		return songFinishedMsg{}
	}
}

func listenForLoaded(c chan bool) tea.Cmd {
	return func() tea.Msg {
		ok := <-c
		return songLoadedMsg{success: ok}
	}
}

type Model struct {
	width       int
	height      int
	errorMsg    string
	musicRoot   string
	currentPath string
	entries     []fsEntry
	allSongs    []string
	cursor      int
	offset      int
	playingIndex int
	loading     bool
	DoneChan    chan bool
	LoadedChan  chan bool
	PlayRequest chan string
	repeat      bool
	shuffle     bool
	lastPlay    time.Time
	progress    float64
	total       float64
}

func InitialModel() Model {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	home, err := os.UserHomeDir()
	if err != nil {
		panic("Gagal mendapatkan direktori home: " + err.Error())
	}

	musicRoot := filepath.Join(home, "Music")

	entries, _ := readDir(musicRoot)
	allSongs, _ := loadAllSongs(musicRoot)
	stateData := state.Load()

	return Model{
		musicRoot:      musicRoot,
		currentPath:    musicRoot,
		entries:        entries,
		allSongs:       allSongs,
		cursor:         0,
		playingIndex:   -1,
		loading:        false,
		repeat:         stateData.Repeat,
		shuffle:        stateData.Shuffle,
		DoneChan:       make(chan bool),
		LoadedChan:     make(chan bool),
		PlayRequest:    make(chan string, 1),
		progress:       0,
		total:          1,
	}
}

func PlaybackManager(playRequest <-chan string, doneChan chan bool, loadedChan chan<- bool) {
	for path := range playRequest {
		err := player.PlayMusic(path, doneChan)
		if err != nil {
			loadedChan <- false
			continue
		} else {
			loadedChan <- true
		}
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg { return tickMsg{} }),
		listenForFinished(m.DoneChan),
		listenForLoaded(m.LoadedChan),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				if m.cursor < m.offset {
					m.offset--
				}
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
				visibleRow := m.height - 12
				if m.cursor >= m.offset+visibleRow {
					m.offset++
				}
			}
		case "right", "l":
			selectedEntry := m.entries[m.cursor]
			if selectedEntry.isDir {
				m.currentPath = selectedEntry.path
				m.entries, _ = readDir(m.currentPath)
				m.cursor = 0
				m.offset = 0
			}
		case "enter":
			if m.loading || time.Since(m.lastPlay) < 300*time.Millisecond {
				break
			}
			selectedEntry := m.entries[m.cursor]
			if selectedEntry.isDir {
				m.currentPath = selectedEntry.path
				m.entries, _ = readDir(m.currentPath)
				m.cursor = 0
				m.offset = 0
			} else {
				m.lastPlay = time.Now()
				m.loading = true
				m.playingIndex = findSongIndex(m.allSongs, selectedEntry.path)
				m.PlayRequest <- selectedEntry.path
				saveState(m)
			}
		case "backspace", "left", "h":
			parentDir := filepath.Dir(m.currentPath)
			if parentDir != "." && parentDir != "" && strings.HasPrefix(m.currentPath, m.musicRoot) && m.currentPath != m.musicRoot {
				m.currentPath = parentDir
				m.entries, _ = readDir(m.currentPath)
				m.cursor = 0
				m.offset = 0
			}
		case "p":
			player.TogglePause()
		case "n":
			if m.playingIndex < len(m.allSongs)-1 && !m.loading {
				m.loading = true
				m.playingIndex++
				m.PlayRequest <- m.allSongs[m.playingIndex]
				saveState(m)
			}
		case "b":
			if m.playingIndex > 0 && !m.loading {
				m.loading = true
				m.playingIndex--
				m.PlayRequest <- m.allSongs[m.playingIndex]
				saveState(m)
			}
		case "r":
			m.repeat = !m.repeat
			saveState(m)
		case "s":
			m.shuffle = !m.shuffle
			saveState(m)
		}

	case tickMsg:
		if m.playingIndex != -1 && !m.loading && !player.IsPaused() {
			m.progress, m.total = player.GetProgress()
		}
		return m, tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg { return tickMsg{} })

	case songLoadedMsg:
		m.loading = false
		if !msg.success {
			m.errorMsg = "Error playing music : " + filepath.Base(m.allSongs[m.playingIndex])
			return m, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
				return songFinishedMsg{}
			})
		} else {
			m.errorMsg = ""
		}
		return m, listenForLoaded(m.LoadedChan)

	case songFinishedMsg:
		m.progress, m.total = 0, 1
		m.loading = false
		m.errorMsg = ""

		if m.repeat && m.playingIndex != -1 {
			m.PlayRequest <- m.allSongs[m.playingIndex]
			return m, tea.Batch(
				listenForFinished(m.DoneChan),
				listenForLoaded(m.LoadedChan),
			)
		} else if m.shuffle && len(m.allSongs) > 0 {
			m.playingIndex = rand.Intn(len(m.allSongs))
			m.PlayRequest <- m.allSongs[m.playingIndex]
			return m, tea.Batch(
				listenForFinished(m.DoneChan),
				listenForLoaded(m.LoadedChan),
			)
		} else {
			if m.playingIndex < len(m.allSongs)-1 {
				m.playingIndex++
				m.PlayRequest <- m.allSongs[m.playingIndex]
				return m, tea.Batch(
					listenForFinished(m.DoneChan),
					listenForLoaded(m.LoadedChan),
				)
			} else {
				m.playingIndex = -1
				return m, nil

			}
		}
	}
	return m, cmd
}

func (m Model) View() string {
	// Header
	header := HeaderStyle.Render("   Dicesong ")
	// File Browser
	var browserContent strings.Builder
	pathHeader := PathHeaderStyle.Render(m.currentPath)
	browserContent.WriteString(header)
	browserContent.WriteString("\n")
	browserContent.WriteString(pathHeader)
	browserContent.WriteString("\n")

	visibleRows := m.height - 8 // Adjusted for header and footer
	end := min(m.offset+visibleRows, len(m.entries))

	for i := m.offset; i < end; i++ {
		entry := m.entries[i]
		icon := "🎵"
		if entry.isDir {
			icon = "📁"
		}

		isPlaying := !entry.isDir && m.playingIndex != -1 && m.allSongs[m.playingIndex] == entry.path
		
		// Responsive display name
		maxWidth := m.width - 4 // Leave space for cursor, icon, etc.
		displayName := fmt.Sprintf("%s %s", icon, entry.name)
		if len(displayName) > maxWidth {
			displayName = displayName[:maxWidth-3] + "..."
		}

		var line string
		if i == m.cursor {
			if isPlaying {
				line = PlayingCursorStyle.Render("▶ " + displayName)
			} else {
				line = CursorStyle.Render("› " + displayName)
			}
		} else {
			if isPlaying {
				line = PlayingStyle.Render("  " + displayName)
			} else {
				line = NormalStyle.Render("  " + displayName)
			}
		}
		browserContent.WriteString(line + "\n")
	}

	// Footer
	var footer strings.Builder

	// Now Playing
	var nowPlayingContent strings.Builder
	if m.loading {
		nowPlayingContent.WriteString(LoadingStyle.Render("◌ Loading..."))
	} else if m.errorMsg != "" {
		nowPlayingContent.WriteString(lipgloss.NewStyle().Foreground(everblushRed).Bold(true).Render(" " + m.errorMsg))
	} else if m.playingIndex != -1 {
		song := filepath.Base(m.allSongs[m.playingIndex])
		nowPlaying := NowPlayingStyle.Render("Now Playing: ") + song
		nowPlayingContent.WriteString(nowPlaying)
	} else {
		nowPlayingContent.WriteString(NowPlayingStyle.Render("Select a song to play."))
	}
	footer.WriteString(nowPlayingContent.String() + "\n")

	// Progress Bar
	if m.playingIndex != -1 {
		bar, percent := renderProgress(m.progress, m.total, m.width)
		progress := lipgloss.JoinHorizontal(lipgloss.Left, bar, percent)
		footer.WriteString(progress + "\n")
	}

	// Controls
	// separator := lipgloss.NewStyle().Foreground(everblushGray).Render(" | ")
	repeatStr := boolStyle(m.repeat, "  ")
	shuffleStr := boolStyle(m.shuffle, "  ")
	pauseStr := MenuStyle.Render("  ")
	backStr := MenuStyle.Render("  ")
	nextStr := MenuStyle.Render("  ")
	quitStr := MenuStyle.Render("  ")

	controls := lipgloss.JoinHorizontal(lipgloss.Left,
		pauseStr,
		// separator,
		backStr,
		// separator,
		nextStr,
		// separator,
		repeatStr,
		// separator,
		shuffleStr,
	)
	
	// Align controls and quit button
	controlsWidth := lipgloss.Width(controls)
	quitWidth := lipgloss.Width(quitStr)
	spacer := lipgloss.NewStyle().Width(m.width - controlsWidth - quitWidth - 4).Render("")

	footer.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, controls, spacer, quitStr))


	// Final Assembly
	browser := lipgloss.NewStyle().Padding(0, 2).Render(browserContent.String())
	finalFooter := lipgloss.NewStyle().Padding(0, 2).Render(footer.String())

	return lipgloss.JoinVertical(lipgloss.Left,
		// header,
		browser,
		finalFooter,
	)
}

func saveState(m Model) {
	state.Save(state.AppState{
		CurrentSong: m.playingIndex,
		Repeat:      m.repeat,
		Shuffle:     m.shuffle,
	})
}

func readDir(path string) ([]fsEntry, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var entries []fsEntry
	for _, file := range files {
		isDir := file.IsDir()
		isMusicFile := strings.HasSuffix(file.Name(), ".mp3") || strings.HasSuffix(file.Name(), ".wav")
		if isDir || isMusicFile {
			entries = append(entries, fsEntry{
				name:  file.Name(),
				path:  filepath.Join(path, file.Name()),
				isDir: isDir,
			})
		}
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].isDir != entries[j].isDir {
			return entries[i].isDir
		}
		return entries[i].name < entries[j].name
	})

	return entries, nil
}

func loadAllSongs(root string) ([]string, error) {
	var songs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (strings.HasSuffix(d.Name(), ".mp3") || strings.HasSuffix(d.Name(), ".wav")) {
			songs = append(songs, path)
		}
		return nil
	})
	return songs, err
}

func findSongIndex(songs []string, path string) int {
	for i, s := range songs {
		if s == path {
			return i
		}
	}
	return -1
}

var (
	PaneStyle              = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(everblushGray).Padding(1, 2)
	PathHeaderStyle        = lipgloss.NewStyle().Foreground(everblushYellow).Bold(true).Underline(true).MarginBottom(1)
	HeaderStyle            = lipgloss.NewStyle().Foreground(everblushBg0).Background(everblushRed).Bold(true)
	NowPlayingSectionStyle = lipgloss.NewStyle()
	NowPlayingStyle        = lipgloss.NewStyle().Foreground(everblushFg).Bold(true)
	CursorStyle            = lipgloss.NewStyle().Foreground(everblushYellow).Bold(true)
	PlayingStyle           = lipgloss.NewStyle().Foreground(everblushGreen)
	PlayingCursorStyle     = lipgloss.NewStyle().Foreground(everblushGreen).Bold(true)
	NormalStyle            = lipgloss.NewStyle().Foreground(everblushGray)
	MenuStyle              = lipgloss.NewStyle().Foreground(everblushFg).Bold(true)
	FooterStyle            = lipgloss.NewStyle().Foreground(everblushGray).MarginTop(1)
	LoadingStyle           = lipgloss.NewStyle().Foreground(everblushGray).Italic(true)
)

func boolStyle(b bool, text string) string {
	if b {
		return lipgloss.NewStyle().Foreground(everblushGreen).Bold(true).Render(text)
	}
	return lipgloss.NewStyle().Foreground(everblushFg).Bold(true).Render(text)
}

func renderProgress(current, total float64, width int) (string, string) {
	barLength := width - 6// Adjusted for percentage
	if total <= 0 {
		return strings.Repeat(" ", barLength), " 0%"
	}
	percent := current / total
	filled := int(percent * float64(barLength))
	bar := strings.Repeat("─", filled) + "󰧞" + strings.Repeat("─", barLength-filled)
	percentStr := fmt.Sprintf(" %.0f%%", percent*100)
	return lipgloss.NewStyle().Foreground(everblushGreen).Render(bar), lipgloss.NewStyle().Foreground(everblushGray).Render(percentStr)
}

