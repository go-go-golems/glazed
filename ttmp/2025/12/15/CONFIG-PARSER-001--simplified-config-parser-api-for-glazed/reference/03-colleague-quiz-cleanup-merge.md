---
Title: Colleague quiz - cleanup/merge/removal rationale for config parsing stack
Ticket: CONFIG-PARSER-001
Status: active
Topics:
    - glazed
    - config
    - api-design
    - parsing
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: glazed/pkg/cmds/middlewares/patternmapper/pattern_mapper.go
      Note: Prefix semantics + config mapping behavior; key to “can we merge/remove?” decisions
    - Path: glazed/pkg/cmds/middlewares/load-parameters-from-json.go
      Note: Config file loader + ConfigMapper hook (often the seam for simplification)
    - Path: glazed/pkg/cmds/runner/run.go
      Note: Programmatic parsing entrypoint and middleware ordering/precedence
    - Path: glazed/pkg/cli/cobra-parser.go
      Note: CLI wiring + default middleware chain; reveals what app code is forced to do today
ExternalSources: []
Summary: "Interview/survey quiz for colleagues to uncover why the current Glazed/Pinocchio config parsing pieces exist, which constraints drove them, and which parts can be merged/removed without breaking users."
LastUpdated: 2025-12-16T00:00:00Z
---

# Colleague quiz: what can we remove/merge/clean up (and why was it built this way)?

## How to use this

- **Format**: 30–60 min interview (best) or async survey (OK).
- **Instruction**: Please answer with **examples** (links to PRs/issues/bugs), **who uses it**, and **what would break** if we removed it.
- **Confidence**: For each answer, add:
  - **Confidence**: high / medium / low
  - **Evidence**: “saw it in prod”, “remember discussion”, “guess”

## 0) Warm-up: context, role, and time-travel

1. **What problem were you solving** when you first touched config parsing here?
   - What was the “pain” that motivated the current approach?
2. **What timeframe / era** are you thinking of? (approx date, release, migration)
3. **Which repos / apps** did this need to serve? (Glazed, Pinocchio, Geppetto, others)
4. **Who are the users**?
   - CLI end-users, library users, internal-only tooling, OSS consumers?
5. **If we rewound time**, what alternative approach was considered and rejected? Why?

## 1) Inventory: what pieces exist, and which are “core” vs “incidental”?

For each item, answer: **core? optional? legacy?**

1. **ParameterDefinitions / types / validation**
2. **ParameterLayers (slugs, prefixes, child layers)**
3. **ParsedLayers as runtime container**
4. **Middlewares chain as the value-population mechanism**
5. **CobraParser and Cobra integration helpers**
6. **Config file loading + mapping (ConfigMapper / patternmapper)**
7. **Profiles / overlays**
8. **Any Viper-era leftovers (if applicable)**

Follow-ups:

9. **Which of these are stable public APIs** (OSS) vs internal implementation details?
10. **Which pieces are hardest to change** because of external callers?

## 2) Precedence and source model: why this order, and which edge cases forced it?

1. **What is the intended precedence model** today?
   - Defaults < config files (low→high) < env < args < flags (or something else?)
2. **What real-world bug** would happen if we changed that order?
3. **Should config files ever override env/flags?**
   - If yes, where did that requirement come from?
4. **Should args exist as a “source” for typed config**, or is it only for command arguments?
5. **Do we need “meta” about where values came from** (for debugging/UX)?
   - If yes, which tools rely on it (print-yaml, print-parsed-parameters, etc.)?

## 3) Layers: are they essential, or accidental complexity?

1. **Why do layers exist** instead of a single flat registry?
2. **Who benefits** from layers? (help grouping, naming, namespacing, reuse, composition)
3. **How often do real apps use multiple layers** beyond “default + settings”?
4. **If we collapsed to one layer** (all params flattened), what would break?
5. **Do layer slugs leak into app code** today? Is that acceptable?
6. **Do child layers matter** in practice, or could we remove that feature?

## 4) Naming and prefixes: what are the intended semantics (and where do they differ)?

We’re trying to understand what “prefix” is supposed to mean across the stack.

1. **For env vars and flags**, should `layer.Prefix` be:
   - A purely external namespace (“redis-host” flag) while internal param key remains “host”?
   - Or part of the canonical parameter name (“redis-host” is the key everywhere)?
2. **For config files**, should prefixes appear in YAML keys?
   - If yes, show an example config that relies on that.
3. **Have you seen collisions** that prefixes were meant to solve?
4. **What naming conversions are required**?
   - Go field `FooBar` → flag `foo-bar`? env `FOO_BAR`? config `foo.bar`?
5. **Where do you expect renames to be handled** (compat aliases, deprecations)?

## 5) Config mapping: nested YAML vs layer-shaped YAML (why both, and which should win)?

1. **What is the preferred user-facing config shape**?
   - Layer-shaped:
     - `layer-slug: { param: value }`
   - Nested YAML:
     - `tools: { redis: { host: ... } }`
2. **Why was `ConfigMapper` added**? What specific cases demanded it?
3. **Does `patternmapper` solve a real pain**, or is it an over-engineered “nice to have”?
4. **Prefix semantics in mapping**:
   - Should mapping rules target leaf keys (“host”) or prefixed keys (“redis-host”)?
5. **If we introduced a new mapper (path/schema mapper)** with “prefix external-only” semantics:
   - Would that conflict with existing users?
   - Could patternmapper be deprecated, or must it stay?
6. **Do we need captures/wildcards** in mappings in real use?
   - Provide examples where `{env}` or `*` was essential.
7. **Unknown keys**:
   - Should unknown config keys be ignored, warned, or error?
   - Who depends on strictness?

## 6) Struct-first API and tags: what do we need, and what can we drop?

1. **Do you want structs as the primary config boundary** (typed settings in, no Glazed types exposed)?
2. **Should struct tags be required**?
   - Or should we derive names by convention and only use tags for overrides?
3. **Which metadata must be expressible** in structs?
   - help text, required, choices, default values, secret/credential marking, etc.
4. **Is “schema-first” (derive schema from struct)** acceptable, or do we need explicit schemas?
5. **Migration strategy**:
   - How do we support existing `glazed` tags during migration?
   - Do we need both directions (struct→defaults and parsed→struct)?
6. **What’s the minimum viable set of Go types** we must support on day 1?
   - (string/bool/int/float, slices, durations, time, maps, custom types?)

## 7) Cobra integration: what’s the real coupling, and what’s accidental?

1. **Is Cobra mandatory** for the intended consumers?
2. **What responsibilities belong in “CLI wiring” vs “config parsing core”?**
3. **Which parts of CobraParserConfig are essential**?
   - AppName/env prefix/config path discovery/config files func/etc.
4. **What should a simplified API generate**:
   - Cobra command directly? layers+parser config? a runner-like entrypoint?
5. **What would it take to support non-Cobra contexts** cleanly (library-only parsing)?

## 8) Profiles / overlays: why do they exist, and can we simplify safely?

1. **What user story requires profiles** (dev/prod/test, customer envs, feature toggles)?
2. **How are profiles selected** today (flag/env/config)?
3. **Have you hit the “profile selection timing” bug** (selection read too early)?
   - What broke, and how was it diagnosed?
4. **Do we need multi-phase parsing** to do profiles correctly?
   - If yes, what is the minimum “phase model” that avoids bugs?
5. **Could we replace profiles with multi-file config overlays**?
   - If not, what’s missing?
6. **If profiles are optional**, can we make them an add-on source rather than core complexity?

## 9) Legacy + duplication: what can we delete, and what’s still needed for compatibility?

1. **Are there legacy implementations still in the tree** (e.g., Viper-based parsing or editing)?
   - Which ones are actively used vs dead code?
2. **What duplication exists** across repos (e.g. Moments `appconfig` vs new proposed API)?
3. **What cleanup is “safe”** (internal-only) vs “risky” (OSS API / external consumers)?
4. **If we had to delete one subsystem** to simplify everything, what would you delete first and why?

## 10) Concrete candidates: react to possible removals/merges

For each candidate, answer:
**Keep / Remove / Merge / Deprecate** + “break risk” + “who complains”.

1. **Collapse to a single internal layer** for struct-first ConfigParser V1
2. **Stop exposing `ParsedLayers` outside parsing boundary** (typed struct only)
3. **Introduce a new “schema/path mapper” and de-emphasize `patternmapper`**
4. **Align prefix semantics** across the stack (prefix external-only)
5. **Remove/contain Viper dependencies** (if still present anywhere in parsing path)
6. **Unify programmatic parsing (runner) and CLI parsing (CobraParser)** behind one façade
7. **Profiles as an optional add-on source** (not always-on complexity)

## 11) “Show me the bodies”: archaeology prompts (high signal)

1. Link to **the PR/commit** that introduced the piece you think is hardest to remove.
2. Link to **an issue/bug** that explains why a “weird” decision was necessary.
3. Name **the most important edge case** that a clean rewrite would miss.
4. Name **one thing you’re confident we can delete** with minimal fallout.

## 12) Summary and decision notes

1. **If we optimize for approachability**, what do we sacrifice?
2. **If we optimize for backwards compatibility**, what complexity must remain?
3. **Your recommended cleanup plan**, in 3 bullets:
   - (a) quickest win
   - (b) most impactful refactor
   - (c) riskiest change worth doing


