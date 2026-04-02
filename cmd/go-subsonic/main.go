package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"go-subsonic/internal/app"
	"go-subsonic/internal/config"
	"go-subsonic/internal/core"
	"go-subsonic/internal/player"
	"go-subsonic/internal/ui"
)

type programRunner interface {
	Run() (tea.Model, error)
}

var (
	loadConfig    = config.Load
	newPlayer     = func() (core.AudioPlayer, error) { return player.NewPlayer() }
	buildProvider = app.BuildProvider
	newTeaProgram = func(m tea.Model, opts ...tea.ProgramOption) programRunner {
		return tea.NewProgram(m, opts...)
	}
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	p, err := newPlayer()
	if err != nil {
		return fmt.Errorf("error initializing audio: %w", err)
	}

	provider, client := buildProvider(cfg)
	if client != nil {
		if err := client.Ping(ctx); err != nil {
			return fmt.Errorf("error connecting to subsonic server: %w", err)
		}
	}

	program := newTeaProgram(ui.NewModel(provider, p), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("program failed: %w", err)
	}

	return nil
}
