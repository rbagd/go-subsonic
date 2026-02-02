package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type StatusModel struct {
	Player              AudioPlayer
	CurrentSong         string
	CurrentArtist       string
	CurrentAlbum        string
	CurrentYear         int
	CurrentTrack        int
	CurrentSongDuration int
	Volume              float64
	Err                 error
	Width               int
}

func NewStatusModel(player AudioPlayer) StatusModel {
	return StatusModel{
		Player:      player,
		CurrentSong: "None",
		Volume:      0.0,
	}
}

func (m StatusModel) Init() tea.Cmd {
	return nil
}

func (m StatusModel) Update(msg tea.Msg) (StatusModel, tea.Cmd) {
	switch msg := msg.(type) {
	case SongStartedMsg:
		m.CurrentSong = msg.Title
		m.CurrentArtist = msg.Artist
		m.CurrentAlbum = msg.Album
		m.CurrentYear = msg.Year
		m.CurrentTrack = msg.Track
		m.CurrentSongDuration = msg.Duration
		m.Err = nil

	case VolumeChangeMsg:
		m.Volume += msg.Delta
		if m.Volume > 2.0 {
			m.Volume = 2.0
		}
		if m.Volume < -5.0 {
			m.Volume = -5.0
		}
		if m.Player != nil {
			m.Player.SetVolume(m.Volume)
		}

	case ErrorMsg:
		m.Err = msg
	}

	return m, nil
}

func (m StatusModel) View() string {
	currentPos := time.Duration(0)
	if m.Player != nil {
		currentPos = m.Player.Position()
	}
	if m.CurrentSongDuration > 0 && currentPos.Seconds() > float64(m.CurrentSongDuration) {
		currentPos = time.Duration(m.CurrentSongDuration) * time.Second
	}

	totalDur := time.Duration(m.CurrentSongDuration) * time.Second

	// Format:
	// Title (Track X)
	// Artist • Album (Year)
	// Time: 00:00 / 05:00

	trackStr := ""
	if m.CurrentTrack != 0 {
		trackStr = fmt.Sprintf(" (Track %d)", m.CurrentTrack)
	}

	albumStr := m.CurrentAlbum
	if m.CurrentYear != 0 {
		albumStr = fmt.Sprintf("%s (%d)", m.CurrentAlbum, m.CurrentYear)
	}
	
	line1 := fmt.Sprintf("%s%s", m.CurrentSong, trackStr)
	line2 := fmt.Sprintf("%s • %s", m.CurrentArtist, albumStr)
	if m.CurrentArtist == "" && m.CurrentAlbum == "" {
		line2 = "Unknown Artist • Unknown Album"
	}
	line3 := fmt.Sprintf("Time: %s / %s", FmtDuration(currentPos), FmtDuration(totalDur))

	content := fmt.Sprintf("%s\n%s\n%s", line1, line2, line3)

	if m.Err != nil {
		content = fmt.Sprintf("Error: %v", m.Err)
	}

	// Height is now 3 lines of content + padding
	return panelStyle.
		Width(m.Width - 4).
		Height(3).
		Render(content)
}

func (m *StatusModel) SetSize(w int) {
	m.Width = w
}
