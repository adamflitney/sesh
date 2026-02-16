package finder

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/adamflitney/sesh/internal/cache"
	"github.com/adamflitney/sesh/internal/zoxide"
)

// Project represents a Git project
type Project struct {
	Name  string
	Path  string
	Score float64 // Combined score from zoxide + recency
}

// FindGitProjects searches for Git repositories in the given directories
// Projects are sorted by frecency (frequency + recency) using zoxide scores
// and the internal recent projects cache
func FindGitProjects(directories []string) ([]Project, error) {
	projectsMap := make(map[string]Project) // Use map to avoid duplicates

	// Directories to skip for performance
	skipDirs := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		"target":       true,
		"build":        true,
		"dist":         true,
		".next":        true,
		".cache":       true,
		"__pycache__":  true,
		".venv":        true,
		"venv":         true,
	}

	for _, dir := range directories {
		// Check if directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: directory does not exist: %s\n", dir)
			continue
		}

		// Walk the directory
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				// Skip directories we can't read
				return filepath.SkipDir
			}

			// Skip common large directories for performance
			if d.IsDir() && skipDirs[d.Name()] {
				return filepath.SkipDir
			}

			// If this is a .git directory, the parent is a Git project
			if d.IsDir() && d.Name() == ".git" {
				projectPath := filepath.Dir(path)
				projectName := filepath.Base(projectPath)

				// Store project (map prevents duplicates)
				projectsMap[projectPath] = Project{
					Name: projectName,
					Path: projectPath,
				}

				// Don't descend into .git directory
				return filepath.SkipDir
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error walking directory %s: %w", dir, err)
		}
	}

	// Convert map to slice
	projects := make([]Project, 0, len(projectsMap))
	for _, project := range projectsMap {
		projects = append(projects, project)
	}

	// Apply frecency scoring
	projects = applyFrecencyScores(projects)

	return projects, nil
}

// applyFrecencyScores combines zoxide scores with recent cache for smart ordering
func applyFrecencyScores(projects []Project) []Project {
	// Get zoxide scores
	zoxideScores, _ := zoxide.GetScores()

	// Get recent projects from cache
	recentProjects, _ := cache.Load()
	recentPaths := make(map[string]int) // path -> recency rank (1 = most recent)
	if recentProjects != nil {
		for i, rp := range recentProjects.GetTop3() {
			recentPaths[rp.Path] = i + 1
		}
	}

	// Calculate combined scores
	for i := range projects {
		score := 0.0

		// Add zoxide score (frecency from all shell usage)
		if zoxideScores != nil {
			if zs, ok := zoxideScores[projects[i].Path]; ok {
				score += zs
			}
		}

		// Boost recent projects heavily (they should appear first)
		// Rank 1 (most recent) gets +10000, rank 2 gets +9000, rank 3 gets +8000
		if rank, ok := recentPaths[projects[i].Path]; ok {
			score += float64(11000 - (rank * 1000))
		}

		projects[i].Score = score
	}

	// Sort by score (highest first), then by name for ties
	sort.Slice(projects, func(i, j int) bool {
		if projects[i].Score != projects[j].Score {
			return projects[i].Score > projects[j].Score
		}
		return projects[i].Name < projects[j].Name
	})

	return projects
}
