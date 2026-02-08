package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	ProjectDirectories []string `mapstructure:"project_directories"`
}

const (
	configDir  = ".config/sesh"
	configFile = "config"
	configType = "yaml"
)

// LoadConfig loads the configuration from the config file or creates a default one
func LoadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	viper.SetConfigName(configFile)
	viper.SetConfigType(configType)
	viper.AddConfigPath(configPath)

	// Set defaults
	viper.SetDefault("project_directories", []string{"~/dev"})

	// Try to read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, create default
			if err := createDefaultConfig(configPath); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
			// Read the newly created config
			if err := viper.ReadInConfig(); err != nil {
				return nil, fmt.Errorf("failed to read config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand home directory in paths
	for i, dir := range cfg.ProjectDirectories {
		cfg.ProjectDirectories[i] = expandPath(dir)
	}

	return &cfg, nil
}

// getConfigPath returns the path to the config directory
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir), nil
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(configPath string) error {
	configFilePath := filepath.Join(configPath, configFile+"."+configType)

	defaultConfig := `# Sesh Configuration
# List directories where your Git projects are located

project_directories:
  - ~/dev
`

	if err := os.WriteFile(configFilePath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if len(path) == 1 {
		return home
	}

	return filepath.Join(home, path[1:])
}

// GetConfigFilePath returns the full path to the config file
func GetConfigFilePath() (string, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(configPath, configFile+"."+configType), nil
}
