---
Title: Remove Logstash logging support from glazed
Ticket: GL-003
Status: active
Topics:
    - backend
    - logging
    - refactor
    - cleanup
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: clay/examples/logstash/logstash_example.go
      Note: Dedicated Logstash example to delete
    - Path: clay/examples/simple/logging_layer_example.go
      Note: General logging example with Logstash references
    - Path: glazed/pkg/cmds/logging/init-early.go
      Note: Pre-cobra logging arg parsing with Logstash flags
    - Path: glazed/pkg/cmds/logging/init.go
      Note: Initializes logger from settings including Logstash writer
    - Path: glazed/pkg/cmds/logging/logstash_writer.go
      Note: LogstashWriter implementation to delete
    - Path: glazed/pkg/cmds/logging/section.go
      Note: Defines LoggingSettings struct and logging section fields/flags
ExternalSources: []
Summary: ""
LastUpdated: 2026-05-02T12:46:20.696035278-04:00
WhatFor: ""
WhenToUse: ""
---







# Remove Logstash logging support from glazed

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- backend
- logging
- refactor
- cleanup

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
