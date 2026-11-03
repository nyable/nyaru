package paths

import (
	"os"
	"path/filepath"
	"strings"
)

const appDir = "nyaru"

// ConfigDir follows the XDG convention used by Scoop and other developer
// tools on every platform:
//   - $XDG_CONFIG_HOME/nyaru when XDG_CONFIG_HOME is set
//   - ~/.config/nyaru otherwise
func ConfigDir() (string, error) {
	if base := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); base != "" && filepath.IsAbs(base) {
		return filepath.Join(base, appDir), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", appDir), nil
}

func ConfigFile() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// LegacyConfigFile is the original ~/.nyaru location. It is only used to
// migrate existing settings when the new file does not exist yet.
func LegacyConfigFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".nyaru", "config.yaml"), nil
}

// PlatformConfigFile is the brief %AppData%\nyaru location used before the
// project standardized on ~/.config. It is kept as a migration source.
func PlatformConfigFile() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appDir, "config.yaml"), nil
}

func SelfUpdateLog() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "self-update.log"), nil
}
