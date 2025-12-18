---
Title: Diary
Ticket: 002-DOCS-PROFILES
Status: active
Topics:
    - docs
    - glazed
    - profiles
    - geppetto
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-18T15:09:37.204895766-05:00
---

# Diary

## Goal

Track the documentation audit + edits required to make profile loading behavior clear and accurate across:

- Glazed docs (`glazed/pkg/doc/...`)
- Geppetto docs (`geppetto/pkg/doc/...`)

## Context

We recently fixed a real-world profile-selection circularity in Geppetto by doing an Option A “bootstrap parse” of `profile-settings` from defaults/config/env/flags before loading `profiles.yaml`. Existing docs are partly Pinocchio-specific, partially outdated (paths and config structure), and don’t clearly explain the precedence / configuration shape.

## Quick Reference

## Step 1: Ticket created + docs audit kickoff

### What I did

- Created ticket `002-DOCS-PROFILES`
- Identified the main existing docs to update:
  - `glazed/pkg/doc/topics/12-profiles-use-code.md`
  - `glazed/pkg/doc/topics/21-cmds-middlewares.md`
  - `geppetto/pkg/doc/topics/01-profiles.md`

### What I learned

- Glazed docs currently have “implementing profiles” material, but it is **Pinocchio/Geppetto-flavored** and doesn’t describe:
  - `profile-settings:` config key shape
  - XDG config search order (for `config.yaml`)
  - how to avoid circularity when profile selection comes from env/config

### Next

- Add a new, authoritative Glazed topic page: `glazed/pkg/doc/topics/15-profiles.md`
- Adjust existing pages to point to it and fix examples.

## Usage Examples

N/A (this is a work diary).

## Related

See ticket docs under `analysis/` for the planned set of edits.
