package core

import (
	"context"
	"io"
	"testing"
	"time"
)

type fakeProvider struct{}

func (fakeProvider) ListArtists(ctx context.Context) ([]Artist, error) { return nil, nil }
func (fakeProvider) ListAlbums(ctx context.Context, artistID string) ([]Album, error) {
	return nil, nil
}
func (fakeProvider) ListRecentlyAddedAlbums(ctx context.Context, size, offset int) ([]Album, error) {
	return nil, nil
}
func (fakeProvider) ListTracks(ctx context.Context, albumID string) ([]Track, error) { return nil, nil }
func (fakeProvider) OpenTrackStream(ctx context.Context, trackID string) (io.ReadCloser, error) {
	return nil, nil
}

type fakePlayer struct{}

func (fakePlayer) Play(r io.ReadCloser) error { return nil }
func (fakePlayer) Pause()                     {}
func (fakePlayer) Stop()                      {}
func (fakePlayer) SetVolume(vol float64)      {}
func (fakePlayer) Position() time.Duration    { return 0 }
func (fakePlayer) Done() <-chan struct{}      { ch := make(chan struct{}); close(ch); return ch }

func TestPortsCompile(t *testing.T) {
	var _ MusicProvider = fakeProvider{}
	var _ AudioPlayer = fakePlayer{}
}
