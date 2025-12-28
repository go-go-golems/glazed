---
Title: Default profile missing-file analysis
Ticket: 004-PROFILE-DEFAULT-NAME
Status: active
Topics:
  - profiles
  - config
DocType: design-doc
Intent: long-term
Owners:
  - manuel
RelatedFiles: []
ExternalSources: []
Summary: Honor custom default profile names in missing-file handling for profiles.yaml.
LastUpdated: 2025-12-28T18:23:07.430603115Z
---

# Default profile missing-file analysis

## Executive Summary
The profile middleware currently treats only the hard-coded profile name "default" as optional when the default profiles.yaml file is missing or lacks a profile entry. This breaks `WithProfileDefaultName`, which allows callers to redefine the default profile name. The fix is to pass the configured default profile name into `GatherFlagsFromProfiles` and use that value when deciding whether missing files or missing profiles should be tolerated.

## Problem Statement
`WithProfileDefaultName` lets callers change the default profile name away from "default". However, `GatherFlagsFromProfiles` still compares the requested profile name to "default" when deciding whether to ignore missing profiles.yaml files or missing profile entries. When a custom default name (e.g., "dev") is configured and profiles.yaml is missing, the middleware erroneously fails, making the option unusable in common "optional default profile" scenarios.

## Proposed Solution
Extend `GatherFlagsFromProfiles` to accept the configured default profile name and compare against it instead of a hard-coded "default" string. Update `WithProfile` to pass `pcfg.defaultProfile` into the middleware. Update documentation that references the function signature and default profile semantics.

## Design Decisions
- **Pass default profile name explicitly**: Keeps `GatherFlagsFromProfiles` logic self-contained and avoids reliance on global state.
- **Preserve existing behavior for callers who don't customize defaults**: When the default profile name is "default", behavior remains unchanged.

## Alternatives Considered
- **Infer the default profile name inside `GatherFlagsFromProfiles`**: Not possible without changing the function signature or adding global configuration.
- **Soft-fail all missing profile file cases**: Would break expectations for explicitly requested non-default profiles.

## Implementation Plan
1. Update `GatherFlagsFromProfiles` signature to accept a `defaultProfileName` argument and use it in missing-file and missing-profile checks.
2. Update `WithProfile` to pass the configured default profile name into the middleware.
3. Adjust documentation in `pkg/doc/topics/*` to reflect the new argument and semantics.
4. Add diary entries, update ticket changelog, and relate touched files.
