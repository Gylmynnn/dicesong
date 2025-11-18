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
	everblushBg0     = lipgloss.Color("#141b1e")
	everblushBg1     = lipgloss.Color("#232a2d")
	everblushRed     = lipgloss.Color("#e57474")
	everblushGreen   = lipgloss.Color("#8ccf7e")
	everblushYellow  = lipgloss.Color("#e5c76b")
	everblushBlue    = lipgloss.Color("#67b0e8")
	everblushMagenta = lipgloss.Color("#c47fd5")
	everblushCyan    = lipgloss.Color("#6cbfbf")
	everblushFg      = lipgloss.Color("#dadada")
	everblushGray    = lipgloss.Color("#5c6a72")
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
	width        int
	height       int
	errorMsg     string
	musicRoot    string
	currentPath  string
	entries      []fsEntry
	allSongs     []string
	cursor       int
	offset       int
	playingIndex int
	loading      bool
	DoneChan     chan bool
	LoadedChan   chan bool
	PlayRequest  chan string
	repeat       bool
	shuffle      bool
	lastPlay     time.Time
	progress     float64
	total        float64
}

func InitialModel() Model {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	home, err := os.UserHomeDir()
	if err != nil {
		panic("Failed to read home directory: " + err.Error())
	}

	musicRoot := filepath.Join(home, "Music")

	entries, _ := readDir(musicRoot)
	allSongs, _ := loadAllSongs(musicRoot)
	stateData := state.Load()

	return Model{
		musicRoot:    musicRoot,
		currentPath:  musicRoot,
		entries:      entries,
		allSongs:     allSongs,
		cursor:       0,
		playingIndex: -1,
		loading:      false,
		repeat:       stateData.Repeat,
		shuffle:      stateData.Shuffle,
		DoneChan:     make(chan bool),
		LoadedChan:   make(chan bool),
		PlayRequest:  make(chan string, 1),
		progress:     0,
		total:        1,
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
				visibleRow := m.height - 16
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
			m.errorMsg = "Error playing: " + filepath.Base(m.allSongs[m.playingIndex])
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
	if m.width == 0 || m.height == 0 {
		return ""
	}

	headerHeight := 3
	playerBarHeight := 3
	browserHeight := m.height - headerHeight - playerBarHeight

	header := m.renderHeader()
	browser := m.renderBrowser(browserHeight)
	playerBar := m.renderPlayerBar()

	return lipgloss.JoinVertical(lipgloss.Top, header, browser, playerBar)
}

func (m Model) renderHeader() string {
	titleContent := HeaderTitleStyle.Render("    DICESONG  ")

	songCount := HeaderInfoStyle.Render(fmt.Sprintf("%d Songs", len(m.allSongs)))

	titleWidth := lipgloss.Width(titleContent)
	infoWidth := lipgloss.Width(songCount)
	spacerWidth := max(m.width-titleWidth-infoWidth-4, 0)
	spacer := strings.Repeat(" ", spacerWidth)

	headerLine := lipgloss.JoinHorizontal(lipgloss.Left,
		titleContent,
		spacer,
		songCount,
	)

	headerBox := HeaderBoxStyle.Width(m.width).Render(headerLine)
	return headerBox
}

func (m Model) renderBrowser(height int) string {
	var content strings.Builder

	home, _ := os.UserHomeDir()
	displayPath := m.currentPath
	if after, ok := strings.CutPrefix(m.currentPath, home); ok {
		displayPath = "~" + after
	}

	pathLine := BrowserPathStyle.Render("  󱍙 " + displayPath + "  ")
	content.WriteString(pathLine + "\n")
	content.WriteString(BrowserSeparatorStyle.Render(strings.Repeat("─", m.width)) + "\n")

	visibleRows := max(height-3, 1)
	end := min(m.offset+visibleRows, len(m.entries))

	for i := m.offset; i < end; i++ {
		entry := m.entries[i]

		icon := " "
		if entry.isDir {
			icon = "󱍙 "
		}

		isPlaying := !entry.isDir && m.playingIndex != -1 && m.allSongs[m.playingIndex] == entry.path

		maxWidth := m.width - 14
		displayName := entry.name
		if len(displayName) > maxWidth {
			displayName = displayName[:maxWidth-3] + "..."
		}

		var line string
		var cursor string

		if i == m.cursor {
			cursor = "▶"
		} else {
			cursor = " "
		}
		if i == m.cursor {
			if isPlaying {
				line = BrowserItemPlayingSelectedStyle.Render(fmt.Sprintf(" %s %s %s ", cursor, icon, displayName))
			} else if entry.isDir {
				line = BrowserItemDirSelectedStyle.Render(fmt.Sprintf(" %s %s %s ", cursor, icon, displayName))
			} else {
				line = BrowserItemSelectedStyle.Render(fmt.Sprintf(" %s %s %s ", cursor, icon, displayName))
			}
		} else {
			if isPlaying {
				line = BrowserItemPlayingStyle.Render(fmt.Sprintf(" %s %s %s ", cursor, icon, displayName))
			} else if entry.isDir {
				line = BrowserItemDirStyle.Render(fmt.Sprintf(" %s %s %s ", cursor, icon, displayName))
			} else {
				line = BrowserItemStyle.Render(fmt.Sprintf(" %s %s %s ", cursor, icon, displayName))
			}
		}
		content.WriteString(line + "\n")
	}

	return BrowserBoxStyle.
		Width(m.width).
		Height(height).
		Render(content.String())
}

func (m Model) renderPlayerBar() string {
	var lines []string
	borderLine := PlayerBorderStyle.Render(strings.Repeat("━", m.width))
	lines = append(lines, borderLine)
	lines = append(lines, m.renderNowPlaying())
	lines = append(lines, m.renderProgressSection())
	lines = append(lines, m.renderControls())
	return strings.Join(lines, "\n")
}

func (m Model) renderNowPlaying() string {
	var nowPlaying string

	if m.loading {
		icon := NowPlayingIconStyle.Render("◌")
		text := NowPlayingTextStyle.Render(" Loading...")
		nowPlaying = "  " + icon + text
	} else if m.errorMsg != "" {
		icon := NowPlayingErrorIconStyle.Render("✕")
		text := NowPlayingErrorTextStyle.Render(" " + m.errorMsg)
		nowPlaying = "  " + icon + text
	} else if m.playingIndex != -1 {
		song := filepath.Base(m.allSongs[m.playingIndex])

		var playIcon string
		if player.IsPaused() {
			playIcon = NowPlayingIconStyle.Render("⏸")
		} else {
			playIcon = NowPlayingIconStyle.Render("♫")
		}

		label := NowPlayingLabelStyle.Render(" Now Playing: ")
		songName := NowPlayingSongStyle.Render(song)

		nowPlaying = "  " + playIcon + label + songName
	} else {
		icon := NowPlayingIdleIconStyle.Render("♪")
		text := NowPlayingIdleTextStyle.Render(" Ready to play - Select a song")
		nowPlaying = "  " + icon + text
	}

	return nowPlaying
}

func (m Model) renderProgressSection() string {
	if m.playingIndex == -1 || m.loading {
		emptyBar := PlayerProgressEmptyStyle.Render(strings.Repeat("─", m.width-4))
		return "  " + emptyBar
	}

	currentTime := formatDuration(m.progress, m.total)
	totalTime := formatDuration(m.total, m.total)

	barWidth := max(m.width - len(currentTime) - len(totalTime) - 8, 10)
	bar := renderProgressBar(m.progress, m.total, barWidth)

	timeLeft := PlayerTimeStyle.Render(currentTime)
	timeRight := PlayerTimeStyle.Render(totalTime)
	progressLine := "  " + timeLeft + " " + bar + " " + timeRight

	return progressLine
}

func (m Model) renderControls() string {
	var controls []string
	showLabels := m.width >= 80

	if m.playingIndex != -1 && !player.IsPaused() {
		if showLabels {
			pauseBtn := ControlButtonActiveStyle.Render(" Pause")
			controls = append(controls, pauseBtn)
		} else {
			pauseBtn := ControlButtonActiveStyle.Render(" ")
			controls = append(controls, pauseBtn)
		}
	} else {
		if showLabels {
			playBtn := ControlButtonStyle.Render("▶ Play")
			controls = append(controls, playBtn)
		} else {
			playBtn := ControlButtonStyle.Render("▶ ")
			controls = append(controls, playBtn)
		}
	}

	controls = append(controls, "  ")

	if showLabels {
		prevBtn := ControlButtonStyle.Render(" Previous")
		controls = append(controls, prevBtn)
	} else {
		prevBtn := ControlButtonStyle.Render("  ")
		controls = append(controls, prevBtn)
	}

	controls = append(controls, "  ")

	if showLabels {
		nextBtn := ControlButtonStyle.Render("Next  ")
		controls = append(controls, nextBtn)
	} else {
		nextBtn := ControlButtonStyle.Render(" ")
		controls = append(controls, nextBtn)
	}

	if showLabels {
		controls = append(controls, "  ")
	} else {
		controls = append(controls, "  ")
	}

	if m.repeat {
		if showLabels {
			repeatBtn := ControlButtonOnStyle.Render("  Repeat")
			controls = append(controls, repeatBtn)
		} else {
			repeatBtn := ControlButtonOnStyle.Render("  ")
			controls = append(controls, repeatBtn)
		}
	} else {
		if showLabels {
			repeatBtn := ControlButtonOffStyle.Render("  Repeat")
			controls = append(controls, repeatBtn)
		} else {
			repeatBtn := ControlButtonOffStyle.Render("  ")
			controls = append(controls, repeatBtn)
		}
	}

	controls = append(controls, "  ")

	if m.shuffle {
		if showLabels {
			shuffleBtn := ControlButtonOnStyle.Render("  Shuffle")
			controls = append(controls, shuffleBtn)
		} else {
			shuffleBtn := ControlButtonOnStyle.Render("  ")
			controls = append(controls, shuffleBtn)
		}
	} else {
		if showLabels {
			shuffleBtn := ControlButtonOffStyle.Render("  Shuffle")
			controls = append(controls, shuffleBtn)
		} else {
			shuffleBtn := ControlButtonOffStyle.Render("  ")
			controls = append(controls, shuffleBtn)
		}
	}

	controlsLine := lipgloss.JoinHorizontal(lipgloss.Left, controls...)

	controlsWidth := lipgloss.Width(controlsLine)
	leftPadding := max((m.width-controlsWidth)/2, 2)

	return strings.Repeat(" ", leftPadding) + controlsLine
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

func formatDuration(current, total float64) string {
	if total <= 0 {
		return "0:00"
	}
	seconds := max(int(current/44100), 0)

	mins := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%d:%02d", mins, secs)
}

func renderProgressBar(current, total float64, width int) string {
	if total <= 0 || width < 10 {
		return PlayerProgressEmptyStyle.Render(strings.Repeat("─", width))
	}

	percent := current / total
	if percent > 1 {
		percent = 1
	}
	if percent < 0 {
		percent = 0
	}

	filled := min(max(int(percent*float64(width)), 0), width)

	var bar string
	if filled > 0 {
		bar = PlayerProgressFilledStyle.Render(strings.Repeat("━", filled))
	}

	if filled < width {
		bar += PlayerProgressIndicatorStyle.Render("●")
		remaining := width - filled - 1
		if remaining > 0 {
			bar += PlayerProgressEmptyStyle.Render(strings.Repeat("─", remaining))
		}
	} else {
		bar = PlayerProgressFilledStyle.Render(strings.Repeat("━", width-1))
		bar += PlayerProgressIndicatorStyle.Render("●")
	}

	return bar
}

var (
	HeaderBoxStyle   = lipgloss.NewStyle().Background(everblushBg1).BorderStyle(lipgloss.RoundedBorder()).BorderForeground(everblushGray).BorderBottom(true).Padding(0, 1)
	HeaderTitleStyle = lipgloss.NewStyle().Foreground(everblushYellow).Bold(true)
	HeaderInfoStyle  = lipgloss.NewStyle().Italic(true)
)

var (
	BrowserBoxStyle                 = lipgloss.NewStyle().Background(everblushBg0).Padding(0)
	BrowserPathStyle                = lipgloss.NewStyle().Foreground(everblushBlue).Bold(true).Background(everblushBg1)
	BrowserSeparatorStyle           = lipgloss.NewStyle().Foreground(everblushGray)
	BrowserItemStyle                = lipgloss.NewStyle().Foreground(everblushFg)
	BrowserItemDirStyle             = lipgloss.NewStyle().Foreground(everblushCyan).Bold(true)
	BrowserItemSelectedStyle        = lipgloss.NewStyle().Foreground(everblushYellow).Background(everblushBg1).Bold(true)
	BrowserItemDirSelectedStyle     = lipgloss.NewStyle().Foreground(everblushCyan).Background(everblushBg1).Bold(true)
	BrowserItemPlayingStyle         = lipgloss.NewStyle().Foreground(everblushGreen).Bold(true)
	BrowserItemPlayingSelectedStyle = lipgloss.NewStyle().Foreground(everblushGreen).Background(everblushBg1).Bold(true)
)

var (
	PlayerBorderStyle            = lipgloss.NewStyle().Foreground(everblushGray)
	NowPlayingIconStyle          = lipgloss.NewStyle().Foreground(everblushMagenta).Bold(true)
	NowPlayingLabelStyle         = lipgloss.NewStyle().Foreground(everblushFg)
	NowPlayingSongStyle          = lipgloss.NewStyle().Foreground(everblushYellow).Bold(true)
	NowPlayingTextStyle          = lipgloss.NewStyle().Foreground(everblushFg)
	NowPlayingIdleIconStyle      = lipgloss.NewStyle().Foreground(everblushGray)
	NowPlayingIdleTextStyle      = lipgloss.NewStyle().Foreground(everblushGray).Italic(true)
	NowPlayingErrorIconStyle     = lipgloss.NewStyle().Foreground(everblushRed).Bold(true)
	NowPlayingErrorTextStyle     = lipgloss.NewStyle().Foreground(everblushRed)
	PlayerTimeStyle              = lipgloss.NewStyle().Foreground(everblushGray)
	PlayerProgressFilledStyle    = lipgloss.NewStyle().Foreground(everblushGreen).Bold(true)
	PlayerProgressIndicatorStyle = lipgloss.NewStyle().Foreground(everblushYellow).Bold(true)
	PlayerProgressEmptyStyle     = lipgloss.NewStyle().Foreground(everblushGray)
	ControlButtonStyle           = lipgloss.NewStyle().Foreground(everblushFg)
	ControlButtonActiveStyle     = lipgloss.NewStyle().Foreground(everblushBlue).Bold(true)
	ControlButtonOnStyle         = lipgloss.NewStyle().Foreground(everblushGreen).Bold(true)
	ControlButtonOffStyle        = lipgloss.NewStyle().Foreground(everblushGray)
)
