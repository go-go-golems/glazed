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

By default, glazed will try to detect the style of the terminal it's being run
in and render in a compatible style (dark or light). There are some instances
where this will break; notably, tmux sometimes causes styles to be detected
incorrectly.

If you want to override the styles, you can do a few things.
- Setting the environment variable `COLORFGBG` will allow detection of light vs. dark mode
- Setting `GLAMOUR_STYLE` to "light", "dark", or "notty" will override the styles to their respective setting, with "notty" useful in pipelines.
- If `GLAMOUR_STYLE` is not set, then glazed will look at whether the file descriptor for stderr is a TTY or not. If it is a TTY, colors etc. will be omitted.

There are some [docs on styles](https://github.com/charmbracelet/glamour#styles) provided in the [Glamour repo](https://github.com/charmbracelet/glamour).
