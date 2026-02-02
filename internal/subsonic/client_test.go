package subsonic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

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
