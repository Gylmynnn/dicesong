package player

import (
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
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

	if strings.HasSuffix(path, ".mp3") {
		stream, localFormat, err = mp3.Decode(f)
	} else if strings.HasSuffix(path, ".wav") {
		stream, localFormat, err = wav.Decode(f)
	} else {
		f.Close()
		return errors.New("format tidak didukung")
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
