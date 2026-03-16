package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Mode string `yaml:"mode"` // sfsu, scoop, or auto
}

var currentConfig *Config

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Mode: "auto",
	}
}

func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".nyaru"), nil
}

// LoadConfig reads the configuration from ~/.nyaru/config.yaml
func LoadConfig() error {
	dir, err := GetConfigDir()
	if err != nil {
		return err
	}
	
	configPath := filepath.Join(dir, "config.yaml")
	
	currentConfig = DefaultConfig()
	
	data, err := os.ReadFile(configPath)
	if err == nil {
		if err := yaml.Unmarshal(data, currentConfig); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	} else if os.IsNotExist(err) {
		// create default config if not exist
		_ = os.MkdirAll(dir, 0755)
		data, _ = yaml.Marshal(currentConfig)
		_ = os.WriteFile(configPath, data, 0644)
	} else {
		return fmt.Errorf("failed to read config: %w", err)
	}
	
	if currentConfig.Mode != "sfsu" && currentConfig.Mode != "scoop" && currentConfig.Mode != "auto" {
		currentConfig.Mode = "auto"
	}

	return nil
}

// GetActiveMode resolves the current active mode.
// If mode is "auto", it checks for "sfsu" and falls back to "scoop"
func GetActiveMode() string {
	if currentConfig == nil {
		_ = LoadConfig()
	}
	
	if currentConfig.Mode == "sfsu" || currentConfig.Mode == "scoop" {
		return currentConfig.Mode
	}
	
	// auto mode
	_, err := exec.LookPath("sfsu")
	if err == nil {
		return "sfsu"
	}
	return "scoop"
}
