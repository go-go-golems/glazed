package config

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	userConfigDirFunc = os.UserConfigDir
	userHomeDirFunc   = os.UserHomeDir
	getwdFunc         = os.Getwd
	gitRootFunc       = gitRoot
)

func SystemAppConfig(appName string) SourceSpec {
	appName = strings.TrimSpace(appName)
	return SourceSpec{
		Name:       "system-app-config",
		Layer:      LayerSystem,
		SourceKind: "app-config",
		Optional:   true,
		Discover: func(ctx context.Context) ([]string, error) {
			if appName == "" {
				return nil, nil
			}
			p := filepath.Join("/etc", appName, "config.yaml")
			if !fileExists(p) {
				return nil, nil
			}
			return []string{p}, nil
		},
	}
}

func XDGAppConfig(appName string) SourceSpec {
	appName = strings.TrimSpace(appName)
	return SourceSpec{
		Name:       "xdg-app-config",
		Layer:      LayerUser,
		SourceKind: "app-config",
		Optional:   true,
		Discover: func(ctx context.Context) ([]string, error) {
			if appName == "" {
				return nil, nil
			}
			xdg, err := userConfigDirFunc()
			if err != nil || xdg == "" {
				return nil, nil
			}
			p := filepath.Join(xdg, appName, "config.yaml")
			if !fileExists(p) {
				return nil, nil
			}
			return []string{p}, nil
		},
	}
}

func HomeAppConfig(appName string) SourceSpec {
	appName = strings.TrimSpace(appName)
	return SourceSpec{
		Name:       "home-app-config",
		Layer:      LayerUser,
		SourceKind: "app-config",
		Optional:   true,
		Discover: func(ctx context.Context) ([]string, error) {
			if appName == "" {
				return nil, nil
			}
			home, err := userHomeDirFunc()
			if err != nil || home == "" {
				return nil, nil
			}
			p := filepath.Join(home, "."+appName, "config.yaml")
			if !fileExists(p) {
				return nil, nil
			}
			return []string{p}, nil
		},
	}
}

func ExplicitFile(path string) SourceSpec {
	path = strings.TrimSpace(path)
	return SourceSpec{
		Name:       "explicit-config-file",
		Layer:      LayerExplicit,
		SourceKind: "explicit-file",
		Discover: func(ctx context.Context) ([]string, error) {
			if path == "" {
				return nil, nil
			}
			if !fileExists(path) {
				return nil, os.ErrNotExist
			}
			return []string{path}, nil
		},
	}
}

func WorkingDirFile(name string) SourceSpec {
	name = strings.TrimSpace(name)
	return SourceSpec{
		Name:       "working-dir-file",
		Layer:      LayerCWD,
		SourceKind: "local-file",
		Optional:   true,
		Discover: func(ctx context.Context) ([]string, error) {
			if name == "" {
				return nil, nil
			}
			cwd, err := getwdFunc()
			if err != nil {
				return nil, err
			}
			p := filepath.Join(cwd, name)
			if !fileExists(p) {
				return nil, nil
			}
			return []string{p}, nil
		},
	}
}

func GitRootFile(name string) SourceSpec {
	name = strings.TrimSpace(name)
	return SourceSpec{
		Name:       "git-root-file",
		Layer:      LayerRepo,
		SourceKind: "local-file",
		Optional:   true,
		Discover: func(ctx context.Context) ([]string, error) {
			if name == "" {
				return nil, nil
			}
			root, err := gitRootFunc(ctx)
			if err != nil || root == "" {
				return nil, nil
			}
			p := filepath.Join(root, name)
			if !fileExists(p) {
				return nil, nil
			}
			return []string{p}, nil
		},
	}
}

func gitRoot(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}
