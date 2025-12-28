---
Title: Report
Ticket: 004-PROFILE-DEFAULT-NAME
Status: active
Topics:
  - profiles
  - config
DocType: reference
Intent: long-term
Owners:
  - manuel
RelatedFiles: []
ExternalSources: []
Summary: Summary of the fix for honoring custom default profile names in missing-file semantics.
LastUpdated: 2025-12-28T18:23:07.591171705Z
---

# Report

## Goal
Provide a quick summary of the bug, the implemented fix, and the documentation updates for default profile name handling.

## Context
`WithProfileDefaultName` allows callers to change the default profile name away from "default". The profile middleware previously treated only the literal "default" as optional when profiles.yaml was missing or lacked the requested profile, which caused errors when a custom default was configured.

## Quick Reference
- Updated `GatherFlagsFromProfiles` to accept `defaultProfileName` and use it in missing-file and missing-profile checks.
- Updated `WithProfile` to pass the configured default profile name into the middleware.
- Updated profile-related docs to reference the new argument and semantics.

## Usage Examples
```go
middlewares.GatherFlagsFromProfiles(
    defaultProfileFile,
    profileFile,
    profileName,
    defaultProfileName,
)
```

## Related
- Design doc: `design-doc/01-default-profile-missing-file-analysis.md`
- Diary: `reference/01-diary.md`
