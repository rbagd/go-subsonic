package subsonic

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
)

func TestProviderMappingAndStreaming(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch path.Base(r.URL.Path) {
		case "getArtists":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"subsonic-response": {
					"status": "ok",
					"version": "1.16.1",
					"artists": {
						"index": [
							{"name": "A", "artist": [{"id": "artist1", "name": "Abba", "albumCount": 5}]}
						]
					}
				}
			}`))
			return

		case "getMusicDirectory":
			w.Header().Set("Content-Type", "application/json")
			id := r.URL.Query().Get("id")
			switch id {
			case "artist1":
				_, _ = w.Write([]byte(`{
					"subsonic-response": {
						"status": "ok",
						"version": "1.16.1",
						"directory": {
							"id": "artist1",
							"name": "Abba",
							"child": [
								{"id": "album1", "title": "Gold", "isDir": true, "year": 1992},
								{"id": "ignoreSong", "title": "Not an album", "isDir": false}
							]
						}
					}
				}`))
				return
			case "album1":
				_, _ = w.Write([]byte(`{
					"subsonic-response": {
						"status": "ok",
						"version": "1.16.1",
						"directory": {
							"id": "album1",
							"name": "Gold",
							"child": [
								{"id": "t1", "title": "Dancing Queen", "isDir": false, "track": 1, "duration": 123, "artist": "Abba", "album": "Gold", "year": 1992},
								{"id": "ignoreDir", "title": "Bonus", "isDir": true}
							]
						}
					}
				}`))
				return
			default:
				http.Error(w, "unknown id", http.StatusBadRequest)
				return
			}

		case "stream":
			if r.URL.Query().Get("id") == "" {
				http.Error(w, "missing id", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "audio/mpeg")
			_, _ = w.Write([]byte("streamdata"))
			return

		case "getAlbumList2":
			if r.URL.Query().Get("type") != "newest" {
				http.Error(w, "unsupported type", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"subsonic-response": {
					"status": "ok",
					"version": "1.16.1",
					"albumList2": {
						"album": [
							{"id": "ra1", "name": "Newest 1", "artist": "Artist A", "year": 2024},
							{"id": "ra2", "name": "Newest 2", "artist": "Artist B", "year": 2023}
						]
					}
				}
			}`))
			return
		default:
			http.NotFound(w, r)
			return
		}
	}))
	defer ts.Close()

	c := NewClient(ts.URL, "user", "pass")
	p := NewProvider(c)

	artists, err := p.ListArtists(context.Background())
	if err != nil {
		t.Fatalf("ListArtists() error = %v", err)
	}
	if len(artists) != 1 {
		t.Fatalf("ListArtists() len = %d, want 1", len(artists))
	}
	if artists[0].ID != "artist1" || artists[0].Name != "Abba" || artists[0].AlbumCount != 5 {
		t.Fatalf("ListArtists() = %#v, want Abba/artist1/5", artists[0])
	}

	albums, err := p.ListAlbums(context.Background(), "artist1")
	if err != nil {
		t.Fatalf("ListAlbums() error = %v", err)
	}
	if len(albums) != 1 {
		t.Fatalf("ListAlbums() len = %d, want 1", len(albums))
	}
	if albums[0].ID != "album1" || albums[0].Title != "Gold" || albums[0].Year != 1992 {
		t.Fatalf("ListAlbums() = %#v, want Gold/album1/1992", albums[0])
	}

	tracks, err := p.ListTracks(context.Background(), "album1")
	if err != nil {
		t.Fatalf("ListTracks() error = %v", err)
	}
	if len(tracks) != 1 {
		t.Fatalf("ListTracks() len = %d, want 1", len(tracks))
	}
	if tracks[0].ID != "t1" || tracks[0].Title != "Dancing Queen" || tracks[0].Track != 1 || tracks[0].Duration != 123 {
		t.Fatalf("ListTracks() = %#v, want t1/Dancing Queen/1/123", tracks[0])
	}

	rc, err := p.OpenTrackStream(context.Background(), "t1")
	if err != nil {
		t.Fatalf("OpenTrackStream() error = %v", err)
	}
	defer rc.Close()

	b, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(b) != "streamdata" {
		t.Fatalf("OpenTrackStream() body = %q, want %q", string(b), "streamdata")
	}

	recent, err := p.ListRecentlyAddedAlbums(context.Background(), 2, 0)
	if err != nil {
		t.Fatalf("ListRecentlyAddedAlbums() error = %v", err)
	}
	if len(recent) != 2 {
		t.Fatalf("ListRecentlyAddedAlbums() len = %d, want 2", len(recent))
	}
	if recent[0].ID != "ra1" || recent[0].Title != "Newest 1" || recent[0].Artist != "Artist A" || recent[0].Year != 2024 {
		t.Fatalf("ListRecentlyAddedAlbums()[0] = %#v, want ra1/Newest 1/Artist A/2024", recent[0])
	}
}

func TestProviderStreamClientHasNoOverallTimeout(t *testing.T) {
	t.Parallel()

	p := NewProvider(NewClient("http://example.com", "user", "pass"))
	if p.httpClient.Timeout != 0 {
		t.Fatalf("NewProvider() stream client Timeout = %s, want 0 (no overall request timeout for long audio streams)", p.httpClient.Timeout)
	}
}
