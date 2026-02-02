package ui

import (
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

	Focus Focus
	Width int
	Height int

	Library  LibraryModel
	Playlist PlaylistModel
	Status   StatusModel
}

func NewModel(client MusicProvider, p AudioPlayer) MainModel {
	lib := NewLibraryModel(client)
	lib.IsFocused = true // Initial focus

	return MainModel{
		Client:   client,
		Player:   p,
		Focus:    FocusLibrary,
		Library:  lib,
		Playlist: NewPlaylistModel(),
		Status:   NewStatusModel(p),
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

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.updateSizes()

	case TickMsg:
		return m, tickCmd()

	case PlaySongMsg:
		return m, m.playSong(msg.Item)

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		
		// Global 'q' to quit (only if not filtering)
		if msg.String() == "q" {
			if m.Library.List.FilterState() != 0 || m.Playlist.List.FilterState() != 0 {
				// Let list handle it
			} else {
				return m, tea.Quit
			}
		}

		if msg.String() == "tab" {
			if m.Focus == FocusLibrary {
				m.Focus = FocusPlaylist
				m.Library.IsFocused = false
				m.Playlist.IsFocused = true
			} else {
				m.Focus = FocusLibrary
				m.Library.IsFocused = true
				m.Playlist.IsFocused = false
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
	var cmd tea.Cmd
	m.Library, cmd = m.Library.Update(msg)
	cmds = append(cmds, cmd)

	m.Playlist, cmd = m.Playlist.Update(msg)
	cmds = append(cmds, cmd)

	m.Status, cmd = m.Status.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *MainModel) updateSizes() {
	availableWidth := m.Width - 2
	leftPanelWidth := availableWidth / 2
	rightPanelWidth := availableWidth - leftPanelWidth
	
	// Adjusted for taller status bar
	// Title(1) + Gap(1) + Status(5) + Gap(1) + Help(1) = 9
	// Adding a bit more buffer for safety
	topPanelHeight := m.Height - 12
	if topPanelHeight < 5 {
		topPanelHeight = 5
	}

	m.Library.SetSize(leftPanelWidth, topPanelHeight)
	m.Playlist.SetSize(rightPanelWidth, topPanelHeight)
	m.Status.SetSize(m.Width)
}

func (m MainModel) playSong(i item) tea.Cmd {
	return func() tea.Msg {
		url, err := m.Client.GetStreamURL(i.id)
		if err != nil {
			return ErrorMsg(err)
		}
		err = m.Player.PlayStream(url)
		if err != nil {
			return ErrorMsg(err)
		}
		return SongStartedMsg{
			Title:    i.title,
			Artist:   i.artist,
			Album:    i.album,
			Year:     i.year,
			Track:    i.track,
			Duration: i.duration,
		}
	}
}

func (m MainModel) View() string {
	if m.Width == 0 || m.Height == 0 {
		return "Initializing..."
	}

	topPanels := lipgloss.JoinHorizontal(lipgloss.Top, m.Library.View(), m.Playlist.View())

	mainView := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Go Subsonic TUI"),
		"\n",
		topPanels,
		"\n",
		m.Status.View(),
		"\n",
		helpStyle.Render(" Tab: Switch Panel • Enter: Select • a: Add Album • Backspace: Back • Space: Pause • +/-: Volume • q: Quit"),
	)
	
	return lipgloss.NewStyle().Padding(0, 1).Render(mainView)
}