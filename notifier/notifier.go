package notifier

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

const replaceID = "991199"

var (
	checkOnce sync.Once
	enabled   bool
)

func isEnabled() bool {
	checkOnce.Do(func() {
		if runtime.GOOS != "linux" {
			return
		}
		_, err := exec.LookPath("notify-send")
		enabled = err == nil
	})
	return enabled
}

func modeText(shuffle, repeat bool) string {
	parts := []string{}
	if shuffle {
		parts = append(parts, "Shuffle")
	}
	if repeat {
		parts = append(parts, "Repeat")
	}
	if len(parts) == 0 {
		return "Normal"
	}
	return strings.Join(parts, " + ")
}

func run(title, body, icon string) {
	if !isEnabled() {
		return
	}

	args := []string{"-a", "dicesong", "-r", replaceID, "-u", "low"}
	if icon != "" {
		args = append(args, "-i", icon)
	}
	args = append(args, title, body)

	cmd := exec.Command("notify-send", args...)
	_ = cmd.Run()
}

func NowPlaying(song string, shuffle, repeat bool) {
	body := fmt.Sprintf("%s\nMode: %s", song, modeText(shuffle, repeat))
	run("Now Playing", body, "media-playback-start")
}

func Playback(song string, paused bool) {
	if paused {
		run("Paused", song, "media-playback-pause")
		return
	}
	run("Resumed", song, "media-playback-start")
}

func Error(message string) {
	run("Playback Error", message, "dialog-error")
}
