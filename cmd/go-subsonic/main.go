package main

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"go-subsonic/internal/config"
	"go-subsonic/internal/player"
	"go-subsonic/internal/subsonic"
	"go-subsonic/internal/ui"
)

func main() {
	// Phase 1: Config & Connection Test
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		fmt.Println("Please create a config.yaml file (see docs/implementation_plan.md)")
	}

	p, err := player.NewPlayer()
	if err != nil {
		fmt.Printf("Error initializing audio: %v\n", err)
		os.Exit(1)
	}

	var client *subsonic.Client
	if cfg != nil {
		fmt.Printf("Config loaded. Server: %s\n", cfg.Server.URL)
		client = subsonic.NewClient(cfg.Server.URL, cfg.Server.Username, cfg.Server.Password)
		err := client.Ping(context.Background())
		if err != nil {
			fmt.Printf("Failed to ping server: %v\n", err)
		} else {
			fmt.Println("Successfully connected to Subsonic server!")
		}
	} else {
		fmt.Println("No configuration found. Skipping server connection test.")
	}
	
	fmt.Println("Starting TUI...")
	// If client is nil (config failed), NewModel handles it gracefully (mostly)
	program := tea.NewProgram(ui.NewModel(client, p))
	if _, err := program.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}