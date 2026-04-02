package core

import (
	"context"
	"io"
	"time"
)

type MusicProvider interface {
	ListArtists(ctx context.Context) ([]Artist, error)
	ListAlbums(ctx context.Context, artistID string) ([]Album, error)
	ListRecentlyAddedAlbums(ctx context.Context, size, offset int) ([]Album, error)
	ListTracks(ctx context.Context, albumID string) ([]Track, error)
	OpenTrackStream(ctx context.Context, trackID string) (io.ReadCloser, error)
}

type AudioPlayer interface {
	Play(r io.ReadCloser) error
	Pause()
	Stop()
	SetVolume(vol float64)
	Position() time.Duration
	// Done is closed when the current playback ends or is stopped.
	Done() <-chan struct{}
}
