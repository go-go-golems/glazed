---
Title: Simplified Config Parser API for Glazed
Ticket: CONFIG-PARSER-001
Status: active
Topics:
    - glazed
    - config
    - api-design
    - parsing
DocType: index
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: glazed/pkg/cli/cobra-parser.go
      Note: CobraParser bridging ParameterLayers to Cobra commands
    - Path: glazed/pkg/cli/cobra.go
      Note: BuildCobraCommand unified command builder
    - Path: glazed/pkg/cmds/layers/layer-impl.go
      Note: Standard ParameterLayerImpl implementation
    - Path: glazed/pkg/cmds/layers/layer.go
      Note: ParameterLayer interface and ParameterLayers collection
    - Path: glazed/pkg/cmds/layers/parsed-layer.go
      Note: ParsedLayer and ParsedLayers runtime value containers
    - Path: glazed/pkg/cmds/middlewares/cobra.go
      Note: Cobra command parsing middleware
    - Path: glazed/pkg/cmds/middlewares/middlewares.go
      Note: Middleware execution framework
    - Path: glazed/pkg/cmds/middlewares/update.go
      Note: Environment variable and map update middlewares
    - Path: glazed/pkg/cmds/parameters/initialize-struct.go
      Note: Struct tag parsing and InitializeStruct implementation
    - Path: glazed/pkg/cmds/parameters/parameters.go
      Note: Core parameter definition types and functions
    - Path: moments/docs/backend/appconfig-quickstart.md
      Note: Prior art quickstart for schema-first typed config boundary
    - Path: pinocchio/cmd/pinocchio/main.go
      Note: Example of current complex setup requiring manual layer and middleware configuration
    - Path: glazed/ttmp/2025/12/15/CONFIG-PARSER-001--simplified-config-parser-api-for-glazed/analysis/01-glazed-parameter-parsing-architecture-analysis.md
      Note: |-
        Comprehensive analysis of Glazed parameter parsing architecture - documents all components
        Comprehensive analysis documenting all components
    - Path: glazed/ttmp/2025/12/15/CONFIG-PARSER-001--simplified-config-parser-api-for-glazed/design-doc/01-design-struct-first-configparser-api-on-top-of-glazed.md
      Note: Proposed struct-first ConfigParser API with implementation designs
    - Path: glazed/ttmp/2025/12/15/CONFIG-PARSER-001--simplified-config-parser-api-for-glazed/reference/01-diary.md
      Note: Diary documenting exploration process with all search queries
    - Path: glazed/ttmp/2025/12/15/CONFIG-PARSER-001--simplified-config-parser-api-for-glazed/reference/02-research-brainstorm-new-config-api-diary.md
      Note: Lab book diary (searches/results + synthesis + Moments appconfig prior art)
ExternalSources: []
Summary: 'Workspace for CONFIG-PARSER-001: research + design for a struct-first ConfigParser API on top of Glazed (flags/env/config files → typed structs), including Moments `appconfig` prior art.'
LastUpdated: 2025-12-15T08:46:54.001880256-05:00
---





# Simplified Config Parser API for Glazed

Document workspace for CONFIG-PARSER-001.

## Key documents

- `analysis/01-glazed-parameter-parsing-architecture-analysis.md`: Glazed architecture “health inspection”
- `reference/02-research-brainstorm-new-config-api-diary.md`: lab book (searches + findings + synthesis)
- `reference/03-colleague-quiz-cleanup-merge.md`: colleague interview/survey quiz to understand what can be removed/merged and why it was built this way
- `reference/04-diary-appconfig-module-design-redo.md`: diary for the `appconfig.Parser` redesign (searches + decisions + finalization)
- `design-doc/02-design-appconfig-module-register-layers-and-parse.md`: **current** proposed `appconfig.Parser` API (register layers + tagged structs, Parse() returns `T`)
- `design-doc/01-design-struct-first-configparser-api-on-top-of-glazed.md`: prior design (kept for history; superseded by 02)
