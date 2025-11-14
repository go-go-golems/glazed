---
Title: Markdown Style
Slug: markdown-style
Short: glazed help system output style can be configured.
Topics:
- help
Commands:
- help
Flags: []
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

## Overview

Glazed renders help pages through [Glamour](https://github.com/charmbracelet/glamour), which auto-detects whether the output target prefers a light or dark palette. Because the help system now writes to **stdout by default**, style detection looks at the stdout file descriptor unless you override the writer (see `help_cmd.SetHelpWriter`). If you route help somewhere else—`os.Stderr`, a file, or an in-memory buffer—the detection logic follows that writer.

When automatic detection misbehaves (tmux panes, nested SSH, etc.) you can pin the style explicitly:

- Set the `COLORFGBG` environment variable so Glamour can infer light vs. dark themes in terminals that hide this metadata.
- Set `GLAMOUR_STYLE` to `"light"`, `"dark"`, or `"notty"` to force a specific palette. `"notty"` is ideal for pipelines, CI logs, or anywhere ANSI colors should be suppressed.
- If you prefer the legacy “help on stderr” behavior—for example to keep stdout clean for machine-readable output—call `help_cmd.SetHelpWriter(os.Stderr)` during startup. Glamour will then use stderr for its TTY checks.

Refer to the [Glamour style documentation](https://github.com/charmbracelet/glamour#styles) for the complete list of built-in themes and customization options.
