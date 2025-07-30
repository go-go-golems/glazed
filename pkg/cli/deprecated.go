package cli

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/spf13/cobra"
)

// Deprecated functions - use the unified CobraOption API instead

// BuildCobraCommandDualMode creates a cobra command that can run in both classic and glaze modes.
// Deprecated: Use BuildCobraCommandFromCommand(c, WithDualMode(true)) instead.
func BuildCobraCommandDualMode(
	c cmds.Command,
	opts ...DualModeOption,
) (*cobra.Command, error) {
	cfg := translateDualModeOptions(opts)
	return BuildCobraCommandFromCommand(c, cfg...)
}

// translateDualModeOptions converts legacy DualModeOption to new CobraOption
func translateDualModeOptions(opts []DualModeOption) []CobraOption {
	cfg := []CobraOption{WithDualMode(true)}
	
	for _, opt := range opts {
		switch v := opt.(type) {
		case glazeToggleFlagOption:
			cfg = append(cfg, WithGlazeToggleFlag(v.name))
		case hiddenGlazeFlagsOption:
			cfg = append(cfg, WithHiddenGlazeFlags(v.flagNames...))
		case defaultToGlazeOption:
			cfg = append(cfg, WithDefaultToGlaze())
		}
	}
	
	return cfg
}

// Legacy option types - moved to deprecated.go

// DualModeOption provides customization options for BuildCobraCommandDualMode
// Deprecated: Use CobraOption instead
type DualModeOption interface {
	apply(*dualModeConfig)
}

type dualModeConfig struct {
	glazeToggleFlag  string
	hiddenGlazeFlags []string
	defaultToGlaze   bool
}

type glazeToggleFlagOption struct {
	name string
}

func (g glazeToggleFlagOption) apply(cfg *dualModeConfig) {
	cfg.glazeToggleFlag = g.name
}

// WithGlazeToggleFlagDualMode lets you rename or shorten the toggle flag
// Deprecated: Use WithGlazeToggleFlag as a CobraOption instead
func WithGlazeToggleFlagDualMode(name string) DualModeOption {
	return glazeToggleFlagOption{name: name}
}

type hiddenGlazeFlagsOption struct {
	flagNames []string
}

func (h hiddenGlazeFlagsOption) apply(cfg *dualModeConfig) {
	cfg.hiddenGlazeFlags = h.flagNames
}

// WithHiddenGlazeFlagsDualMode marks specific Glaze‑layer flags to stay hidden even
// after the toggle; use when you expose only JSON rendering, for instance.
// Deprecated: Use WithHiddenGlazeFlags as a CobraOption instead
func WithHiddenGlazeFlagsDualMode(flagNames ...string) DualModeOption {
	return hiddenGlazeFlagsOption{flagNames: flagNames}
}

type defaultToGlazeOption struct{}

func (d defaultToGlazeOption) apply(cfg *dualModeConfig) {
	cfg.defaultToGlaze = true
}

// WithDefaultToGlazeDualMode makes Glaze mode the default unless the user disables it
// with --no-glaze-output (builder auto‑creates the negated flag).
// Deprecated: Use WithDefaultToGlaze as a CobraOption instead
func WithDefaultToGlazeDualMode() DualModeOption {
	return defaultToGlazeOption{}
}
