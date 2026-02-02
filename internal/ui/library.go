package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type LibraryLevel int

const (
	LibraryArtists LibraryLevel = iota
	LibraryAlbums
	LibrarySongs
)

type LibraryModel struct {
	Client    MusicProvider
	List      list.Model
	Level     LibraryLevel
	ArtistID  string
	AlbumID   string
	Width     int
	Height    int
	IsFocused bool
}

func NewLibraryModel(client MusicProvider) LibraryModel {
	delegate := CustomDelegate{}
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Artists"
	l.SetShowHelp(false)

	return LibraryModel{
		Client: client,
		List:   l,
		Level:  LibraryArtists,
	}
}

func (m LibraryModel) Init() tea.Cmd {
	return m.fetchArtists()
}

func (m LibraryModel) Update(msg tea.Msg) (LibraryModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case []list.Item:
		m.List.SetItems(msg)
		return m, nil

	case tea.KeyMsg:
		if !m.IsFocused || m.List.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "a":
			if m.Level == LibraryAlbums {
				sel := m.List.SelectedItem()
				if sel != nil {
					i := sel.(item)
					return m, m.fetchAlbumForPlaylist(i.id)
				}
			}

		case "backspace", "ctrl+h":
			if m.Level == LibraryAlbums {
				m.Level = LibraryArtists
				m.List.Title = "Artists"
				return m, m.fetchArtists()
			} else if m.Level == LibrarySongs {
				m.Level = LibraryAlbums
				m.List.Title = "Albums"
				// Re-fetch albums for current artist
				return m, m.fetchDirectory(m.ArtistID)
			}

		case "enter":
			sel := m.List.SelectedItem()
			if sel != nil {
				i := sel.(item)
				if i.isDir {
					if m.Level == LibraryArtists {
						m.Level = LibraryAlbums
						m.ArtistID = i.id
						m.List.Title = "Albums"
						return m, m.fetchDirectory(i.id)
					} else if m.Level == LibraryAlbums {
						m.Level = LibrarySongs
						m.AlbumID = i.id
						m.List.Title = "Songs"
						return m, m.fetchDirectory(i.id)
					}
				} else {
					return m, func() tea.Msg {
						return PlaySongMsg{Item: i}
					}
				}
			}
		}
	}

	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m LibraryModel) View() string {
	style := panelStyle
	if m.IsFocused {
		style = focusedPanelStyle
	}
	return style.Width(m.Width).Height(m.Height).Render(m.List.View())
}

func (m *LibraryModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.List.SetSize(w-2, h-2)
}

func (m LibraryModel) fetchArtists() tea.Cmd {
	return func() tea.Msg {
		if m.Client == nil {
			return nil
		}
		artists, err := m.Client.GetArtists(context.Background())
		if err != nil {
			return ErrorMsg(err)
		}

		items := []list.Item{}
		for _, idx := range artists.Index {
			for _, artist := range idx.Artist {
				items = append(items, item{
					title:  artist.Name,
					desc:   fmt.Sprintf("%d Albums", artist.AlbumCount),
					id:     artist.ID,
					isDir:  true,
					artist: artist.Name,
				})
			}
		}
		return items
	}
}

func (m LibraryModel) fetchDirectory(id string) tea.Cmd {
	return func() tea.Msg {
		dir, err := m.Client.GetMusicDirectory(context.Background(), id)
		if err != nil {
			return ErrorMsg(err)
		}

		items := []list.Item{}
		for _, child := range dir.Child {
			desc := "Song"
			title := child.Title
			if child.IsDir {
				desc = "Album / Directory"
				if child.Year != 0 {
					title = fmt.Sprintf("%d, %s", child.Year, child.Title)
				}
			} else if child.Track != 0 {
				title = fmt.Sprintf("%d. %s", child.Track, child.Title)
			}
			items = append(items, item{
				title:    title,
				desc:     desc,
				id:       child.ID,
				isDir:    child.IsDir,
				duration: child.Duration,
				artist:   child.Artist,
				album:    child.Album,
				year:     child.Year,
				track:    child.Track,
			})
		}
		return items
	}
}

func (m LibraryModel) fetchAlbumForPlaylist(id string) tea.Cmd {
	return func() tea.Msg {
		dir, err := m.Client.GetMusicDirectory(context.Background(), id)
		if err != nil {
			return ErrorMsg(err)
		}

		items := []list.Item{}
		for _, child := range dir.Child {
			if !child.IsDir {
				title := child.Title
				if child.Track != 0 {
					title = fmt.Sprintf("%d. %s", child.Track, child.Title)
				}
				items = append(items, item{
					title:    title,
					desc:     "Song",
					id:       child.ID,
					isDir:    false,
					duration: child.Duration,
					artist:   child.Artist,
					album:    child.Album,
					year:     child.Year,
					track:    child.Track,
				})
			}
		}
		return AddToPlaylistMsg(items)
	}
}

// CustomDelegate and item move to a shared place or kept here if they are only used by lists.
// Since Playlist also uses them, I'll move them to a shared file later or keep them in model.go for now.
// Actually, let's put them in internal/ui/delegate.go
