package ui

import (
	"context"
	"errors"
	time "time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Focus int

const (
	FocusLibrary Focus = iota
	FocusPlaylist
)

type MainModel struct {
	Client MusicProvider
	Player AudioPlayer

	Focus  Focus
	Width  int
	Height int

	Library  LibraryModel
	Playlist PlaylistModel
	Status   StatusModel

	ShowQuitConfirm bool

	playbackID       uint64
	activePlaybackID uint64
}

func NewModel(client MusicProvider, p AudioPlayer) *MainModel {
	lib := NewLibraryModel(client)
	lib.SetFocused(true)
	pl := NewPlaylistModel()
	pl.SetFocused(false)

	return &MainModel{
		Client:          client,
		Player:          p,
		Focus:           FocusLibrary,
		Library:         lib,
		Playlist:        pl,
		Status:          NewStatusModel(p),
		ShowQuitConfirm: false,
	}
}

type TickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m MainModel) Init() tea.Cmd {
	return tea.Batch(m.Library.Init(), tickCmd())
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.updateSizes()
		// Don't propagate WindowSizeMsg to sub-models as we handle sizing via SetSize
		return m, nil

	case TickMsg:
		return m, tickCmd()

	case PlaySongMsg:
		m.playbackID++
		m.activePlaybackID = m.playbackID
		if msg.HasPlaylistIndex {
			m.Playlist.SetNowPlayingAt(msg.Item.id, msg.PlaylistIndex)
		} else {
			m.Playlist.SetNowPlaying(msg.Item.id)
		}
		return m, m.playSong(msg.Item, m.playbackID)

	case SongStartedMsg:
		m.activePlaybackID = msg.PlaybackID
		if msg.HasPlaylistIndex {
			m.Playlist.SetNowPlayingAt(msg.TrackID, msg.PlaylistIndex)
		} else {
			m.Playlist.SetNowPlaying(msg.TrackID)
		}
		if m.Player != nil {
			playbackID := msg.PlaybackID
			trackID := msg.TrackID
			playlistIndex := msg.PlaylistIndex
			hasPlaylistIndex := msg.HasPlaylistIndex
			cmds = append(cmds, func() tea.Msg {
				<-m.Player.Done()
				return SongFinishedMsg{PlaybackID: playbackID, TrackID: trackID, PlaylistIndex: playlistIndex, HasPlaylistIndex: hasPlaylistIndex}
			})
		}
		m.Status, _ = m.Status.Update(msg)
		m.updateSizes()

	case SongFinishedMsg:
		if msg.PlaybackID != m.activePlaybackID {
			break
		}

		items := m.Playlist.List.Items()
		idx := -1
		if msg.HasPlaylistIndex && msg.PlaylistIndex >= 0 && msg.PlaylistIndex < len(items) {
			if plItem, ok := items[msg.PlaylistIndex].(item); ok && plItem.id == msg.TrackID {
				idx = msg.PlaylistIndex
			}
		}

		if idx == -1 && m.Playlist.NowPlayingIndex >= 0 && m.Playlist.NowPlayingIndex < len(items) {
			if plItem, ok := items[m.Playlist.NowPlayingIndex].(item); ok && plItem.id == msg.TrackID {
				idx = m.Playlist.NowPlayingIndex
			}
		}

		if idx == -1 {
			for i, it := range items {
				plItem := it.(item)
				if plItem.id == msg.TrackID {
					idx = i
					break
				}
			}
		}

		if idx == -1 {
			m.Playlist.SetNowPlaying("")
			m.Status.CurrentSong = ""
			m.updateSizes()
			break
		}

		m.Playlist.List.RemoveItem(idx)
		m.Playlist.SetNowPlaying("")

		items = m.Playlist.List.Items()
		if idx < len(items) {
			next := items[idx].(item)
			m.Playlist.SetNowPlayingAt(next.id, idx)
			return m, func() tea.Msg { return PlaySongMsg{Item: next, PlaylistIndex: idx, HasPlaylistIndex: true} }
		}
		m.Status.CurrentSong = ""
		m.updateSizes()
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if m.ShowQuitConfirm {
			switch msg.String() {
			case "y", "Y":
				return m, tea.Quit
			case "n", "N", "q", "esc":
				m.ShowQuitConfirm = false
			}
			return m, nil
		}

		// Global 'q' to quit (only if not filtering)
		if msg.String() == "q" {
			if m.Library.List.FilterState() != 0 || m.Playlist.List.FilterState() != 0 {
				// Let list handle it
			} else {
				m.ShowQuitConfirm = true
				return m, nil
			}
		}

		if msg.String() == "tab" {
			if m.Focus == FocusLibrary {
				m.Focus = FocusPlaylist
				m.Library.SetFocused(false)
				m.Playlist.SetFocused(true)
			} else {
				m.Focus = FocusLibrary
				m.Library.SetFocused(true)
				m.Playlist.SetFocused(false)
			}
		} else if msg.String() == "+" || msg.String() == "=" {
			m.Status, _ = m.Status.Update(VolumeChangeMsg{Delta: 0.1})
		} else if msg.String() == "-" || msg.String() == "_" {
			m.Status, _ = m.Status.Update(VolumeChangeMsg{Delta: -0.1})
		} else if msg.String() == " " {
			if m.Player != nil {
				m.Player.Pause()
			}
		}
	}

	// Propagate updates to sub-models
	libraryMsg := msg
	playlistMsg := msg
	if _, ok := msg.(tea.KeyMsg); ok {
		if m.Focus == FocusLibrary {
			playlistMsg = nil
		} else {
			libraryMsg = nil
		}
	}

	var cmd tea.Cmd
	m.Library, cmd = m.Library.Update(libraryMsg)
	cmds = append(cmds, cmd)

	m.Playlist, cmd = m.Playlist.Update(playlistMsg)
	cmds = append(cmds, cmd)

	statusBefore := m.Status.CurrentSong != "" || m.Status.Err != nil
	m.Status, cmd = m.Status.Update(msg)
	cmds = append(cmds, cmd)
	statusAfter := m.Status.CurrentSong != "" || m.Status.Err != nil

	if statusBefore != statusAfter {
		m.updateSizes()
	}

	return m, tea.Batch(cmds...)
}

func (m *MainModel) updateSizes() {
	availableWidth := m.Width - 4 // Account for Padding(1, 2)
	leftPanelWidth := availableWidth / 2
	rightPanelWidth := availableWidth - leftPanelWidth

	// Basic non-panel height = 2 padding-top + 1 help + 1 padding-bottom + 2 gaps = 6
	// If status is present, it adds 7 more lines.
	statusHeight := 0
	if m.Status.CurrentSong != "" || m.Status.Err != nil {
		statusHeight = 7
	}

	topPanelHeight := m.Height - (6 + statusHeight)
	if topPanelHeight < 5 {
		topPanelHeight = 5
	}

	m.Library.SetSize(leftPanelWidth, topPanelHeight)
	m.Playlist.SetSize(rightPanelWidth, topPanelHeight)
	m.Status.SetSize(m.Width - 4)
}

func (m MainModel) playSong(i item, playbackID uint64) tea.Cmd {
	return func() tea.Msg {
		if m.Client == nil {
			return ErrorMsg(errors.New("no music provider configured"))
		}
		if m.Player == nil {
			return ErrorMsg(errors.New("no audio player configured"))
		}

		r, err := m.Client.OpenTrackStream(context.Background(), i.id)
		if err != nil {
			return ErrorMsg(err)
		}
		if err := m.Player.Play(r); err != nil {
			_ = r.Close()
			return ErrorMsg(err)
		}
		started := SongStartedMsg{
			PlaybackID: playbackID,
			TrackID:    i.id,
			Title:      i.title,
			Artist:     i.artist,
			Album:      i.album,
			Year:       i.year,
			Track:      i.track,
			Duration:   i.duration,
		}
		if m.Playlist.NowPlayingIndex >= 0 {
			started.PlaylistIndex = m.Playlist.NowPlayingIndex
			started.HasPlaylistIndex = true
		}
		return started
	}
}

func (m MainModel) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Initializing..."
	}

	topPanels := lipgloss.JoinHorizontal(lipgloss.Top, m.Library.View(), m.Playlist.View())

	helpLine := helpStyle.Render(" Tab: Switch Panel • Enter: Select • s: Random • a: Add • x: Remove • X: Clear • Backspace: Back • Space: Pause • +/-: Volume • q: Quit")
	if m.ShowQuitConfirm {
		helpLine = quitConfirmStyle.Render(" Really quit? (y/n) ")
	}

	var mainView string
	statusView := m.Status.View()
	if statusView != "" {
		mainView = lipgloss.JoinVertical(lipgloss.Left,
			topPanels,
			statusView,
			helpLine,
		)
	} else {
		mainView = lipgloss.JoinVertical(lipgloss.Left,
			topPanels,
			helpLine,
		)
	}

	return lipgloss.NewStyle().Padding(2, 2, 1, 2).Render(mainView)
}
