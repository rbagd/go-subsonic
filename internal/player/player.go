//go:build !noaudio

package player

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

type Player struct {
	sampleRate beep.SampleRate
	ctrl       *beep.Ctrl
	streamer   beep.StreamSeekCloser
	volume     *effects.Volume

	// Playback completion
	doneMu    sync.Mutex
	done      chan struct{}
	doneClose func()

	// Progress tracking
	progressStreamer *progressStreamer
}

type progressStreamer struct {
	Streamer beep.Streamer
	Samples  int64 // Atomic
}

func (ps *progressStreamer) Stream(samples [][2]float64) (n int, ok bool) {
	n, ok = ps.Streamer.Stream(samples)
	if n > 0 {
		atomic.AddInt64(&ps.Samples, int64(n))
	}
	return n, ok
}

func (ps *progressStreamer) Err() error {
	return ps.Streamer.Err()
}

func NewPlayer() (*Player, error) {
	sr := beep.SampleRate(44100)
	err := speaker.Init(sr, sr.N(time.Second/10))
	if err != nil {
		return nil, err
	}

	return &Player{
		sampleRate: sr,
		volume:     &effects.Volume{Base: 2, Volume: 0, Silent: false},
	}, nil
}

func (p *Player) Play(r io.ReadCloser) error {
	if r == nil {
		return errors.New("nil stream")
	}

	streamer, format, err := mp3.Decode(r)
	if err != nil {
		_ = r.Close()
		return err
	}

	resampled := beep.Resample(4, format.SampleRate, p.sampleRate, streamer)

	p.Stop() // Stop current playback if any

	done := make(chan struct{})
	var doneOnce sync.Once
	closeDone := func() {
		doneOnce.Do(func() {
			close(done)
		})
	}

	p.doneMu.Lock()
	p.done = done
	p.doneClose = closeDone
	p.doneMu.Unlock()

	// Wrap in progress tracker
	p.progressStreamer = &progressStreamer{Streamer: resampled}

	seq := beep.Seq(
		p.progressStreamer,
		beep.Callback(func() {
			closeDone()
			_ = streamer.Close()
		}),
	)

	p.ctrl = &beep.Ctrl{Streamer: seq, Paused: false}
	p.volume.Streamer = p.ctrl
	p.streamer = streamer

	speaker.Play(p.volume)

	return nil
}

func (p *Player) Done() <-chan struct{} {
	p.doneMu.Lock()
	defer p.doneMu.Unlock()
	if p.done == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return p.done
}

func (p *Player) Position() time.Duration {
	if p.progressStreamer == nil {
		return 0
	}
	samples := atomic.LoadInt64(&p.progressStreamer.Samples)
	return p.sampleRate.D(int(samples))
}

func (p *Player) Pause() {
	if p.ctrl != nil {
		speaker.Lock()
		p.ctrl.Paused = !p.ctrl.Paused
		speaker.Unlock()
	}
}

func (p *Player) Stop() {
	p.doneMu.Lock()
	closeDone := p.doneClose
	p.doneClose = nil
	p.doneMu.Unlock()

	if closeDone != nil {
		closeDone()
	}

	if p.streamer != nil {
		speaker.Clear()
		_ = p.streamer.Close()
		p.streamer = nil
		p.ctrl = nil
	}
}

func (p *Player) SetVolume(vol float64) {
	speaker.Lock()
	p.volume.Volume = vol
	speaker.Unlock()
}
