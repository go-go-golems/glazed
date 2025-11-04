package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveAppConfigPath returns the first existing config file path for the given app.
// Search order (when explicit == ""):
//  1. $XDG_CONFIG_HOME/<appName>/config.yaml (using os.UserConfigDir)
//  2. $HOME/.<appName>/config.yaml
//  3. /etc/<appName>/config.yaml
//
// If explicit is provided and exists, it is returned directly.
// If no candidate exists, returns "" with nil error.
func ResolveAppConfigPath(appName string, explicit string) (string, error) {
	if explicit != "" {
		if fileExists(explicit) {
			return explicit, nil
		}
		return "", fmt.Errorf("explicit config file not found: %s", explicit)
	}

	// 1) XDG config home
	if xdg, err := os.UserConfigDir(); err == nil && xdg != "" {
		p := filepath.Join(xdg, appName, "config.yaml")
		if fileExists(p) {
			return p, nil
		}
	}

	// 2) $HOME/.<appName>/config.yaml
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		p := filepath.Join(home, "."+appName, "config.yaml")
		if fileExists(p) {
			return p, nil
		}
	}

	// 3) /etc/<appName>/config.yaml
	p := filepath.Join("/etc", appName, "config.yaml")
	if fileExists(p) {
		return p, nil
	}

	return "", nil
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}
