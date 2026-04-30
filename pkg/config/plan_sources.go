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

// GitRootFile discovers an optional config file relative to the Git worktree
// containing the current process working directory. Discovery intentionally
// ignores Git's local repository environment variables (the variables reported
// by `git rev-parse --local-env-vars`, such as GIT_DIR and GIT_WORK_TREE) so
// callers launched from Git hooks still discover the repository for their cwd
// rather than the repository whose hook invoked them. General Git configuration
// variables such as GIT_SSH_COMMAND and GIT_TRACE are preserved.
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

// gitRoot returns the top-level worktree for the current working directory.
// It scrubs only Git's local repository environment variables before invoking
// git so inherited hook state cannot override cwd-based repository discovery.
func gitRoot(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel")
	cmd.Env = scrubGitLocalEnv(os.Environ())
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func scrubGitLocalEnv(env []string) []string {
	ret := make([]string, 0, len(env))
	for _, entry := range env {
		key, _, ok := strings.Cut(entry, "=")
		if ok && gitLocalEnvVars[key] {
			continue
		}
		ret = append(ret, entry)
	}
	return ret
}

var gitLocalEnvVars = map[string]bool{
	"GIT_ALTERNATE_OBJECT_DIRECTORIES": true,
	"GIT_CONFIG":                       true,
	"GIT_CONFIG_COUNT":                 true,
	"GIT_CONFIG_PARAMETERS":            true,
	"GIT_DIR":                          true,
	"GIT_GRAFT_FILE":                   true,
	"GIT_IMPLICIT_WORK_TREE":           true,
	"GIT_INDEX_FILE":                   true,
	"GIT_NO_REPLACE_OBJECTS":           true,
	"GIT_OBJECT_DIRECTORY":             true,
	"GIT_PREFIX":                       true,
	"GIT_REPLACE_REF_BASE":             true,
	"GIT_SHALLOW_FILE":                 true,
	"GIT_COMMON_DIR":                   true,
	"GIT_WORK_TREE":                    true,
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}
