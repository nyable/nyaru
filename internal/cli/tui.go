package cli

import (
	"github.com/nyable/nyaru/internal/config"
	"github.com/nyable/nyaru/internal/tui"
)

// runTUI fills in the persistent nyaru settings every TUI entry point needs.
func runTUI(opts tui.Options) error {
	settings := currentSettings()
	opts.Settings = &settings
	opts.SaveSettings = saveSettings
	return tui.Run(opts)
}

func currentSettings() tui.SettingsValues {
	cfg := config.Get()
	return tui.SettingsValues{
		Mode:          cfg.Mode,
		ConfirmBatch:  cfg.ConfirmBatch,
		StatusTimeout: cfg.StatusTimeoutSeconds(),
	}
}

func saveSettings(values tui.SettingsValues) error {
	cfg := config.Get()
	cfg.Mode = values.Mode
	cfg.ConfirmBatch = values.ConfirmBatch
	cfg.SetStatusTimeout(values.StatusTimeout)
	return cfg.Save()
}
