package state

import (
	"encoding/json"
	"os"
)

type AppState struct {
	CurrentSong int  `json:"current_song"`
	Repeat      bool `json:"repeat"`
	Shuffle     bool `json:"shuffle"`
}

const stateFile = "state.json"

func Save(state AppState) {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(stateFile, data, 0o644)
}

func Load() AppState {
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return AppState{}
	}
	var state AppState
	_ = json.Unmarshal(data, &state)
	return state
}
