package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/adamflitney/sesh/internal/finder"
)

// SessionExists checks if a tmux session with the given name exists
func SessionExists(name string) (bool, error) {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exit code 1 means session doesn't exist
			if exitError.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, fmt.Errorf("error checking session: %w", err)
	}
	return true, nil
}

// SanitizeSessionName converts a project name to a valid tmux session name
// Replaces spaces and special characters with hyphens
func SanitizeSessionName(name string) string {
	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	sanitized := reg.ReplaceAllString(name, "-")

	// Remove leading/trailing hyphens
	sanitized = strings.Trim(sanitized, "-")

	// Convert to lowercase for consistency
	sanitized = strings.ToLower(sanitized)

	return sanitized
}

// tmuxCmd creates an exec.Command for tmux with TMUX env var removed
// This allows running tmux commands from within a tmux session (e.g., popup)
func tmuxCmd(args ...string) *exec.Cmd {
	cmd := exec.Command("tmux", args...)
	// Filter out TMUX from environment to allow nested tmux commands
	env := os.Environ()
	filteredEnv := make([]string, 0, len(env))
	for _, e := range env {
		if !strings.HasPrefix(e, "TMUX=") {
			filteredEnv = append(filteredEnv, e)
		}
	}
	cmd.Env = filteredEnv
	return cmd
}

// CreateSession creates a new tmux session with three windows
func CreateSession(project finder.Project) error {
	sessionName := SanitizeSessionName(project.Name)

	// Create new session with first window (neovim)
	cmd := tmuxCmd("new-session", "-d", "-s", sessionName, "-c", project.Path, "-n", "neovim")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	// Send nvim command to first window (use window name instead of index)
	cmd = tmuxCmd("send-keys", "-t", sessionName+":neovim", "nvim .", "Enter")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send nvim command: %w", err)
	}

	// Create second window (opencode)
	cmd = tmuxCmd("new-window", "-t", sessionName, "-n", "opencode", "-c", project.Path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create opencode window: %w", err)
	}

	// Send opencode command to second window
	// Start with --port flag so opencode.nvim can connect to it
	cmd = tmuxCmd("send-keys", "-t", sessionName+":opencode", "opencode --port 0 .", "Enter")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send opencode command: %w", err)
	}

	// Create third window (zsh)
	cmd = tmuxCmd("new-window", "-t", sessionName, "-n", "zsh", "-c", project.Path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create zsh window: %w", err)
	}

	// Select the first window (use window name)
	cmd = tmuxCmd("select-window", "-t", sessionName+":neovim")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to select first window: %w", err)
	}

	return nil
}

// AttachSession attaches to an existing tmux session
func AttachSession(sessionName string) error {
	// We need to replace the current process with tmux
	// This is done using syscall.Exec
	tmuxPath, err := exec.LookPath("tmux")
	if err != nil {
		return fmt.Errorf("tmux not found in PATH: %w", err)
	}

	args := []string{"tmux", "attach-session", "-t", sessionName}
	env := os.Environ()

	// Replace current process with tmux
	return syscall.Exec(tmuxPath, args, env)
}

// GetOrCreateSession creates a new session if it doesn't exist, or attaches to an existing one
func GetOrCreateSession(project finder.Project) error {
	sessionName := SanitizeSessionName(project.Name)

	// Check if tmux is installed
	if _, err := exec.LookPath("tmux"); err != nil {
		return fmt.Errorf("tmux is not installed. Please install tmux first")
	}

	// Check if session exists
	exists, err := SessionExists(sessionName)
	if err != nil {
		return err
	}

	if exists {
		fmt.Printf("Attaching to existing session '%s'...\n", sessionName)
	} else {
		fmt.Printf("Creating new session '%s'...\n", sessionName)
		if err := CreateSession(project); err != nil {
			return err
		}
	}

	// Check if we're inside tmux
	if os.Getenv("TMUX") != "" {
		// Switch to session instead of attaching
		return SwitchSession(sessionName)
	}

	// Attach to session (this will replace the current process)
	return AttachSession(sessionName)
}

// SwitchSession switches to an existing tmux session (used when already inside tmux)
func SwitchSession(sessionName string) error {
	// Check if we have a target client from the environment (set by Raycast script)
	targetClient := os.Getenv("SESH_TARGET_CLIENT")

	var cmd *exec.Cmd
	if targetClient != "" {
		// Target the specific client passed from the popup launcher
		cmd = exec.Command("tmux", "switch-client", "-t", sessionName, "-c", targetClient)
	} else {
		// Default: switch current client
		cmd = exec.Command("tmux", "switch-client", "-t", sessionName)
	}

	return cmd.Run()
}

// ListSessions returns a list of active tmux session names
func ListSessions() ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var sessions []string
	for _, line := range lines {
		if line != "" {
			sessions = append(sessions, line)
		}
	}
	return sessions, nil
}
