package main

import (
	"fmt"
	"os"

	"github.com/Gylmynnn/dicesong/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := tui.InitialModel()
	go tui.PlaybackManager(m.PlayRequest, m.DoneChan, m.LoadedChan)
	p := tea.NewProgram(m)
	_, err := p.Run()
	if err != nil {
		fmt.Println("Gagal:", err)
		os.Exit(1)
	}
}
