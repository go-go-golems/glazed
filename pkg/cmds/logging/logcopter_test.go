package logging

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-go-golems/logcopter/pkg/logcopter"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func TestParseAreaOverridesAcceptsRepeatedCommaColonAndEquals(t *testing.T) {
	got, err := ParseAreaOverrides([]string{"app.view:debug", "app.db=warn,lib.parser=trace"})
	if err != nil {
		t.Fatalf("ParseAreaOverrides returned error: %v", err)
	}
	want := map[string]string{"app.view": "debug", "app.db": "warn", "lib.parser": "trace"}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("ParseAreaOverrides()[%q] = %q, want %q (all: %#v)", k, got[k], v, got)
		}
	}
}

func TestParseAreaOverridesRejectsMalformed(t *testing.T) {
	if _, err := ParseAreaOverrides([]string{"app.view"}); err == nil {
		t.Fatalf("expected malformed override error")
	}
	if _, err := ParseAreaOverrides([]string{"app.view:"}); err == nil {
		t.Fatalf("expected empty level error")
	}
}

func TestLoadLoggingSettingsFileSupportsWrappedAndDirectShapes(t *testing.T) {
	dir := t.TempDir()
	wrappedPath := filepath.Join(dir, "wrapped.yaml")
	if err := os.WriteFile(wrappedPath, []byte(`logging:
  log-level: debug
  log-format: json
  areas:
    app.view: trace
  strict-log-areas: true
`), 0o644); err != nil {
		t.Fatalf("write wrapped profile: %v", err)
	}
	wrapped, err := LoadLoggingSettingsFile(wrappedPath)
	if err != nil {
		t.Fatalf("LoadLoggingSettingsFile wrapped returned error: %v", err)
	}
	if wrapped.LogLevel != "debug" || wrapped.LogFormat != "json" || wrapped.Areas["app.view"] != "trace" || !wrapped.StrictAreas {
		t.Fatalf("unexpected wrapped settings: %#v", wrapped)
	}

	directPath := filepath.Join(dir, "direct.yaml")
	if err := os.WriteFile(directPath, []byte(`level: warn
format: text
output: stdout
caller: true
strict_areas: true
areas:
  app.db: error
`), 0o644); err != nil {
		t.Fatalf("write direct profile: %v", err)
	}
	direct, err := LoadLoggingSettingsFile(directPath)
	if err != nil {
		t.Fatalf("LoadLoggingSettingsFile direct returned error: %v", err)
	}
	if direct.LogLevel != "warn" || direct.LogFormat != "text" || !direct.LogToStdout || !direct.WithCaller || !direct.StrictAreas || direct.Areas["app.db"] != "error" {
		t.Fatalf("unexpected direct settings: %#v", direct)
	}
}

func TestInitLoggerFromCobraUsesProfileValuesWhenFlagsAreNotChanged(t *testing.T) {
	_ = logcopter.Package("app.profile")
	dir := t.TempDir()
	logFile := filepath.Join(dir, "profile.log")
	profilePath := filepath.Join(dir, "profile.yaml")
	if err := os.WriteFile(profilePath, []byte(`logging:
  log-level: warn
  log-format: json
  log-file: `+logFile+`
  areas:
    app.profile: error
`), 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	cmd := &cobra.Command{Use: "test"}
	if err := AddLoggingSectionToRootCommand(cmd, "test"); err != nil {
		t.Fatalf("AddLoggingSectionToRootCommand returned error: %v", err)
	}
	if err := cmd.ParseFlags([]string{"--log-config", profilePath}); err != nil {
		t.Fatalf("ParseFlags returned error: %v", err)
	}
	if err := InitLoggerFromCobra(cmd); err != nil {
		t.Fatalf("InitLoggerFromCobra returned error: %v", err)
	}

	if got := zlog.Logger.GetLevel(); got != zerolog.WarnLevel {
		t.Fatalf("global logger level = %s, want warn", got)
	}
	if got := logcopter.EffectiveLevel("app.other").String(); got != "warn" {
		t.Fatalf("default logcopter level = %s, want warn", got)
	}
	if got := logcopter.EffectiveLevel("app.profile").String(); got != "error" {
		t.Fatalf("app.profile level = %s, want error", got)
	}

	zlog.Logger.Warn().Str("source", "profile").Msg("profile json smoke")
	b, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("read profile log file: %v", err)
	}
	line := strings.TrimSpace(string(b))
	if line == "" {
		t.Fatalf("expected profile log file to contain one JSON log line")
	}
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("profile log output was not JSON: %v\nline: %s", err, line)
	}
	if entry["level"] != "warn" || entry["message"] != "profile json smoke" || entry["source"] != "profile" {
		t.Fatalf("unexpected profile log entry: %#v", entry)
	}
}

func TestInitLoggerFromCobraMergesProfilesThenCLIOverrides(t *testing.T) {
	_ = logcopter.Package("app.view")
	_ = logcopter.Package("app.db")
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "profile.yaml")
	if err := os.WriteFile(profilePath, []byte(`level: info
areas:
  app.view: trace
  app.db: error
`), 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	cmd := &cobra.Command{Use: "test"}
	if err := AddLoggingSectionToRootCommand(cmd, "test"); err != nil {
		t.Fatalf("AddLoggingSectionToRootCommand returned error: %v", err)
	}
	if err := cmd.ParseFlags([]string{"--log-config", profilePath, "--log-area", "app.db=warn"}); err != nil {
		t.Fatalf("ParseFlags returned error: %v", err)
	}
	if err := InitLoggerFromCobra(cmd); err != nil {
		t.Fatalf("InitLoggerFromCobra returned error: %v", err)
	}

	if got := logcopter.EffectiveLevel("app.view").String(); got != "trace" {
		t.Fatalf("app.view level = %s, want trace", got)
	}
	if got := logcopter.EffectiveLevel("app.db").String(); got != "warn" {
		t.Fatalf("app.db level = %s, want warn", got)
	}
}

func TestInitEarlyLoggingFromArgsParsesProfilesAndAreaOverrides(t *testing.T) {
	_ = logcopter.Package("app.early")
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "profile.yaml")
	if err := os.WriteFile(profilePath, []byte(`level: info
areas:
  app.early: error
`), 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}

	if err := InitEarlyLoggingFromArgs([]string{"run", "--unknown", "x", "--log-config", profilePath, "--log-area", "app.early=debug"}, "test"); err != nil {
		t.Fatalf("InitEarlyLoggingFromArgs returned error: %v", err)
	}
	if got := logcopter.EffectiveLevel("app.early").String(); got != "debug" {
		t.Fatalf("app.early level = %s, want debug", got)
	}
}
