---
Title: 'Debate Round 1: Option A naming + API design (schema/fields/values/sources)'
Ticket: GLAZED-LAYER-RENAMING
Status: active
Topics:
    - glazed
    - api-design
    - naming
    - migration
    - cli
    - cobra
DocType: debate
Intent: working-document
Owners:
    - manuel
RelatedFiles:
    - Path: glazed/pkg/cli/cobra-parser.go
      Note: CobraParserConfig and default middlewares chain (flags/args/env/config/defaults)
    - Path: glazed/pkg/cli/cobra.go
      Note: Current cobra glue and command definition expectations
    - Path: glazed/pkg/cmds/layers/layer.go
      Note: Current schema grouping types whose naming we’re replacing under Option A
    - Path: glazed/pkg/cmds/layers/parsed-layer.go
      Note: Current values container types (ParsedLayer/ParsedLayers) and InitializeStruct semantics
    - Path: glazed/pkg/cmds/middlewares/layers.go
      Note: ExecuteMiddlewares and source-chain semantics
    - Path: glazed/pkg/cmds/parameters/parameters.go
      Note: Current field spec type (ParameterDefinition) targeted by Option A
    - Path: glazed/pkg/doc/tutorials/05-build-first-command.md
      Note: Tutorial anchor for CLI registration discussion
    - Path: glazed/ttmp/2025/12/16/GLAZED-LAYER-RENAMING--rename-parameterlayer-parsedlayers-parameterdefinition-for-clarity/analysis/01-option-a-implementation-plan-schema-field-values-renaming-transitional-api.md
      Note: Option A implementation plan referenced during debate
    - Path: glazed/ttmp/2025/12/16/GLAZED-LAYER-RENAMING--rename-parameterlayer-parsedlayers-parameterdefinition-for-clarity/reference/01-debate-prep-candidates-and-questions-for-renaming-layers-parameters.md
      Note: Debate prep doc (agenda/candidates)
ExternalSources: []
Summary: Debate round exploring Option A names and API surfaces, plus command-definition and CLI registration implications
LastUpdated: 2025-12-17T08:42:20.752014633-05:00
---


## Pre-Debate Research

This debate is constrained to **no performance** and **no security** arguments.

### Research commands (reproducible)

The following repository queries were run to gather evidence:

- `rg "cli.BuildCobraCommand|CobraParserConfig|CobraCommandDefaultMiddlewares" glazed/pkg -n`
- `rg "ParseFromCobraCommand\\(|GatherArguments\\(|ExecuteMiddlewares\\(" glazed/pkg -n`
- `rg "type\\s+ParameterLayer\\b|type\\s+ParsedLayer\\b|type\\s+ParsedLayers\\b" glazed/pkg/cmds/layers -n`

### Findings (key facts, with file pointers)

#### 1) Cobra bridge wiring is already non-trivial

The CLI integration tutorial uses `cli.BuildCobraCommand` with a `cli.CobraParserConfig` and a default middleware function (`cli.CobraCommandDefaultMiddlewares`):

- Tutorial: `glazed/pkg/doc/tutorials/05-build-first-command.md` (CLI Application Integration section)

#### 2) “Middlewares” are functionally “sources/resolvers”

The default cobra parsing flow is explicitly a precedence chain that reads like “sources”:

- `glazed/pkg/cli/cobra-parser.go`
  - `CobraCommandDefaultMiddlewares` composes:
    - `ParseFromCobraCommand` (flags)
    - `GatherArguments` (positional args)
    - `SetFromDefaults` (defaults)
  - `NewCobraParserFromLayers` optionally composes:
    - flags, args
    - env (`UpdateFromEnv`) derived from `AppName`
    - config (`LoadParametersFromResolvedFilesForCobra`)
    - defaults

This is semantically a *source resolution chain*, not generic middleware.

#### 3) Command “definition” is centered on `cmds.CommandDescription` and layers

The cobra glue (`glazed/pkg/cli/cobra.go`) expects commands that expose `s.Description()` returning a `*cmds.CommandDescription` with a `Layers` collection.

That means any Option A package rename needs to decide:

- do we keep command descriptions speaking in terms of “layers” internally?
- do we provide a new `schema` vocabulary at the edge while keeping internals stable?

## Opening Statements (Round 1)

### The Architect

Option A is the right vocabulary, but it only “lands” if the import paths and nouns are coherent:

- `schema` should mean: section schemas, composed of fields, with metadata (name/slug/prefix).
- `fields` should mean: field definitions/specs (type/default/help/required).
- `values` should mean: resolved values with provenance; and decoding/hydration helpers.
- `sources` should mean: env/config/cobra/defaults, composed in an ordered chain.

Crucially, the code already *behaves* like sources (see `cli/cobra-parser.go` chains), so the rename is aligning words with reality.

I’m strongly in favor of introducing façade packages that re-export legacy implementations (type aliases and thin wrappers) rather than moving code immediately. That lets us:

- preserve downstream compatibility
- standardize docs and examples
- incrementally shift `cmds.CommandDescription`-level APIs later, if we choose.

### The Pragmatist

We’ve decided to rename packages, but we should keep the initial Option A package layer extremely thin and non-invasive.

The biggest risk is a vocabulary split where:

- docs say “schema/fields/values”
- but the code still reads `schema.Section` and `fields.Definition`
- and users have to learn both anyway.

So my proposal:

- start by building `schema/fields/values/sources` as façades with *aliases* + a few key helper functions
- don’t rename struct tags or serialized shapes
- don’t try to restructure `cmds.CommandDescription` in the first wave

We should also pick names that are “low drama”:

- keep “layer” internally (implementation), expose “schema section” in docs
- prefer verbs that match the existing behavior: `DecodeInto` is fine, but “Hydrate” might suggest magic; “Decode” implies mapping from a value store.

### The Migration Engineer

The core compatibility reality:

- package moves are breaking; façade packages are additive.
- type aliases (`type X = Y`) are additive.
- method renames can’t be done with aliases; use helper functions or wrapper types.

Therefore, the safe transitional API is:

1. Add new packages (`schema`, `fields`, `values`, `sources`) that re-export old types via aliases.
2. Add helper functions for the “verbs” we want (`DecodeSectionInto`, etc.).
3. Deprecate old names later, not immediately.

I’m also flagging a subtlety: `ParsedLayers.InitializeStruct(DefaultSlug, dst)` special-cases “default”. If we introduce a new `values.DecodeInto(...)`, we must decide whether that special-casing is preserved or hidden.

### The Doc Maintainer

From the tutorial evidence (`05-build-first-command.md`), the biggest learnability problem isn’t only nouns—it’s “how do I wire a command into Cobra”.

So Option A naming needs to “meet users where they are”:

- show an opinionated happy-path for registration (maybe a CLI builder / registry)
- make the glossary consistent across docs and examples
- and avoid introducing *yet another* quasi-framework unless it reduces cognitive load.

I’d prioritize:

- naming the user-facing story: “define schema sections + resolve sources + decode into settings”
- and then make sure the cobra tutorial can express that cleanly.

### The New Developer

As a newcomer:

- “ParameterLayer” reads like “a precedence layer/source”, not “a schema section”.
- “ParsedLayers” sounds like “parsed sources”, not “resolved values”.

So Option A is a big win, but only if it’s consistent:

- `schema.Section` / `fields.String` / `sources.Env` / `values.DecodeSectionInto`

Also: “middlewares” should not be the word users see. If the tutorial says “middlewares”, it needs to immediately translate to “sources” in the new world.

## Rebuttals (Round 2)

### Architect → Pragmatist

Agree on façades-first. The only point I’d push: even if internals keep “layers”, the public API should not. If we keep “layer” in `schema` package names/types, we’re not solving the confusion.

### Pragmatist → Architect

Agree, but we need to avoid pushing confusing words like “schema” into places where we don’t validate. We should be explicit in docs: Glazed schema is descriptive, not necessarily validating.

### Migration Engineer → All

We should bake the compatibility matrix into the plan:

- Phase 1: façades + helper verbs (non-breaking)
- Phase 2: docs/examples migrate
- Phase 3: optional deprecations

And we should not let “rename packages” mean “move code”. It should mean “new packages re-export old code” until we’re ready.

### Doc Maintainer → CLI discussion

If we decide to improve CLI registration, we need to keep the tutorial story short:

- a single “builder” entrypoint
- minimal config knobs in the default path

Otherwise we’ll end up with a well-named but still hard-to-use system.

## Moderator Summary

### What seems aligned

- Everyone supports **Option A vocabulary** as the baseline.
- Everyone prefers **façade packages** (additive) over immediate code/package moves.
- Everyone agrees method/verb renames should be done via **helper functions** (or wrappers), not by trying to rename methods on existing types.

### Key tensions (to resolve next)

- **What do we call the value container?** `values.Values` vs `values.ResolvedValues` vs `values.Store`.
- **What do we call the grouping concept publicly?** keep “layer” vs “section/group/namespace”.
- **Verb choice**: `DecodeInto` vs `BindInto` vs `HydrateInto` (and how much provenance is exposed).
- **CLI builder scope**: can we meaningfully simplify `BuildCobraCommand` + `CobraParserConfig` into an opinionated “registry/builder” without exploding surface area?

### Concrete next steps suggested by the round

- Prototype façade packages:
  - `schema`, `fields`, `values`, `sources`
  - mostly alias-based, plus a small set of helper verbs.
- Draft an updated “build first command” tutorial that uses Option A vocabulary and a simpler registration story (even if implemented as wrappers around current CLI).
