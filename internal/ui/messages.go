package ui

import "github.com/charmbracelet/bubbles/list"

// PlaySongMsg is sent when a song should be played immediately.
type PlaySongMsg struct {
	Item             item
	PlaylistIndex    int
	HasPlaylistIndex bool
}

// AddToPlaylistMsg is sent when items should be added to the playlist.
type AddToPlaylistMsg []list.Item

// SongStartedMsg is sent when a song successfully starts playing.
type SongStartedMsg struct {
	PlaybackID       uint64
	TrackID          string
	PlaylistIndex    int
	HasPlaylistIndex bool
	Title            string
	Artist           string
	Album            string
	Year             int
	Track            int
	Duration         int
}

// SongFinishedMsg is sent when the current song finishes (or is stopped).
type SongFinishedMsg struct {
	PlaybackID       uint64
	TrackID          string
	PlaylistIndex    int
	HasPlaylistIndex bool
}

// VolumeChangeMsg is sent to change the player volume.
type VolumeChangeMsg struct {
	Delta float64
}

// ErrorMsg is a wrapper for errors to be displayed in the UI.
type ErrorMsg error

// RandomAlbumSelectedMsg is sent when a random artist and album are selected.
type RandomAlbumSelectedMsg struct {
	ArtistID   string
	ArtistName string
	Albums     []list.Item
	Index      int
}
