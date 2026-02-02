package ui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type PlaylistModel struct {
	List      list.Model
	Width     int
	Height    int
	IsFocused bool
}

func NewPlaylistModel() PlaylistModel {
	delegate := CustomDelegate{}
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Playlist"
	l.SetShowHelp(false)

	return PlaylistModel{
		List: l,
	}
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

	case tea.KeyMsg:
		if !m.IsFocused || m.List.FilterState() == list.Filtering {
			break
		}

		if msg.String() == "enter" {
			sel := m.List.SelectedItem()
			if sel != nil {
				i := sel.(item)
				return m, func() tea.Msg {
					return PlaySongMsg{Item: i}
				}
			}
		}
	}

	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m PlaylistModel) View() string {
	style := panelStyle
	if m.IsFocused {
		style = focusedPanelStyle
	}
	return style.Width(m.Width).Height(m.Height).Render(m.List.View())
}

func (m *PlaylistModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.List.SetSize(w-2, h-2)
}
