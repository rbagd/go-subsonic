package ui

import "github.com/charmbracelet/bubbles/list"

// PlaySongMsg is sent when a song should be played immediately.
type PlaySongMsg struct {
	Item item
}

// AddToPlaylistMsg is sent when items should be added to the playlist.
type AddToPlaylistMsg []list.Item

// SongStartedMsg is sent when a song successfully starts playing.
type SongStartedMsg struct {
	Title    string
	Artist   string
	Album    string
	Year     int
	Track    int
	Duration int
}

// VolumeChangeMsg is sent to change the player volume.
type VolumeChangeMsg struct {
	Delta float64
}

// ErrorMsg is a wrapper for errors to be displayed in the UI.
type ErrorMsg error
