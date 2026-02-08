package main

import (
	"fmt"
	"os"

	"github.com/adamflitney/sesh/internal/config"
	"github.com/adamflitney/sesh/internal/finder"
	"github.com/adamflitney/sesh/internal/tmux"
	"github.com/adamflitney/sesh/internal/ui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Find all Git projects
	projects, err := finder.FindGitProjects(cfg.ProjectDirectories)
	if err != nil {
		return fmt.Errorf("failed to find projects: %w", err)
	}

	// If no projects found, show helpful message
	if len(projects) == 0 {
		configPath, _ := config.GetConfigFilePath()
		return fmt.Errorf("no Git projects found in configured directories.\n\nConfigured directories:\n%v\n\nEdit your config at: %s",
			cfg.ProjectDirectories, configPath)
	}

	// Display project selector UI
	selectedProject, err := ui.SelectProject(projects)
	if err != nil {
		return fmt.Errorf("failed to select project: %w", err)
	}

	// If user quit without selecting, exit gracefully
	if selectedProject == nil {
		return nil
	}

	// Create or attach to tmux session
	if err := tmux.GetOrCreateSession(*selectedProject); err != nil {
		return fmt.Errorf("failed to manage tmux session: %w", err)
	}

	return nil
}
