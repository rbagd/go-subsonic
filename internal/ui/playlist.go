package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PlaylistModel struct {
	List      list.Model
	Width     int
	Height    int
	IsFocused bool

	NowPlayingTrackID string
	NowPlayingIndex   int
}

func NewPlaylistModel() PlaylistModel {
	delegate := CustomDelegate{Focused: false}
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Playlist"
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.Styles.NoItems = lipgloss.NewStyle().MaxHeight(0).Padding(0).Margin(0)

	return PlaylistModel{
		List:            l,
		NowPlayingIndex: -1,
	}
}

func (m *PlaylistModel) updateDelegate() {
	m.List.SetDelegate(CustomDelegate{Focused: m.IsFocused, PlayingID: m.NowPlayingTrackID})
}

func (m *PlaylistModel) SetFocused(focused bool) {
	m.IsFocused = focused
	m.updateDelegate()
}

func (m *PlaylistModel) SetNowPlaying(trackID string) {
	m.NowPlayingIndex = -1
	m.NowPlayingTrackID = trackID
	m.updateDelegate()
}

func (m *PlaylistModel) SetNowPlayingAt(trackID string, idx int) {
	m.NowPlayingIndex = idx
	m.NowPlayingTrackID = trackID
	m.updateDelegate()
}

func (m PlaylistModel) Init() tea.Cmd {
	return nil
}

func (m PlaylistModel) Update(msg tea.Msg) (PlaylistModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case AddToPlaylistMsg:
		for _, i := range msg {
			m.List.InsertItem(len(m.List.Items()), i)
		}
		return m, nil

	case SongStartedMsg:
		if msg.HasPlaylistIndex {
			m.SetNowPlayingAt(msg.TrackID, msg.PlaylistIndex)
		} else {
			m.SetNowPlaying(msg.TrackID)
		}
		return m, nil

	case SongFinishedMsg:
		if msg.TrackID == m.NowPlayingTrackID {
			m.SetNowPlaying("")
		}
		return m, nil

	case tea.KeyMsg:
		if !m.IsFocused || m.List.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "x":
			if len(m.List.Items()) == 0 {
				return m, nil
			}
			idx := m.List.GlobalIndex()
			sel := m.List.SelectedItem()
			if sel != nil {
				i := sel.(item)
				if i.id == m.NowPlayingTrackID {
					m.SetNowPlaying("")
				} else if idx >= 0 && m.NowPlayingIndex >= 0 && idx < m.NowPlayingIndex {
					m.NowPlayingIndex--
				}
			}
			m.List.RemoveItem(idx)
			return m, nil

		case "X":
			m.List.SetItems([]list.Item{})
			m.SetNowPlaying("")
			return m, nil

		case "enter":
			sel := m.List.SelectedItem()
			if sel != nil {
				i := sel.(item)
				idx := m.List.GlobalIndex()
				return m, func() tea.Msg {
					return PlaySongMsg{Item: i, PlaylistIndex: idx, HasPlaylistIndex: true}
				}
			}
		}
	}

	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m PlaylistModel) View() string {
	style := panelStyle
	headerStyle := panelHeaderStyle
	if m.IsFocused {
		style = focusedPanelStyle
		headerStyle = focusedPanelHeaderStyle
	}

	headerWidth := m.Width - 4
	if headerWidth < 0 {
		headerWidth = 0
	}
	header := headerStyle.Width(headerWidth).Render("Playlist")
	summary := m.summaryBox()

	var listView string
	if len(m.List.Items()) == 0 {
		emptyMsg := "(empty)"
		innerW := m.Width - 4
		innerH := m.Height - 4
		if innerW < 0 {
			innerW = 0
		}
		if innerH < 0 {
			innerH = 0
		}
		listView = lipgloss.NewStyle().
			Width(innerW).
			Height(innerH - 1). // -1 for header
			Align(lipgloss.Center, lipgloss.Center).
			Render(emptyMsg)
	} else {
		listView = m.List.View()
	}

	var content string
	if summary != "" {
		innerW := m.Width - 4
		if innerW < 0 {
			innerW = 0
		}
		rightAlignedSummary := lipgloss.NewStyle().Width(innerW).Align(lipgloss.Right).Render(summary)
		content = lipgloss.JoinVertical(lipgloss.Left, header, rightAlignedSummary, listView)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left, header, listView)
	}
	return style.Width(m.Width).Height(m.Height).Render(content)
}

func (m *PlaylistModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	innerW := w - 4
	innerH := h - 4
	if innerW < 0 {
		innerW = 0
	}

	summaryHeight := 0
	if m.summaryBox() != "" {
		summaryHeight = 4
	}

	listH := innerH - 1 - summaryHeight
	if listH < 0 {
		listH = 0
	}
	m.List.SetSize(innerW, listH)
}

func (m PlaylistModel) summaryBox() string {
	items := m.List.Items()
	if len(items) == 0 {
		return ""
	}

	songs := len(items)
	totalDur := 0
	for _, li := range items {
		i := li.(item)
		totalDur += i.duration
	}

	songLabel := "songs"
	if songs == 1 {
		songLabel = "song"
	}

	stats := fmt.Sprintf("%d %s • %s", songs, songLabel, FmtDuration(time.Duration(totalDur)*time.Second))
	return statBoxStyle.Render(stats)
}
