package ui

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go-subsonic/internal/core"
)

type fakeProvider struct {
	artists []core.Artist
	albums  []core.Album
	tracks  []core.Track
	recent  []core.Album
}

func (f fakeProvider) ListArtists(ctx context.Context) ([]core.Artist, error) { return f.artists, nil }
func (f fakeProvider) ListAlbums(ctx context.Context, artistID string) ([]core.Album, error) {
	return f.albums, nil
}
func (f fakeProvider) ListTracks(ctx context.Context, albumID string) ([]core.Track, error) {
	return f.tracks, nil
}
func (f fakeProvider) ListRecentlyAddedAlbums(ctx context.Context, size, offset int) ([]core.Album, error) {
	return f.recent, nil
}
func (f fakeProvider) OpenTrackStream(ctx context.Context, trackID string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(nil)), nil
}

type fakePlayer struct{}

func (fakePlayer) Play(r io.ReadCloser) error { return r.Close() }
func (fakePlayer) Pause()                     {}
func (fakePlayer) Stop()                      {}
func (fakePlayer) SetVolume(vol float64)      {}
func (fakePlayer) Position() time.Duration    { return 0 }
func (fakePlayer) Done() <-chan struct{}      { ch := make(chan struct{}); close(ch); return ch }

func TestKeyRoutingDoesNotSyncNavigation(t *testing.T) {
	m := NewModel(nil, nil)
	m.Library.List.SetItems([]list.Item{item{title: "a"}, item{title: "b"}})
	m.Playlist.List.SetItems([]list.Item{item{title: "p1"}, item{title: "p2"}})

	m.Focus = FocusLibrary
	m.Library.SetFocused(true)
	m.Playlist.SetFocused(false)

	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	nm := newM.(*MainModel)

	if nm.Library.List.Index() != 1 {
		t.Fatalf("library index = %d, want 1", nm.Library.List.Index())
	}
	if nm.Playlist.List.Index() != 0 {
		t.Fatalf("playlist index = %d, want 0", nm.Playlist.List.Index())
	}
}

func TestLibraryAddSongWithA(t *testing.T) {
	lib := NewLibraryModel(nil)
	lib.SetFocused(true)
	lib.Level = LibrarySongs
	lib.List.Title = "Songs"
	lib.List.SetItems([]list.Item{
		item{title: "Dancing Queen", id: "t1", isDir: false, duration: 123, artist: "Abba", album: "Gold"},
	})

	_, cmd := lib.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd == nil {
		t.Fatalf("cmd = nil, want non-nil")
	}

	msg := cmd()
	add, ok := msg.(AddToPlaylistMsg)
	if !ok {
		t.Fatalf("msg type = %T, want AddToPlaylistMsg", msg)
	}
	if len(add) != 1 {
		t.Fatalf("AddToPlaylistMsg len = %d, want 1", len(add))
	}
	got := add[0].(item)
	if got.id != "t1" {
		t.Fatalf("added item id = %q, want %q", got.id, "t1")
	}
}

func TestFetchDirectorySongsDoesNotPrefixTitle(t *testing.T) {
	lib := NewLibraryModel(fakeProvider{
		tracks: []core.Track{
			{ID: "t1", Title: "Dancing Queen", Track: 1, Duration: 123, Artist: "Abba", Album: "Gold", Year: 1992},
		},
	})
	lib.Level = LibrarySongs

	cmd := lib.fetchDirectory("album1")
	msg := cmd()
	items, ok := msg.([]list.Item)
	if !ok {
		t.Fatalf("msg type = %T, want []list.Item", msg)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	it := items[0].(item)
	if strings.HasPrefix(it.title, "1.") || strings.HasPrefix(it.title, "1. ") {
		t.Fatalf("song title = %q, want no track-number prefix", it.title)
	}
	if it.title != "Dancing Queen" {
		t.Fatalf("song title = %q, want %q", it.title, "Dancing Queen")
	}
}

func TestPlaylistRendersArtistAlbum(t *testing.T) {
	pl := NewPlaylistModel()
	pl.SetFocused(true)
	pl.SetSize(60, 14)
	pl.List.SetItems([]list.Item{
		item{title: "Dancing Queen", id: "t1", isDir: false, duration: 123, artist: "Abba", album: "Gold"},
	})

	view := pl.View()
	if !strings.Contains(view, "Abba") || !strings.Contains(view, "Gold") {
		t.Fatalf("playlist view missing artist/album metadata: %q", view)
	}
}

func TestPanelHeaderAlwaysShowsContext(t *testing.T) {
	lib := NewLibraryModel(nil)
	lib.SetFocused(true)
	lib.Level = LibraryArtists
	lib.List.Title = "Artists"
	lib.SetSize(60, 10)

	// Enter filtering mode; bubbles/list replaces the title area.
	lib.List, _ = lib.List.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})

	view := lib.View()
	if !strings.Contains(view, "Library") || !strings.Contains(view, "Artists") {
		t.Fatalf("panel header missing expected context: %q", view)
	}
}

func TestBackspaceNavigatesUpInLibrary(t *testing.T) {
	m := NewModel(nil, nil)
	m.Focus = FocusLibrary
	m.Library.SetFocused(true)
	m.Playlist.SetFocused(false)

	m.Library.Level = LibraryAlbums
	m.Library.List.Title = "Albums"

	newM, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	nm := newM.(*MainModel)
	if nm.Library.Level != LibraryArtists {
		t.Fatalf("library level = %v, want %v", nm.Library.Level, LibraryArtists)
	}
}

func TestPlaylistDeleteSelectedItem(t *testing.T) {
	pl := NewPlaylistModel()
	pl.SetFocused(true)
	pl.List.SetItems([]list.Item{
		item{title: "one", id: "1"},
		item{title: "two", id: "2"},
	})

	pl.List.Select(0)
	newPL, _ := pl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})

	if len(newPL.List.Items()) != 1 {
		t.Fatalf("items len = %d, want 1", len(newPL.List.Items()))
	}
	got := newPL.List.Items()[0].(item)
	if got.id != "2" {
		t.Fatalf("remaining item id = %q, want %q", got.id, "2")
	}
}

func TestPlaylistClearAllItems(t *testing.T) {
	pl := NewPlaylistModel()
	pl.SetFocused(true)
	pl.List.SetItems([]list.Item{
		item{title: "one", id: "1"},
		item{title: "two", id: "2"},
	})

	newPL, _ := pl.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("X")})
	if len(newPL.List.Items()) != 0 {
		t.Fatalf("items len = %d, want 0", len(newPL.List.Items()))
	}
}

func TestPlaylistIndicatesNowPlaying(t *testing.T) {
	pl := NewPlaylistModel()
	pl.SetFocused(true)
	pl.SetSize(60, 10)
	pl.List.SetItems([]list.Item{
		item{title: "First", id: "t1", isDir: false, duration: 1, artist: "A", album: "B"},
		item{title: "Second", id: "t2", isDir: false, duration: 1, artist: "A", album: "B"},
	})

	pl, _ = pl.Update(SongStartedMsg{TrackID: "t2", PlaybackID: 1, Title: "Second"})
	pl.List.Select(1)

	view := pl.View()
	if !strings.Contains(view, "> Second") {
		t.Fatalf("playlist view missing now-playing indicator: %q", view)
	}
	if strings.Contains(view, "> First") {
		t.Fatalf("playlist view incorrectly marks non-playing item: %q", view)
	}
}

func TestPlaylistAutoAdvancesOnSongFinished(t *testing.T) {
	m := NewModel(fakeProvider{}, fakePlayer{})
	m.Playlist.List.SetItems([]list.Item{
		item{title: "First", id: "t1", isDir: false, duration: 1, artist: "A", album: "B"},
		item{title: "Second", id: "t2", isDir: false, duration: 1, artist: "A", album: "B"},
	})

	newM, _ := m.Update(SongStartedMsg{TrackID: "t1", PlaybackID: 1, Title: "First"})
	m = newM.(*MainModel)

	newM, cmd := m.Update(SongFinishedMsg{TrackID: "t1", PlaybackID: 1})
	m = newM.(*MainModel)

	if len(m.Playlist.List.Items()) != 1 {
		t.Fatalf("playlist items len = %d, want 1", len(m.Playlist.List.Items()))
	}
	remaining := m.Playlist.List.Items()[0].(item)
	if remaining.id != "t2" {
		t.Fatalf("remaining item id = %q, want %q", remaining.id, "t2")
	}

	if cmd == nil {
		t.Fatalf("cmd = nil, want non-nil")
	}
	msg := cmd()
	play, ok := msg.(PlaySongMsg)
	if !ok {
		t.Fatalf("msg type = %T, want PlaySongMsg", msg)
	}
	if play.Item.id != "t2" {
		t.Fatalf("next item id = %q, want %q", play.Item.id, "t2")
	}
}

func TestPlaylistDoesNotAutoAdvanceAfterLastSong(t *testing.T) {
	m := NewModel(fakeProvider{}, fakePlayer{})
	m.Playlist.List.SetItems([]list.Item{
		item{title: "Only", id: "t1", isDir: false, duration: 1, artist: "A", album: "B"},
	})

	newM, _ := m.Update(SongStartedMsg{TrackID: "t1", PlaybackID: 1, Title: "Only"})
	m = newM.(*MainModel)

	newM, cmd := m.Update(SongFinishedMsg{TrackID: "t1", PlaybackID: 1})
	m = newM.(*MainModel)

	if len(m.Playlist.List.Items()) != 0 {
		t.Fatalf("playlist items len = %d, want 0", len(m.Playlist.List.Items()))
	}
	if cmd != nil {
		t.Fatalf("cmd = non-nil, want nil")
	}
}

func TestPlaylistAutoAdvanceRemovesCorrectDuplicateTrackInstance(t *testing.T) {
	m := NewModel(fakeProvider{}, fakePlayer{})
	m.Playlist.List.SetItems([]list.Item{
		item{title: "First", id: "t1", isDir: false, duration: 1, artist: "A", album: "B"},
		item{title: "Middle", id: "t2", isDir: false, duration: 1, artist: "A", album: "B"},
		item{title: "Duplicate", id: "t1", isDir: false, duration: 1, artist: "A", album: "B"},
	})

	newM, _ := m.Update(SongStartedMsg{TrackID: "t1", PlaybackID: 1, Title: "Duplicate", PlaylistIndex: 2, HasPlaylistIndex: true})
	m = newM.(*MainModel)

	newM, _ = m.Update(SongFinishedMsg{TrackID: "t1", PlaybackID: 1, PlaylistIndex: 2, HasPlaylistIndex: true})
	m = newM.(*MainModel)

	if len(m.Playlist.List.Items()) != 2 {
		t.Fatalf("playlist items len = %d, want 2", len(m.Playlist.List.Items()))
	}
	first := m.Playlist.List.Items()[0].(item)
	second := m.Playlist.List.Items()[1].(item)
	if first.title != "First" || second.title != "Middle" {
		t.Fatalf("remaining items = [%q, %q], want [\"First\", \"Middle\"]", first.title, second.title)
	}
}

func TestStatusViewRespectsConfiguredWidth(t *testing.T) {
	st := NewStatusModel(nil)
	st.CurrentSong = "Song Title"
	st.SetSize(30)
	view := st.View()

	maxWidth := 0
	for _, line := range strings.Split(view, "\n") {
		w := lipgloss.Width(line)
		if w > maxWidth {
			maxWidth = w
		}
	}

	if maxWidth != 30 {
		t.Fatalf("rendered max width = %d, want 30", maxWidth)
	}
}

func TestLibraryResetFilterOnNavigation(t *testing.T) {
	lib := NewLibraryModel(nil)
	lib.List.SetItems([]list.Item{
		item{title: "Artist A", id: "a1", isDir: true},
		item{title: "Artist B", id: "b1", isDir: true},
	})

	// Simulate filtering for "A"
	lib.List, _ = lib.List.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	lib.List, _ = lib.List.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("A")})
	lib.List, _ = lib.List.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if lib.List.FilterValue() != "A" {
		t.Fatalf("filter value = %q, want %q", lib.List.FilterValue(), "A")
	}

	// Simulate selecting the item and pressing enter to navigate
	lib.IsFocused = true
	newLib, _ := lib.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Now simulate receiving new items (e.g. albums)
	newItems := []list.Item{
		item{title: "Album 1", id: "al1", isDir: true},
	}
	finalLib, _ := newLib.Update(newItems)

	if finalLib.List.FilterValue() != "" {
		t.Fatalf("filter value = %q after navigation, want empty", finalLib.List.FilterValue())
	}
}

func TestLibraryPreservesPositionOnNavigation(t *testing.T) {
	lib := NewLibraryModel(nil)
	lib.List.SetItems([]list.Item{
		item{title: "Artist A", id: "a1", isDir: true},
		item{title: "Artist B", id: "b1", isDir: true},
	})
	lib.List.Select(1) // Select Artist B
	lib.IsFocused = true

	// Navigate to Albums
	newLib, _ := lib.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if newLib.Level != LibraryAlbums {
		t.Fatalf("level = %v, want LibraryAlbums", newLib.Level)
	}
	if newLib.ArtistIndex != 1 {
		t.Fatalf("saved artist index = %d, want 1", newLib.ArtistIndex)
	}

	// Navigate back to Artists
	backLib, _ := newLib.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if backLib.Level != LibraryArtists {
		t.Fatalf("level = %v, want LibraryArtists", backLib.Level)
	}

	// Simulate receiving artists list again
	artistsItems := []list.Item{
		item{title: "Artist A", id: "a1", isDir: true},
		item{title: "Artist B", id: "b1", isDir: true},
	}
	finalLib, _ := backLib.Update(artistsItems)

	if finalLib.List.Index() != 1 {
		t.Fatalf("index after back = %d, want 1", finalLib.List.Index())
	}
}

func TestLibraryArtistsIncludesRecentlyAddedEntryFirst(t *testing.T) {
	lib := NewLibraryModel(fakeProvider{artists: []core.Artist{{ID: "a1", Name: "Abba", AlbumCount: 5}}})

	msg := lib.fetchArtists()()
	items, ok := msg.([]list.Item)
	if !ok {
		t.Fatalf("msg type = %T, want []list.Item", msg)
	}
	if len(items) < 2 {
		t.Fatalf("items len = %d, want at least 2", len(items))
	}
	first := items[0].(item)
	if first.title != "[Recently Added Albums]" {
		t.Fatalf("first item title = %q, want %q", first.title, "[Recently Added Albums]")
	}
}

func TestSelectingRecentlyAddedLoadsNewestAlbums(t *testing.T) {
	lib := NewLibraryModel(fakeProvider{recent: []core.Album{{ID: "ra1", Title: "Newest 1", Artist: "Artist A", Year: 2024}}})
	lib.SetFocused(true)
	lib.List.SetItems([]list.Item{item{title: "[Recently Added Albums]", id: "__recent__", isDir: true}})

	newLib, cmd := lib.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("cmd = nil, want non-nil")
	}
	if newLib.Level != LibraryRecentAlbums {
		t.Fatalf("level = %v, want %v", newLib.Level, LibraryRecentAlbums)
	}

	msg := cmd()
	items, ok := msg.([]list.Item)
	if !ok {
		t.Fatalf("msg type = %T, want []list.Item", msg)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	got := items[0].(item)
	if got.id != "ra1" || got.title != "2024, Artist A - Newest 1" {
		t.Fatalf("recent album item = %#v, want id ra1 and title '2024, Artist A - Newest 1'", got)
	}
}

func TestLibraryAddAlbumWithAFromRecentlyAdded(t *testing.T) {
	lib := NewLibraryModel(fakeProvider{
		tracks: []core.Track{{ID: "t1", Title: "Song 1", Artist: "Artist A", Album: "Newest 1", Duration: 123}},
	})
	lib.SetFocused(true)
	lib.Level = LibraryRecentAlbums
	lib.List.Title = "Recent Albums"
	lib.List.SetItems([]list.Item{
		item{title: "2024, Artist A - Newest 1", id: "ra1", isDir: true, artist: "Artist A", album: "Newest 1", year: 2024},
	})

	_, cmd := lib.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	if cmd == nil {
		t.Fatalf("cmd = nil, want non-nil")
	}

	msg := cmd()
	add, ok := msg.(AddToPlaylistMsg)
	if !ok {
		t.Fatalf("msg type = %T, want AddToPlaylistMsg", msg)
	}
	if len(add) != 1 {
		t.Fatalf("AddToPlaylistMsg len = %d, want 1", len(add))
	}
	got := add[0].(item)
	if got.id != "t1" {
		t.Fatalf("added item id = %q, want %q", got.id, "t1")
	}
}
