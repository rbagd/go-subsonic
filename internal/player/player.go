package player

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

type Player struct {
	SampleRate beep.SampleRate
	Ctrl       *beep.Ctrl
	Streamer   beep.StreamSeekCloser
	Volume     *effects.Volume
	
	// Progress tracking
	progressStreamer *ProgressStreamer
}

type ProgressStreamer struct {
	Streamer beep.Streamer
	Samples  int64 // Atomic
}

func (ps *ProgressStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = ps.Streamer.Stream(samples)
	if n > 0 {
		atomic.AddInt64(&ps.Samples, int64(n))
	}
	return n, ok
}

func (ps *ProgressStreamer) Err() error {
	return ps.Streamer.Err()
}

func NewPlayer() (*Player, error) {
	sr := beep.SampleRate(44100)
	err := speaker.Init(sr, sr.N(time.Second/10))
	if err != nil {
		return nil, err
	}

	return &Player{
		SampleRate: sr,
		Volume:     &effects.Volume{Base: 2, Volume: 0, Silent: false},
	}, nil
}

func (p *Player) PlayStream(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http error: %s", resp.Status)
	}

	streamer, format, err := mp3.Decode(resp.Body)
	if err != nil {
		return err
	}

	resampled := beep.Resample(4, format.SampleRate, p.SampleRate, streamer)

	p.Stop() // Stop current playback if any

	// Wrap in progress tracker
	p.progressStreamer = &ProgressStreamer{Streamer: resampled}

	p.Ctrl = &beep.Ctrl{Streamer: p.progressStreamer, Paused: false}
	p.Volume.Streamer = p.Ctrl
	p.Streamer = streamer

	speaker.Play(p.Volume)

	return nil
}

func (p *Player) Position() time.Duration {
	if p.progressStreamer == nil {
		return 0
	}
	samples := atomic.LoadInt64(&p.progressStreamer.Samples)
	return p.SampleRate.D(int(samples))
}

func (p *Player) Pause() {
	if p.Ctrl != nil {
		speaker.Lock()
		p.Ctrl.Paused = !p.Ctrl.Paused
		speaker.Unlock()
	}
}

func (p *Player) Stop() {
	if p.Streamer != nil {
		speaker.Clear()
		p.Streamer.Close()
		p.Streamer = nil
		p.Ctrl = nil
	}
}

func (p *Player) SetVolume(vol float64) {
	speaker.Lock()
	p.Volume.Volume = vol
	speaker.Unlock()
}
