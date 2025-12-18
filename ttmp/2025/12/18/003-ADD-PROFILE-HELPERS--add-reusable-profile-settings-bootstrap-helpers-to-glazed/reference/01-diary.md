---
Title: Diary
Ticket: 003-ADD-PROFILE-HELPERS
Status: active
Topics:
    - glazed
    - profiles
    - cobra
    - middleware
    - refactor
    - docs
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-18T15:17:08.649837621-05:00
---

# Diary

## Goal

Track the investigation + design for moving reusable “profile-settings bootstrap parse” logic out of Geppetto and into Glazed as helper methods.

## Context

`geppetto/pkg/layers/layers.go` currently contains a verbose but correct Option A bootstrap parse:

- resolve `command-settings` (for config file selection)
- resolve config file list (low → high)
- resolve `profile-settings` from defaults + config + env + flags
- instantiate `GatherFlagsFromProfiles` with resolved values

The goal is to make Geppetto (and other apps) simpler by providing **reusable helpers** in Glazed.

## Quick Reference

## Step 1: Ticket setup + initial audit

### What I did

- Created ticket `003-ADD-PROFILE-HELPERS`
- Related key code files (Geppetto middleware builder, Glazed Cobra parser, Glazed profile middleware)
- Audited existing Glazed extension points and building blocks:
  - `cli.CobraParserConfig` (AppName + ConfigFilesFunc + config discovery)
  - `middlewares.LoadParametersFromResolvedFilesForCobra`
  - `middlewares.GatherFlagsFromProfiles` / `GatherFlagsFromCustomProfiles`

### What I learned

- Glazed already has a clean mechanism for resolving config files via `CobraParserConfig.ConfigFilesFunc`.
- What’s missing is a **first-class helper** that:
  - resolves `profile-settings` using the same precedence sources as the main chain (defaults/config/env/flags)
  - returns a ready-to-use profile middleware (or resolved settings + metadata)

## Usage Examples

<!-- Show how to use this reference in practice -->

## Related

<!-- Link to related documents or resources -->
