package app

import (
	"go-subsonic/internal/config"
	"go-subsonic/internal/core"
	"go-subsonic/internal/subsonic"
)

func BuildProvider(cfg *config.Config) (core.MusicProvider, *subsonic.Client) {
	if cfg == nil {
		return nil, nil
	}

	client := subsonic.NewClient(cfg.Server.URL, cfg.Server.Username, cfg.Server.Password)
	return subsonic.NewProvider(client), client
}
