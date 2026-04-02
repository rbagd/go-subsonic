package subsonic

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGetIndexes(t *testing.T) {
	// Mock Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"subsonic-response": {
				"status": "ok",
				"version": "1.16.1",
				"indexes": {
					"lastModified": 1600000000,
					"index": [
						{
							"name": "A",
							"artist": [
								{"id": "1", "name": "Abba", "albumCount": 5}
							]
						}
					]
				}
			}
		}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "user", "pass")
	indexes, err := client.GetIndexes(context.Background())
	if err != nil {
		t.Fatalf("GetIndexes failed: %v", err)
	}

	if len(indexes.Index) != 1 {
		t.Errorf("Expected 1 index, got %d", len(indexes.Index))
	}
	if indexes.Index[0].Artist[0].Name != "Abba" {
		t.Errorf("Expected artist Abba, got %s", indexes.Index[0].Artist[0].Name)
	}
}

func TestGetMusicDirectory(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"subsonic-response": {
				"status": "ok",
				"version": "1.16.1",
				"directory": {
					"id": "10",
					"name": "Gold",
					"child": [
						{"id": "101", "title": "Dancing Queen", "isDir": false, "track": 1}
					]
				}
			}
		}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "user", "pass")
	dir, err := client.GetMusicDirectory(context.Background(), "10")
	if err != nil {
		t.Fatalf("GetMusicDirectory failed: %v", err)
	}

	if dir.Name != "Gold" {
		t.Errorf("Expected dir name Gold, got %s", dir.Name)
	}
	if len(dir.Child) != 1 {
		t.Errorf("Expected 1 child, got %d", len(dir.Child))
	}
	if dir.Child[0].Title != "Dancing Queen" {
		t.Errorf("Expected title Dancing Queen, got %s", dir.Child[0].Title)
	}
}

func TestClientUsesInjectedHTTPClient(t *testing.T) {
	t.Parallel()

	var calls int
	rt := roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		body := `{"subsonic-response":{"status":"ok","version":"1.16.1"}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    r,
		}, nil
	})

	client := NewClient("http://should-not-hit-network.invalid", "user", "pass")
	client.HTTPClient = &http.Client{Transport: rt}

	if err := client.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() error = %v, want nil", err)
	}
	if calls != 1 {
		t.Fatalf("HTTP transport calls = %d, want 1", calls)
	}
}

func TestNewRequestReturnsErrorWhenSaltGenerationFails(t *testing.T) {
	origGenerateSalt := generateSalt
	t.Cleanup(func() { generateSalt = origGenerateSalt })

	generateSalt = func(n int) (string, error) {
		return "", errors.New("rng unavailable")
	}

	client := NewClient("http://example.com", "user", "pass")
	if err := client.Ping(context.Background()); err == nil {
		t.Fatalf("Ping() error = nil, want non-nil")
	}
}
