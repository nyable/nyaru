package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nyable/nyaru/internal/paths"
	"gopkg.in/yaml.v3"
)

// Config is the persistent nyaru configuration.
type Config struct {
	Mode          string `yaml:"mode"`           // auto | sfsu | scoop
	Language      string `yaml:"language"`       // zh | en (reserved)
	ConfirmBatch  bool   `yaml:"confirm_batch"`  // ask before batch operations
	StatusTimeout *int   `yaml:"status_timeout"` // seconds; 0 means never clear
}

var current *Config

const DefaultStatusTimeout = 5

func DefaultConfig() *Config {
	timeout := DefaultStatusTimeout
	return &Config{
		Mode:          "auto",
		Language:      "zh",
		ConfirmBatch:  true,
		StatusTimeout: &timeout,
	}
}

// Path returns the location of nyaru's own configuration file.
func Path() (string, error) {
	return paths.ConfigFile()
}

// Load reads the XDG configuration file, creating a default file when needed.
// Previous ~/.nyaru and %AppData%\nyaru locations are migrated on first run.
func Load() (*Config, error) {
	path, err := paths.ConfigFile()
	if err != nil {
		return nil, err
	}
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	migrated := false
	if os.IsNotExist(err) {
		for _, candidate := range []func() (string, error){
			paths.PlatformConfigFile,
			paths.LegacyConfigFile,
		} {
			legacyPath, legacyErr := candidate()
			if legacyErr != nil || legacyPath == path {
				continue
			}
			if legacyData, readErr := os.ReadFile(legacyPath); readErr == nil {
				data = legacyData
				err = nil
				migrated = true
				break
			}
		}
	}
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("解析配置文件失败: %w", err)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
		if data, err := yaml.Marshal(cfg); err == nil {
			_ = os.WriteFile(path, data, 0o644)
		}
	} else {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	cfg.normalize()
	if migrated {
		if err := cfg.Save(); err != nil {
			return nil, err
		}
	}
	current = cfg
	return cfg, nil
}

func (c *Config) normalize() {
	switch strings.ToLower(strings.TrimSpace(c.Mode)) {
	case "sfsu", "scoop", "auto":
		c.Mode = strings.ToLower(strings.TrimSpace(c.Mode))
	default:
		c.Mode = "auto"
	}
	if c.Language == "" {
		c.Language = "zh"
	}
	if c.StatusTimeout == nil {
		timeout := DefaultStatusTimeout
		c.StatusTimeout = &timeout
	} else if *c.StatusTimeout < 0 {
		*c.StatusTimeout = DefaultStatusTimeout
	} else if *c.StatusTimeout > 3600 {
		*c.StatusTimeout = 3600
	}
}

// StatusTimeoutSeconds returns the configured status-message lifetime.
// A value of zero means the message stays until the next screen transition.
func (c *Config) StatusTimeoutSeconds() int {
	if c == nil || c.StatusTimeout == nil {
		return DefaultStatusTimeout
	}
	return *c.StatusTimeout
}

func (c *Config) SetStatusTimeout(seconds int) {
	c.StatusTimeout = &seconds
	c.normalize()
}

// SettingValue returns the current value of a user-facing nyaru setting.
func (c *Config) SettingValue(key string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "mode":
		return c.Mode, nil
	case "confirm_batch":
		return strconv.FormatBool(c.ConfirmBatch), nil
	case "status_timeout":
		return strconv.Itoa(c.StatusTimeoutSeconds()), nil
	default:
		return "", fmt.Errorf("未知设置项: %s", key)
	}
}

// SetSetting validates and applies one user-facing nyaru setting.
func (c *Config) SetSetting(key, value string) error {
	key = strings.ToLower(strings.TrimSpace(key))
	value = strings.TrimSpace(value)
	switch key {
	case "mode":
		mode := strings.ToLower(value)
		switch mode {
		case "auto", "sfsu", "scoop":
			c.Mode = mode
			return nil
		default:
			return fmt.Errorf("mode 只能是 auto、sfsu 或 scoop")
		}
	case "confirm_batch":
		parsed, err := parseBool(value)
		if err != nil {
			return err
		}
		c.ConfirmBatch = parsed
		return nil
	case "status_timeout":
		if strings.EqualFold(value, "never") || strings.EqualFold(value, "off") {
			c.SetStatusTimeout(0)
			return nil
		}
		seconds, err := strconv.Atoi(value)
		if err != nil || seconds < 0 || seconds > 3600 {
			return fmt.Errorf("status_timeout 必须是 0-3600 之间的秒数，或 never")
		}
		c.SetStatusTimeout(seconds)
		return nil
	default:
		return fmt.Errorf("未知设置项: %s", key)
	}
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on", "enable", "enabled":
		return true, nil
	case "0", "false", "no", "n", "off", "disable", "disabled":
		return false, nil
	default:
		return false, fmt.Errorf("布尔值只能是 true/false、yes/no 或 on/off")
	}
}

// SettingsKeys lists the settings exposed by the CLI and TUI.
func SettingsKeys() []string {
	return []string{"mode", "confirm_batch", "status_timeout"}
}

// Get returns the loaded config, loading it lazily if necessary.
func Get() *Config {
	if current == nil {
		if cfg, err := Load(); err == nil {
			return cfg
		}
		return DefaultConfig()
	}
	return current
}

// ResolveMode resolves auto to sfsu when the executable is available.
// LookPath is injectable for tests.
func ResolveMode(mode string, lookPath func(string) (string, error)) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "sfsu":
		return "sfsu"
	case "scoop":
		return "scoop"
	}
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	if _, err := lookPath("sfsu"); err == nil {
		return "sfsu"
	}
	return "scoop"
}

// ActiveMode resolves the mode using the real PATH.
func (c *Config) ActiveMode() string {
	if c == nil {
		c = Get()
	}
	return ResolveMode(c.Mode, exec.LookPath)
}

// Save writes the configuration back to disk.
func (c *Config) Save() error {
	path, err := paths.ConfigFile()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
