//go:build noaudio

package player

import (
	"io"
	"strings"
	"testing"
)

func TestNoAudioBuildPlayer(t *testing.T) {
	p, err := NewPlayer()
	if err != nil {
		t.Fatalf("NewPlayer() error = %v, want nil", err)
	}
	if p == nil {
		t.Fatalf("NewPlayer() player = nil, want non-nil")
	}

	if err := p.Play(io.NopCloser(strings.NewReader("not-mp3"))); err != ErrAudioDisabled {
		t.Fatalf("Play() error = %v, want %v", err, ErrAudioDisabled)
	}

	// No-op methods should not panic.
	p.Pause()
	p.Stop()
	p.SetVolume(0.5)
	_ = p.Position()
}
