package player

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/flac"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
	"github.com/faiface/beep/wav"
)

var (
	ctrl     *beep.Ctrl
	seeker   beep.StreamSeekCloser
	format   beep.Format
	initOnce sync.Once
	mutex    sync.Mutex
)

func InitSpeaker(sampleRate beep.SampleRate) {
	initOnce.Do(func() {
		speaker.Init(sampleRate, sampleRate.N(time.Second/10))
	})
}

func stopCurrent() {
	speaker.Clear()
	if seeker != nil {
		seeker.Close()
		seeker = nil
		ctrl = nil
	}
}

func StopCurrent() {
	mutex.Lock()
	defer mutex.Unlock()
	stopCurrent()
}

func PlayMusic(path string, done chan bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	var stream beep.StreamSeekCloser
	var localFormat beep.Format

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp3":
		stream, localFormat, err = mp3.Decode(f)
	case ".wav":
		stream, localFormat, err = wav.Decode(f)
	case ".flac":
		stream, localFormat, err = flac.Decode(f)
	case ".ogg", ".oga":
		stream, localFormat, err = vorbis.Decode(f)
	default:
		f.Close()
		return errors.New("format tidak didukung (gunakan: mp3, wav, flac, ogg)")
	}

	if err != nil {
		f.Close()
		return err
	}

	mutex.Lock()
	defer mutex.Unlock()
	stopCurrent()
	seeker = stream
	format = localFormat
	InitSpeaker(format.SampleRate)
	ctrl = &beep.Ctrl{Streamer: beep.Seq(stream, beep.Callback(func() {
		done <- true
	})), Paused: false}
	speaker.Play(ctrl)
	return nil
}

func TogglePause() {
	mutex.Lock()
	defer mutex.Unlock()
	if ctrl != nil {
		speaker.Lock()
		ctrl.Paused = !ctrl.Paused
		speaker.Unlock()
	}
}

func IsPaused() bool {
	mutex.Lock()
	defer mutex.Unlock()
	if ctrl == nil {
		return false
	}
	return ctrl.Paused
}

func GetProgress() (float64, float64) {
	mutex.Lock()
	defer mutex.Unlock()
	if seeker == nil || ctrl == nil {
		return 0, 1
	}

	position := seeker.Position()
	length := seeker.Len()
	if length == 0 {
		return 0, 1
	}

	return float64(position), float64(length)
}
