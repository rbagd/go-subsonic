package ui

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type LibraryLevel int

const (
	LibraryArtists LibraryLevel = iota
	LibraryAlbums
	LibraryRecentAlbums
	LibrarySongs
)

const (
	recentlyAddedNodeID    = "__recent__"
	recentlyAddedNodeTitle = "[Recently Added Albums]"
	recentAlbumsPageSize   = 100
)

type LibraryModel struct {
	Client     MusicProvider
	List       list.Model
	Level      LibraryLevel
	ArtistID   string
	ArtistName string
	AlbumID    string
	AlbumName  string
	Width      int
	Height     int
	IsFocused  bool

	// State for navigation persistence
	ArtistIndex  int
	ArtistOffset int
	AlbumIndex   int
	AlbumOffset  int
}

func NewLibraryModel(client MusicProvider) LibraryModel {
	delegate := CustomDelegate{Focused: false}
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Artists"
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	return LibraryModel{
		Client: client,
		List:   l,
		Level:  LibraryArtists,
	}
}

func (m *LibraryModel) SetFocused(focused bool) {
	m.IsFocused = focused
	m.List.SetDelegate(CustomDelegate{Focused: focused})
}

func (m LibraryModel) Init() tea.Cmd {
	return m.fetchArtists()
}

func (m LibraryModel) Update(msg tea.Msg) (LibraryModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case RandomAlbumSelectedMsg:
		if m.Level == LibraryArtists {
			m.ArtistIndex = m.List.Index()
		}
		m.Level = LibraryAlbums
		m.ArtistID = msg.ArtistID
		m.ArtistName = msg.ArtistName
		m.List.Title = "Albums"
		m.List.ResetFilter()
		m.List.SetItems(msg.Albums)
		m.List.Select(msg.Index)
		if msg.Index < len(msg.Albums) {
			m.AlbumName = msg.Albums[msg.Index].(item).title
		}
		return m, nil

	case []list.Item:
		m.List.ResetFilter()
		m.List.SetItems(msg)

		// Restore position if we just went back
		if m.Level == LibraryArtists && m.ArtistIndex > 0 {
			m.List.Select(m.ArtistIndex)
			// Bubbletea list doesn't have a direct SetOffset, but Update will handle it if we select.
			// However, we might need a way to force the offset if possible.
			// Actually, list.Model handles visibility of selected item.
		} else if (m.Level == LibraryAlbums || m.Level == LibraryRecentAlbums) && m.AlbumIndex > 0 {
			m.List.Select(m.AlbumIndex)
		}

		return m, nil

	case tea.KeyMsg:
		if !m.IsFocused || m.List.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "s":
			return m, m.selectRandomAlbum()

		case "a":
			switch m.Level {
			case LibraryAlbums, LibraryRecentAlbums:
				sel := m.List.SelectedItem()
				if sel != nil {
					i := sel.(item)
					return m, m.fetchAlbumForPlaylist(i.id)
				}
			case LibrarySongs:
				sel := m.List.SelectedItem()
				if sel != nil {
					i := sel.(item)
					if !i.isDir {
						return m, func() tea.Msg {
							return AddToPlaylistMsg([]list.Item{i})
						}
					}
				}
			}

		case "backspace", "ctrl+h":
			switch m.Level {
			case LibraryAlbums:
				m.Level = LibraryArtists
				m.List.Title = "Artists"
				return m, m.fetchArtists()
			case LibraryRecentAlbums:
				m.Level = LibraryArtists
				m.List.Title = "Artists"
				return m, m.fetchArtists()
			case LibrarySongs:
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
						if i.id == recentlyAddedNodeID {
							m.Level = LibraryRecentAlbums
							m.ArtistName = "Recently Added"
							m.List.Title = "Recent Albums"
							m.AlbumIndex = 0
							return m, m.fetchRecentlyAddedAlbums()
						}

						m.ArtistIndex = m.List.Index()
						m.Level = LibraryAlbums
						m.ArtistID = i.id
						m.ArtistName = i.title
						m.List.Title = "Albums"
						// Reset album position since we're entering a new artist
						m.AlbumIndex = 0
						return m, m.fetchDirectory(i.id)
					} else if m.Level == LibraryAlbums || m.Level == LibraryRecentAlbums {
						m.AlbumIndex = m.List.Index()
						m.Level = LibrarySongs
						m.AlbumID = i.id
						m.AlbumName = i.title
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

	// Update AlbumName if we are in Album level and selection changed
	if m.Level == LibraryAlbums {
		sel := m.List.SelectedItem()
		if sel != nil {
			m.AlbumName = sel.(item).title
		}
	}

	return m, cmd
}

func (m LibraryModel) View() string {
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
	header := headerStyle.Width(headerWidth).Render(m.panelTitle())
	summary := m.summaryBox()

	var content string
	if summary != "" {
		innerW := m.Width - 4
		if innerW < 0 {
			innerW = 0
		}
		rightAlignedSummary := lipgloss.NewStyle().Width(innerW).Align(lipgloss.Right).Render(summary)
		content = lipgloss.JoinVertical(lipgloss.Left, header, rightAlignedSummary, m.List.View())
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left, header, m.List.View())
	}
	return style.Width(m.Width).Height(m.Height).Render(content)
}

func (m *LibraryModel) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	innerW := w - 4
	innerH := h - 4
	if innerW < 0 {
		innerW = 0
	}

	summaryHeight := 0
	if m.summaryBox() != "" {
		summaryHeight = 4 // 1 line + 2 borders + 1 margin
	}

	listH := innerH - 1 - summaryHeight // -1 for header
	if listH < 0 {
		listH = 0
	}
	m.List.SetSize(innerW, listH)
}

func (m LibraryModel) summaryBox() string {
	items := m.List.Items()
	if len(items) == 0 {
		return ""
	}

	var stats string
	switch m.Level {
	case LibraryArtists:
		artists := 0
		albums := 0
		for _, li := range items {
			i := li.(item)
			if i.id == recentlyAddedNodeID {
				continue
			}
			artists++
			var count int
			fmt.Sscanf(i.desc, "%d", &count)
			albums += count
		}
		stats = fmt.Sprintf("%d artists • %d albums", artists, albums)
		if artists == 1 {
			stats = strings.Replace(stats, "1 artists", "1 artist", 1)
		}
		if albums == 1 {
			stats = strings.Replace(stats, "1 albums", "1 album", 1)
		}

	case LibraryAlbums, LibraryRecentAlbums:
		albums := len(items)
		minYear, maxYear := 0, 0
		for _, li := range items {
			i := li.(item)
			if i.year > 0 {
				if minYear == 0 || i.year < minYear {
					minYear = i.year
				}
				if i.year > maxYear {
					maxYear = i.year
				}
			}
		}
		
		albumLabel := "albums"
		if albums == 1 {
			albumLabel = "album"
		}

		if minYear > 0 {
			if minYear == maxYear {
				stats = fmt.Sprintf("%d %s • %d", albums, albumLabel, minYear)
			} else {
				stats = fmt.Sprintf("%d %s • %d-%d", albums, albumLabel, minYear, maxYear)
			}
		} else {
			stats = fmt.Sprintf("%d %s", albums, albumLabel)
		}

	case LibrarySongs:
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
		
		stats = fmt.Sprintf("%d %s • %s", songs, songLabel, FmtDuration(time.Duration(totalDur)*time.Second))
	}

	if stats == "" {
		return ""
	}
	return statBoxStyle.Render(stats)
}

func (m LibraryModel) panelTitle() string {
	switch m.Level {
	case LibraryArtists:
		return "Library: Artists"
	case LibraryAlbums:
		return fmt.Sprintf("Library: %s", m.ArtistName)
	case LibraryRecentAlbums:
		return "Library: Recently Added"
	case LibrarySongs:
		return fmt.Sprintf("Library: %s -> %s", m.ArtistName, m.AlbumName)
	default:
		return "Library"
	}
}

func (m LibraryModel) fetchArtists() tea.Cmd {
	return func() tea.Msg {
		if m.Client == nil {
			return nil
		}
		artists, err := m.Client.ListArtists(context.Background())
		if err != nil {
			return ErrorMsg(err)
		}

		items := []list.Item{}
		items = append(items, item{
			title: recentlyAddedNodeTitle,
			desc:  "Newest albums",
			id:    recentlyAddedNodeID,
			isDir: true,
		})
		for _, artist := range artists {
			items = append(items, item{
				title:  artist.Name,
				desc:   fmt.Sprintf("%d", artist.AlbumCount),
				id:     artist.ID,
				isDir:  true,
				artist: artist.Name,
			})
		}
		return items
	}
}

func (m LibraryModel) fetchRecentlyAddedAlbums() tea.Cmd {
	return func() tea.Msg {
		if m.Client == nil {
			return nil
		}

		albums, err := m.Client.ListRecentlyAddedAlbums(context.Background(), recentAlbumsPageSize, 0)
		if err != nil {
			return ErrorMsg(err)
		}

		items := []list.Item{}
		for _, a := range albums {
			title := a.Title
			artist := a.Artist
			if artist != "" {
				title = fmt.Sprintf("%s - %s", artist, a.Title)
			}
			if a.Year != 0 {
				title = fmt.Sprintf("%d, %s", a.Year, title)
			}
			items = append(items, item{
				title:  title,
				desc:   "Album / Directory",
				id:     a.ID,
				isDir:  true,
				artist: a.Artist,
				album:  a.Title,
				year:   a.Year,
			})
		}
		return items
	}
}

func (m LibraryModel) fetchDirectory(id string) tea.Cmd {
	return func() tea.Msg {
		items := []list.Item{}
		switch m.Level {
		case LibraryAlbums:
			albums, err := m.Client.ListAlbums(context.Background(), id)
			if err != nil {
				return ErrorMsg(err)
			}
			for _, a := range albums {
				title := a.Title
				if a.Year != 0 {
					title = fmt.Sprintf("%d, %s", a.Year, a.Title)
				}
				items = append(items, item{
					title:  title,
					desc:   "Album / Directory",
					id:     a.ID,
					isDir:  true,
					artist: a.Artist,
					album:  a.Title,
					year:   a.Year,
				})
			}

		case LibrarySongs:
			tracks, err := m.Client.ListTracks(context.Background(), id)
			if err != nil {
				return ErrorMsg(err)
			}
			for _, t := range tracks {
				items = append(items, item{
					title:    t.Title,
					desc:     "Song",
					id:       t.ID,
					isDir:    false,
					duration: t.Duration,
					artist:   t.Artist,
					album:    t.Album,
					year:     t.Year,
					track:    t.Track,
				})
			}
		}
		return items
	}
}

func (m LibraryModel) fetchAlbumForPlaylist(id string) tea.Cmd {
	return func() tea.Msg {
		tracks, err := m.Client.ListTracks(context.Background(), id)
		if err != nil {
			return ErrorMsg(err)
		}

		items := []list.Item{}
		for _, t := range tracks {
			items = append(items, item{
				title:    t.Title,
				desc:     "Song",
				id:       t.ID,
				isDir:    false,
				duration: t.Duration,
				artist:   t.Artist,
				album:    t.Album,
				year:     t.Year,
				track:    t.Track,
			})
		}
		return AddToPlaylistMsg(items)
	}
}

func (m LibraryModel) selectRandomAlbum() tea.Cmd {
	return func() tea.Msg {
		if m.Client == nil {
			return nil
		}

		// 1. Get Artists
		artists, err := m.Client.ListArtists(context.Background())
		if err != nil {
			return ErrorMsg(err)
		}
		if len(artists) == 0 {
			return nil
		}

		// 2. Pick random artist
		artist := artists[rand.Intn(len(artists))]

		// 3. Get Albums for artist
		albums, err := m.Client.ListAlbums(context.Background(), artist.ID)
		if err != nil {
			return ErrorMsg(err)
		}
		if len(albums) == 0 {
			return nil
		}

		// 4. Pick random album
		albumIndex := rand.Intn(len(albums))

		items := make([]list.Item, len(albums))
		for i, a := range albums {
			title := a.Title
			if a.Year != 0 {
				title = fmt.Sprintf("%d, %s", a.Year, a.Title)
			}
			items[i] = item{
				title:  title,
				desc:   "Album / Directory",
				id:     a.ID,
				isDir:  true,
				artist: a.Artist,
				album:  a.Title,
				year:   a.Year,
			}
		}

		return RandomAlbumSelectedMsg{
			ArtistID:   artist.ID,
			ArtistName: artist.Name,
			Albums:     items,
			Index:      albumIndex,
		}
	}
}

// CustomDelegate and item move to a shared place or kept here if they are only used by lists.
// Since Playlist also uses them, I'll move them to a shared file later or keep them in model.go for now.
// Actually, let's put them in internal/ui/delegate.go
