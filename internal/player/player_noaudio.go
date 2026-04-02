//go:build noaudio

package player

import (
	"io"
	"time"
)

type Player struct{}

func NewPlayer() (*Player, error) {
	return &Player{}, nil
}

func (p *Player) Play(r io.ReadCloser) error {
	if r != nil {
		_ = r.Close()
	}
	return ErrAudioDisabled
}

func (p *Player) Position() time.Duration {
	return 0
}

func (p *Player) Done() <-chan struct{} { ch := make(chan struct{}); close(ch); return ch }

func (p *Player) Pause() {}

func (p *Player) Stop() {}

func (p *Player) SetVolume(vol float64) {}
