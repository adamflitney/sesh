package zoxide

import (
	"os/exec"
	"strconv"
	"strings"
)

// Score represents a zoxide score for a path
type Score struct {
	Path  string
	Score float64
}

// IsAvailable checks if zoxide is installed
func IsAvailable() bool {
	_, err := exec.LookPath("zoxide")
	return err == nil
}

// GetScores returns zoxide scores for all tracked directories
func GetScores() (map[string]float64, error) {
	if !IsAvailable() {
		return nil, nil
	}

	// zoxide query -l -s returns paths with scores
	cmd := exec.Command("zoxide", "query", "-l", "-s")
	output, err := cmd.Output()
	if err != nil {
		// zoxide might not have any data yet
		return make(map[string]float64), nil
	}

	scores := make(map[string]float64)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		// Format: "score path" (e.g., "123.45 /Users/adam/Dev/project")
		parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
		if len(parts) != 2 {
			continue
		}

		score, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			continue
		}

		path := parts[1]
		scores[path] = score
	}

	return scores, nil
}

// Add adds a path to zoxide database
func Add(path string) error {
	if !IsAvailable() {
		return nil
	}

	cmd := exec.Command("zoxide", "add", path)
	return cmd.Run()
}

// GetScore returns the zoxide score for a specific path
func GetScore(path string) float64 {
	scores, err := GetScores()
	if err != nil || scores == nil {
		return 0
	}
	return scores[path]
}
