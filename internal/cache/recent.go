package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// RecentProject represents a recently used project
type RecentProject struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	LastUsed time.Time `json:"last_used"`
}

// RecentProjects manages the list of recently used projects
type RecentProjects struct {
	Projects []RecentProject `json:"projects"`
}

// getCachePath returns the path to the recent projects cache file
func getCachePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cacheDir := filepath.Join(home, ".cache", "sesh")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, "recent.json"), nil
}

// Load reads the recent projects from cache
func Load() (*RecentProjects, error) {
	cachePath, err := getCachePath()
	if err != nil {
		return &RecentProjects{Projects: []RecentProject{}}, nil
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &RecentProjects{Projects: []RecentProject{}}, nil
		}
		return nil, err
	}

	var recent RecentProjects
	if err := json.Unmarshal(data, &recent); err != nil {
		return &RecentProjects{Projects: []RecentProject{}}, nil
	}

	return &recent, nil
}

// Save writes the recent projects to cache
func (r *RecentProjects) Save() error {
	cachePath, err := getCachePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}

// Add records a project as recently used
func (r *RecentProjects) Add(name, path string) {
	// Remove if already exists
	for i, p := range r.Projects {
		if p.Path == path {
			r.Projects = append(r.Projects[:i], r.Projects[i+1:]...)
			break
		}
	}

	// Add to front
	r.Projects = append([]RecentProject{{
		Name:     name,
		Path:     path,
		LastUsed: time.Now(),
	}}, r.Projects...)

	// Keep only top 3
	if len(r.Projects) > 3 {
		r.Projects = r.Projects[:3]
	}
}

// GetTop3 returns the top 3 most recently used projects
func (r *RecentProjects) GetTop3() []RecentProject {
	if len(r.Projects) > 3 {
		return r.Projects[:3]
	}
	return r.Projects
}
