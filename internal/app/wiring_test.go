package app

import (
	"testing"

	"go-subsonic/internal/config"
)

func TestBuildProviderNilConfig(t *testing.T) {
	provider, client := BuildProvider(nil)
	if provider != nil {
		t.Fatalf("BuildProvider(nil) provider = %T, want nil", provider)
	}
	if client != nil {
		t.Fatalf("BuildProvider(nil) client = %T, want nil", client)
	}
}

func TestBuildProviderNonNilConfig(t *testing.T) {
	cfg := &config.Config{
		Server: config.ServerConfig{
			URL:      "http://example.com",
			Username: "user",
			Password: "pass",
		},
	}

	provider, client := BuildProvider(cfg)
	if provider == nil {
		t.Fatalf("BuildProvider(cfg) provider = nil, want non-nil")
	}
	if client == nil {
		t.Fatalf("BuildProvider(cfg) client = nil, want non-nil")
	}
}
