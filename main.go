package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/adamflitney/sesh/internal/cache"
	"github.com/adamflitney/sesh/internal/config"
	"github.com/adamflitney/sesh/internal/finder"
	"github.com/adamflitney/sesh/internal/tmux"
	"github.com/adamflitney/sesh/internal/ui"
	"github.com/adamflitney/sesh/internal/zoxide"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "list":
			return runList(os.Args[2:])
		case "connect":
			if len(os.Args) < 3 {
				return fmt.Errorf("usage: sesh connect <project-name>")
			}
			return runConnect(strings.Join(os.Args[2:], " "))
		case "switch":
			return runSwitch()
		case "help", "-h", "--help":
			printUsage()
			return nil
		case "version", "-v", "--version":
			fmt.Println("sesh v0.2.0")
			return nil
		default:
			// Unknown subcommand - treat as project name for quick connect
			return runConnect(strings.Join(os.Args[1:], " "))
		}
	}

	// Default: interactive TUI
	return runInteractive()
}

func printUsage() {
	fmt.Println(`sesh - Smart tmux session manager

Usage:
  sesh                  Interactive project picker (TUI)
  sesh list             List all projects (one per line)
  sesh list -t          List only active tmux sessions
  sesh list --json      List projects as JSON
  sesh connect <name>   Connect to project by name
  sesh switch           Interactive picker for active sessions only
  sesh <name>           Quick connect (same as 'sesh connect <name>')
  sesh help             Show this help
  sesh version          Show version

Examples:
  sesh                  # Open interactive picker
  sesh yoto-club-api    # Connect directly to project
  sesh list | fzf       # Use with external fuzzy finder
  sesh switch           # Quick switch between open projects`)
}

func runList(args []string) error {
	// Parse flags
	tmuxOnly := false
	jsonOutput := false
	for _, arg := range args {
		switch arg {
		case "-t", "--tmux":
			tmuxOnly = true
		case "--json":
			jsonOutput = true
		}
	}

	if tmuxOnly {
		return listTmuxSessions()
	}

	return listProjects(jsonOutput)
}

func listProjects(jsonOutput bool) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	projects, err := finder.FindGitProjects(cfg.ProjectDirectories)
	if err != nil {
		return err
	}

	if jsonOutput {
		fmt.Println("[")
		for i, p := range projects {
			comma := ","
			if i == len(projects)-1 {
				comma = ""
			}
			fmt.Printf("  {\"name\": %q, \"path\": %q}%s\n", p.Name, p.Path, comma)
		}
		fmt.Println("]")
		return nil
	}

	for _, p := range projects {
		fmt.Println(p.Name)
	}
	return nil
}

func listTmuxSessions() error {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		// tmux not running or no sessions
		return nil
	}

	sessions := strings.TrimSpace(string(output))
	if sessions != "" {
		fmt.Println(sessions)
	}
	return nil
}

func runConnect(name string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	projects, err := finder.FindGitProjects(cfg.ProjectDirectories)
	if err != nil {
		return err
	}

	// First, try exact match (case-insensitive)
	nameLower := strings.ToLower(name)
	for _, p := range projects {
		if strings.ToLower(p.Name) == nameLower {
			// Record in recent history and zoxide
			recent, _ := cache.Load()
			if recent != nil {
				recent.Add(p.Name, p.Path)
				_ = recent.Save()
			}
			_ = zoxide.Add(p.Path) // Track in zoxide for frecency
			return tmux.GetOrCreateSession(p)
		}
	}

	// Try sanitized session name match
	sanitizedName := tmux.SanitizeSessionName(name)
	for _, p := range projects {
		if tmux.SanitizeSessionName(p.Name) == sanitizedName {
			recent, _ := cache.Load()
			if recent != nil {
				recent.Add(p.Name, p.Path)
				_ = recent.Save()
			}
			_ = zoxide.Add(p.Path) // Track in zoxide for frecency
			return tmux.GetOrCreateSession(p)
		}
	}

	// Try prefix match as fallback
	for _, p := range projects {
		if strings.HasPrefix(strings.ToLower(p.Name), nameLower) {
			recent, _ := cache.Load()
			if recent != nil {
				recent.Add(p.Name, p.Path)
				_ = recent.Save()
			}
			_ = zoxide.Add(p.Path) // Track in zoxide for frecency
			return tmux.GetOrCreateSession(p)
		}
	}

	return fmt.Errorf("project not found: %s\n\nAvailable projects:\n%s",
		name, getProjectList(projects))
}

func getProjectList(projects []finder.Project) string {
	var names []string
	for _, p := range projects {
		names = append(names, "  - "+p.Name)
	}
	return strings.Join(names, "\n")
}

func runSwitch() error {
	// Get active tmux sessions
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}:#{session_path}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("no active tmux sessions")
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return fmt.Errorf("no active tmux sessions")
	}

	// Convert to Project structs for the UI
	var sessions []finder.Project
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		name := parts[0]
		path := ""
		if len(parts) > 1 {
			path = parts[1]
		}
		sessions = append(sessions, finder.Project{
			Name: name,
			Path: path,
		})
	}

	// Display session selector UI
	selectedSession, err := ui.SelectProject(sessions)
	if err != nil {
		return fmt.Errorf("failed to select session: %w", err)
	}

	if selectedSession == nil {
		return nil
	}

	// Switch to selected session
	return tmux.SwitchSession(selectedSession.Name)
}

func runInteractive() error {
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

	// Record the selected project in recent history and zoxide
	recent, _ := cache.Load()
	if recent != nil {
		recent.Add(selectedProject.Name, selectedProject.Path)
		_ = recent.Save() // Ignore errors for cache saves
	}
	_ = zoxide.Add(selectedProject.Path) // Track in zoxide for frecency

	// Create or attach to tmux session
	if err := tmux.GetOrCreateSession(*selectedProject); err != nil {
		return fmt.Errorf("failed to manage tmux session: %w", err)
	}

	return nil
}
