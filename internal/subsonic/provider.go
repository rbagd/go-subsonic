package subsonic

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"go-subsonic/internal/core"
)

type Provider struct {
	c          *Client
	httpClient *http.Client
}

func NewProvider(c *Client) *Provider {
	return &Provider{
		c:          c,
		httpClient: &http.Client{},
	}
}

func (p *Provider) ListArtists(ctx context.Context) ([]core.Artist, error) {
	if p.c == nil {
		return nil, fmt.Errorf("nil subsonic client")
	}
	artists, err := p.c.GetArtists(ctx)
	if err != nil {
		return nil, err
	}
	if artists == nil {
		return nil, nil
	}

	out := make([]core.Artist, 0)
	for _, idx := range artists.Index {
		for _, a := range idx.Artist {
			out = append(out, core.Artist{
				ID:         a.ID,
				Name:       a.Name,
				AlbumCount: a.AlbumCount,
			})
		}
	}
	return out, nil
}

func (p *Provider) ListAlbums(ctx context.Context, artistID string) ([]core.Album, error) {
	if p.c == nil {
		return nil, fmt.Errorf("nil subsonic client")
	}
	dir, err := p.c.GetMusicDirectory(ctx, artistID)
	if err != nil {
		return nil, err
	}
	if dir == nil {
		return nil, nil
	}

	out := make([]core.Album, 0)
	for _, child := range dir.Child {
		if !child.IsDir {
			continue
		}
		out = append(out, core.Album{
			ID:     child.ID,
			Title:  child.Title,
			Artist: child.Artist,
			Year:   child.Year,
		})
	}
	return out, nil
}

func (p *Provider) ListTracks(ctx context.Context, albumID string) ([]core.Track, error) {
	if p.c == nil {
		return nil, fmt.Errorf("nil subsonic client")
	}
	dir, err := p.c.GetMusicDirectory(ctx, albumID)
	if err != nil {
		return nil, err
	}
	if dir == nil {
		return nil, nil
	}

	out := make([]core.Track, 0)
	for _, child := range dir.Child {
		if child.IsDir {
			continue
		}
		out = append(out, core.Track{
			ID:       child.ID,
			Title:    child.Title,
			Artist:   child.Artist,
			Album:    child.Album,
			Year:     child.Year,
			Track:    child.Track,
			Duration: child.Duration,
		})
	}
	return out, nil
}

func (p *Provider) ListRecentlyAddedAlbums(ctx context.Context, size, offset int) ([]core.Album, error) {
	if p.c == nil {
		return nil, fmt.Errorf("nil subsonic client")
	}
	albumList, err := p.c.GetAlbumList2(ctx, "newest", size, offset)
	if err != nil {
		return nil, err
	}
	if albumList == nil {
		return nil, nil
	}

	out := make([]core.Album, 0, len(albumList.Album))
	for _, a := range albumList.Album {
		title := a.Name
		if title == "" {
			title = a.Title
		}
		artist := a.Artist
		if artist == "" {
			artist = a.ArtistName
		}
		out = append(out, core.Album{
			ID:     a.ID,
			Title:  title,
			Artist: artist,
			Year:   a.Year,
		})
	}

	return out, nil
}

func (p *Provider) OpenTrackStream(ctx context.Context, trackID string) (io.ReadCloser, error) {
	if p.c == nil {
		return nil, fmt.Errorf("nil subsonic client")
	}
	u, err := p.c.GetStreamURL(trackID)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("http error: %s", resp.Status)
	}

	return resp.Body, nil
}
