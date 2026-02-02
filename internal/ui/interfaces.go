package ui

import (
	"context"
	"go-subsonic/internal/subsonic"
	"time"
)

// MusicProvider defines the interface for fetching music data.
type MusicProvider interface {
	GetArtists(ctx context.Context) (*subsonic.ArtistsID3, error)
	GetMusicDirectory(ctx context.Context, id string) (*subsonic.Directory, error)
	GetStreamURL(id string) (string, error)
}

// AudioPlayer defines the interface for audio playback.
type AudioPlayer interface {
	PlayStream(url string) error
	Pause()
	Stop()
	SetVolume(vol float64)
	Position() time.Duration
}
