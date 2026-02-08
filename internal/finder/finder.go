package finder

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Project represents a Git project
type Project struct {
	Name string
	Path string
}

// FindGitProjects searches for Git repositories in the given directories
func FindGitProjects(directories []string) ([]Project, error) {
	projectsMap := make(map[string]Project) // Use map to avoid duplicates

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

	// Sort projects by name for consistent display
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}
