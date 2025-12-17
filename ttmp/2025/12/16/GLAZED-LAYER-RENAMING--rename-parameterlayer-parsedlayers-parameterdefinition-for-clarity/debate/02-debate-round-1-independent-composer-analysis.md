---
Title: 'Debate Round 1: Independent Composer Analysis - Schema/Field/Values Renaming'
Ticket: GLAZED-LAYER-RENAMING
Status: active
Topics:
    - glazed
    - api-design
    - naming
    - migration
    - debate
DocType: reference
Intent: working-document
Owners:
    - manuel
RelatedFiles:
    - Path: glazed/pkg/cmds/layers/layer.go
      Note: Core ParameterLayer interface and ParameterLayers collection
    - Path: glazed/pkg/cmds/layers/parsed-layer.go
      Note: ParsedLayer and ParsedLayers implementation with InitializeStruct
    - Path: glazed/pkg/cmds/parameters/parameters.go
      Note: ParameterDefinition and ParameterDefinitions types
    - Path: glazed/pkg/cmds/middlewares/layers.go
      Note: Middleware functions operating on ParsedLayers
    - Path: glazed/pkg/cli/cobra.go
      Note: BuildCobraCommand registration patterns
    - Path: glazed/pkg/doc/tutorials/05-build-first-command.md
      Note: Tutorial showing current CLI registration patterns
Summary: 'Independent debate round analyzing Schema/Field/Values renaming with codebase evidence'
CreatedBy: composer-1
LastUpdated: 2025-12-17T08:45:00.000000000-05:00
---

# Debate Round 1: Independent Composer Analysis - Schema/Field/Values Renaming

## Pre-Debate Research

### Research Methodology

Each candidate conducted research using actual codebase queries and file analysis to ground their arguments in evidence.

### Research Findings

#### Usage Statistics

**ParameterLayer/ParsedLayer usage:**
- Found **2,525 matches** across **163 files** for `ParameterLayer|ParsedLayer|ParameterDefinition`
- Found **256 matches** across **61 files** for `InitializeStruct|DecodeInto|BindInto`

**Import patterns:**
- `glazed/pkg/cmds/layers` imported across multiple packages
- `glazed/pkg/cmds/parameters` used extensively for field definitions
- `glazed/pkg/cmds/middlewares` operates on `layers.ParsedLayers`

#### Key Code Evidence

**1. ParameterLayer Interface Structure** (`glazed/pkg/cmds/layers/layer.go:18-30`)
```go
type ParameterLayer interface {
    AddFlags(flag ...*parameters.ParameterDefinition)
    GetParameterDefinitions() *parameters.ParameterDefinitions
    InitializeParameterDefaultsFromStruct(s interface{}) error
    GetName() string
    GetSlug() string
    GetDescription() string
    GetPrefix() string
    Clone() ParameterLayer
}
```

**Key observation:** The interface mixes schema definition (`GetParameterDefinitions`) with metadata (`GetName`, `GetSlug`). The name "ParameterLayer" doesn't clearly indicate it's a schema container.

**2. ParsedLayer Structure** (`glazed/pkg/cmds/layers/parsed-layer.go:14-17`)
```go
type ParsedLayer struct {
    Layer      ParameterLayer
    Parameters *parameters.ParsedParameters
}
```

**Key observation:** `ParsedLayer` contains both a reference to the schema (`Layer`) and the parsed values (`Parameters`). The name suggests it's "parsed" but doesn't indicate it's a collection of resolved values.

**3. InitializeStruct Method** (`glazed/pkg/cmds/layers/parsed-layer.go:100-102`)
```go
func (ppl *ParsedLayer) InitializeStruct(s interface{}) error {
    return ppl.Parameters.InitializeStruct(s)
}
```

**Key observation:** The verb `InitializeStruct` is generic and doesn't convey that it's decoding/binding values into a struct. Alternative verbs like `DecodeInto` or `BindInto` might be clearer.

**4. ParameterDefinition** (`glazed/pkg/cmds/parameters/parameters.go:20-29`)
```go
type ParameterDefinition struct {
    Name       string        `yaml:"name"`
    ShortFlag  string        `yaml:"shortFlag,omitempty"`
    Type       ParameterType `yaml:"type"`
    Help       string        `yaml:"help,omitempty"`
    Default    *interface{}  `yaml:"default,omitempty"`
    Choices    []string      `yaml:"choices,omitempty"`
    Required   bool          `yaml:"required,omitempty"`
    IsArgument bool          `yaml:"-"`
}
```

**Key observation:** `ParameterDefinition` is clearly a field specification, not a runtime value. The name is accurate but verbose. "FieldDefinition" would be shorter and clearer.

**5. CLI Registration Pattern** (`glazed/pkg/cli/cobra.go:232-298`)
```go
func BuildCobraCommandFromCommandAndFunc(
    s cmds.Command,
    run CobraRunFunc,
    opts ...CobraOption,
) (*cobra.Command, error) {
    // ... creates parser from description.Layers
    cobraParser, err := NewCobraParserFromLayers(description.Layers, &cfg.ParserCfg)
    // ... adds to cobra command
    err = cobraParser.AddToCobraCommand(cmd)
}
```

**Key observation:** The API uses `description.Layers` (which is `*layers.ParameterLayers`). This is the public-facing API that tutorials reference.

**6. Middleware Pattern** (`glazed/pkg/cmds/middlewares/layers.go:11-26`)
```go
func ReplaceParsedLayer(layerSlug string, newLayer *layers.ParsedLayer) Middleware {
    return func(next HandlerFunc) HandlerFunc {
        return func(layers_ *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
            // ... operates on parsedLayers
        }
    }
}
```

**Key observation:** Middlewares operate on both `ParameterLayers` (schema) and `ParsedLayers` (values). The package name "middlewares" doesn't convey that these are value resolvers/sources.

## Opening Statements

### The Pragmatist

**Position:** Minimal changes, maximum compatibility, incremental adoption.

**Evidence-based argument:**

Looking at the codebase, I see **2,525 matches** across **163 files** using the current names. That's a lot of code to break. My research shows:

1. **Type aliases are sufficient** - Go's `type X = Y` aliases provide zero-cost compatibility. We can add:
   ```go
   type SchemaSection = ParameterLayer
   type SectionValues = ParsedLayer
   type FieldDefinition = ParameterDefinition
   ```
   This gives us new names without breaking anything.

2. **Package moves are expensive** - If we move from `pkg/cmds/layers` to `pkg/cmds/schema`, every import breaks. That's **163 files** to update. Instead, we should:
   - Keep packages where they are
   - Add type aliases for clarity
   - Update documentation and tutorials
   - Let developers migrate gradually

3. **Verb renames are optional** - `InitializeStruct` appears **256 times**. Renaming to `DecodeInto` requires method wrappers or breaking changes. The current name, while not perfect, is functional. We can improve it later if needed.

**Recommendation:**
- Add type aliases: `SchemaSection`, `SectionValues`, `FieldDefinition`
- Keep packages: `layers`, `parameters`
- Update docs/tutorials to use new names
- Deprecate old names with comments, but keep them forever
- **Zero breaking changes**

**Cost:** Low - just add aliases and update docs
**Benefit:** Immediate clarity improvement without migration pain

### The Architect

**Position:** Coherent domain model with clean package boundaries.

**Evidence-based argument:**

My analysis reveals a fundamental architectural issue: **we're mixing schema and values in the same concepts**. Look at the evidence:

1. **ParameterLayer mixes concerns** - It contains both schema (`GetParameterDefinitions()`) and metadata (`GetName()`, `GetSlug()`). This violates single responsibility. A "SchemaSection" should only define structure.

2. **ParsedLayer is ambiguous** - It contains both a schema reference (`Layer`) and values (`Parameters`). The name doesn't convey that it's "resolved values for a schema section."

3. **Package organization matters** - Current structure:
   - `pkg/cmds/layers` - contains both schema and values
   - `pkg/cmds/parameters` - contains field definitions
   - `pkg/cmds/middlewares` - contains value resolvers

   Better structure:
   - `pkg/cmds/schema` - schema sections and field definitions
   - `pkg/cmds/fields` - field definitions (or merge into schema)
   - `pkg/cmds/values` - resolved values
   - `pkg/cmds/sources` - value resolvers (rename middlewares)

4. **Import clarity** - With new packages:
   ```go
   import (
       "github.com/go-go-golems/glazed/pkg/cmds/schema"
       "github.com/go-go-golems/glazed/pkg/cmds/values"
   )
   
   schemaSections := schema.NewSections(...)
   resolvedValues := values.NewSectionValues(...)
   ```
   This is immediately clear: schema defines structure, values contain data.

**Recommendation:**
- Create new packages: `schema`, `fields`, `values`, `sources`
- Use façade pattern: new packages re-export from old implementations
- Gradually migrate internal code
- Keep old packages as compatibility layer
- **Clean separation of concerns**

**Cost:** Medium - new packages, façade layer, gradual migration
**Benefit:** Long-term architectural clarity and maintainability

### The Migration Engineer

**Position:** Explicit compatibility matrix, clear deprecation policy, safe upgrade path.

**Evidence-based argument:**

I've analyzed the compatibility requirements systematically:

**Breaking vs Non-Breaking Changes:**

1. **Type aliases (NON-BREAKING)** ✅
   ```go
   type SchemaSection = layers.ParameterLayer
   type SectionValues = layers.ParsedLayer
   ```
   - Old code continues to work
   - New code can use new names
   - Zero migration cost

2. **Package moves (BREAKING)** ❌
   - Moving `layers` → `schema` breaks all imports
   - **163 files** need updates
   - Downstream repos break
   - Requires major version bump

3. **Method renames (BREAKING)** ❌
   - `InitializeStruct` → `DecodeInto` breaks callers
   - **256 usages** need updates
   - Requires wrapper functions or breaking change

**Compatibility Strategy:**

**Phase 1: Additive (v1.x)**
- Add type aliases in existing packages
- Add helper functions for new verbs
- Update documentation
- **Zero breaking changes**

**Phase 2: Transitional (v2.0-alpha)**
- Create new packages with façade pattern
- Old packages re-export from new
- Both import paths work
- **Backward compatible**

**Phase 3: Deprecation (v2.x)**
- Mark old names as deprecated
- Provide migration guide
- Keep old names for 2+ major versions
- **Gradual migration**

**Recommendation:**
- **Phase 1 now:** Type aliases + docs
- **Phase 2 later:** New packages with façades
- **Phase 3 eventually:** Deprecate old names
- **Never remove:** Keep old names forever for compatibility

**Cost:** Low initially, medium over time
**Benefit:** Safe, predictable migration path

### The Doc Maintainer

**Position:** Tutorial clarity, searchability, consistent glossary.

**Evidence-based argument:**

I've reviewed the documentation and tutorials:

**Current Documentation Issues:**

1. **Tutorial confusion** (`glazed/pkg/doc/tutorials/05-build-first-command.md`):
   - Uses `layers.ParameterLayers` - what is a "layer"?
   - Uses `parsedLayers.InitializeStruct()` - what does "initialize" mean?
   - New developers struggle with terminology

2. **Glossary inconsistency:**
   - Docs use "layer" but don't explain it clearly
   - "Parameter" vs "Field" used interchangeably
   - No clear mental model

3. **Searchability problems:**
   - Searching for "schema" finds nothing
   - Searching for "field definition" finds "ParameterDefinition"
   - Developers can't discover concepts

**Proposed Documentation Structure:**

**Glossary:**
- **Schema Section:** A named group of field definitions (e.g., "output", "config")
- **Field Definition:** Specification of a single parameter (name, type, default, help)
- **Section Values:** Resolved values for a schema section
- **Value Sources:** Components that resolve values (CLI flags, env vars, config files)

**Tutorial Updates:**
```go
// Old (confusing)
layers := layers.NewParameterLayers(...)
parsedLayers := layers.NewParsedLayers(...)

// New (clear)
schemaSections := schema.NewSections(...)
resolvedValues := values.NewSectionValues(...)
```

**Recommendation:**
- **Update glossary first** - establish vocabulary
- **Rewrite tutorials** - use new names in examples
- **Add concept docs** - explain schema vs values distinction
- **Update all doc references** - consistency across docs
- **Keep old names in migration guide** - help developers transition

**Cost:** Medium - documentation rewrite
**Benefit:** Dramatically improved learnability and discoverability

### The New Developer

**Position:** Intuitive names, discoverable concepts, minimal learning curve.

**Evidence-based argument:**

As someone new to the codebase, I found the current names confusing:

**First Impressions:**

1. **"ParameterLayer"** - What is a "parameter layer"? Is it a layer of parameters? A parameter that's layered? Unclear.

2. **"ParsedLayers"** - Parsed what? Layers of what? The name doesn't tell me it contains resolved values.

3. **"ParameterDefinition"** - This one is clearer, but "FieldDefinition" would be more intuitive.

4. **"InitializeStruct"** - Initialize from what? Into what? "DecodeInto" or "BindInto" would be clearer.

**Discoverability Test:**

If I search for "how do I define a schema?" - I find nothing.
If I search for "how do I get parsed values?" - I find "ParsedLayers" but it's not clear.

With new names:
- "How do I define a schema?" → `schema.NewSection()`
- "How do I get values?" → `values.SectionValues`
- "How do I decode into a struct?" → `values.DecodeInto()`

**Recommendation:**
- **Schema/Field/Values** - immediately intuitive
- **New packages** - clean import paths
- **Clear verbs** - `DecodeInto` instead of `InitializeStruct`
- **Consistent naming** - everything follows the same pattern

**Cost:** Learning new names (but they're easier to learn!)
**Benefit:** Faster onboarding, less confusion, better developer experience

### pkg/cmds/layers ("The Layer Librarian")

**Position:** I encode both schema and values efficiently; splitting might help, but don't break my users.

**Evidence-based argument:**

I've been serving the codebase for a long time. Let me show you what I actually do:

**My Current Structure:**
- `ParameterLayer` interface: defines schema + metadata
- `ParsedLayer` struct: contains schema reference + values
- `ParameterLayers`: ordered collection of schemas
- `ParsedLayers`: ordered collection of resolved values

**Why I'm Structured This Way:**

1. **Schema and values are coupled** - A `ParsedLayer` needs its `Layer` reference to validate and understand values. Separating them completely would require passing both around.

2. **Order matters** - I use `OrderedMap` because layer order affects precedence. This is important for merging and resolution.

3. **I'm efficient** - My current structure minimizes allocations and lookups. Splitting might add indirection.

**What I'm Willing to Accept:**

- **Type aliases:** Fine, I can live with `SchemaSection = ParameterLayer`
- **New packages:** OK, but keep me as the implementation
- **Method renames:** `DecodeInto` is fine, but keep `InitializeStruct` as alias
- **Don't break my API:** My methods are used everywhere

**Recommendation:**
- Keep me as the implementation layer
- Add type aliases for clarity
- Create new packages that wrap/re-export me
- Gradually migrate, but keep me forever

**Cost:** Wrapper layer overhead
**Benefit:** Better names without breaking my efficient implementation

### pkg/cmds/parameters ("The Field Spec")

**Position:** I'm a field specification, not a runtime value; my name should reflect that.

**Evidence-based argument:**

I define field specifications. Look at my structure:

```go
type ParameterDefinition struct {
    Name       string
    Type       ParameterType
    Help       string
    Default    *interface{}
    Required   bool
    // ...
}
```

I'm **not** a runtime value. I'm a **specification** of what a field should be.

**Current Problems:**

1. **"Parameter" is vague** - Could mean the definition or the value
2. **"Definition" is verbose** - "Field" would be clearer
3. **I live in "parameters" package** - Should I be in "fields" or "schema"?

**What I Want:**

- **"FieldDefinition"** - Clear, concise, accurate
- **Package location:** Either `fields` package or `schema` package (I define schema fields)
- **Keep my structure:** I'm well-designed, just need better naming

**Recommendation:**
- Rename to `FieldDefinition` (with alias for compatibility)
- Move to `schema` package (fields are part of schema)
- Or create `fields` package if we want separation
- Keep my efficient implementation

**Cost:** Type alias + package move (with compatibility)
**Benefit:** Clear, accurate naming

### pkg/cmds/middlewares ("The Source Chain")

**Position:** I'm fundamentally about value resolution from sources, not generic middleware.

**Evidence-based argument:**

Look at what I actually do:

```go
func SetFromDefaults(...)
func UpdateFromEnv(...)
func LoadParametersFromFiles(...)
func ParseFromCobraCommand(...)
func ExecuteMiddlewares(...)
```

I'm **resolving values from sources**:
- Defaults
- Environment variables
- Config files
- CLI flags/args

I'm not generic "middleware" - I'm a **value resolution chain**.

**Current Problems:**

1. **"middlewares" is misleading** - Suggests generic HTTP-style middleware
2. **My functions operate on ParsedLayers** - I'm filling values, not transforming requests
3. **Package name doesn't match purpose** - Should be "sources" or "resolvers"

**What I Want:**

- **Package rename:** `middlewares` → `sources` or `resolvers`
- **Function clarity:** Keep my functions, but update docs to explain I'm resolving from sources
- **API clarity:** Make it clear I'm about value resolution, not generic middleware

**Recommendation:**
- Rename package to `sources` (with compatibility alias)
- Update function docs to emphasize "resolving from source"
- Keep my efficient implementation
- Gradually migrate references

**Cost:** Package rename + compatibility layer
**Benefit:** Accurate naming that matches purpose

### pkg/cli ("The Cobra Bridge")

**Position:** Names must map cleanly to CLI mental model while supporting config/env.

**Evidence-based argument:**

I bridge Glazed commands to Cobra. My API is:

```go
func BuildCobraCommand(command cmds.Command, opts ...CobraOption) (*cobra.Command, error)
```

I work with `description.Layers` (which is `*layers.ParameterLayers`).

**CLI Mental Model:**

Users think in terms of:
- **Flags/Arguments** - what they can pass
- **Sections/Groups** - how flags are organized
- **Values** - what was actually passed

**Current API Issues:**

1. **"Layers" doesn't map to CLI concepts** - CLI users don't think in "layers"
2. **"ParameterLayers" is implementation detail** - Should be "SchemaSections" at API level
3. **My tutorials use confusing names** - Makes onboarding harder

**What I Need:**

- **Clear API names** - `SchemaSections` instead of `ParameterLayers`
- **CLI-friendly terminology** - "sections" or "groups" not "layers"
- **Backward compatibility** - Can't break existing code
- **Documentation clarity** - Tutorials should use clear names

**Recommendation:**
- Use new names in my public API (with compatibility)
- Update tutorials to use clear names
- Keep old names as aliases
- Make CLI registration story clearer

**Cost:** API updates + compatibility layer
**Benefit:** Clearer CLI integration story

### pkg/appconfig ("The AppConfig Boundary")

**Position:** Clear schema/values naming makes my API easier to explain; I'll benefit from new nouns.

**Evidence-based argument:**

I provide a high-level config parser API:

```go
type Parser struct {
    // registers schema sections
    // parses from multiple sources
    // returns resolved values
}
```

**Current Challenges:**

1. **I use "layers" terminology** - Makes my API harder to explain
2. **Schema vs values distinction unclear** - Users confuse what I register vs what I return
3. **My docs are confusing** - Because underlying concepts are confusing

**What I'll Gain:**

With new names:
- **Register schema sections** - Clear: I'm registering structure
- **Parse to resolved values** - Clear: I'm returning data
- **Value sources** - Clear: I resolve from multiple sources

**Recommendation:**
- **Strongly support** Schema/Field/Values naming
- **Update my API** to use new names
- **Improve my docs** with clearer terminology
- **Benefit from clarity** - My API becomes easier to understand

**Cost:** API updates + docs
**Benefit:** Much clearer high-level API

## Rebuttals

### The Pragmatist → The Architect

**Architect, you're proposing a lot of churn for theoretical purity.** Yes, mixing schema and values isn't perfect, but it works. Your façade packages add indirection and complexity. My type aliases give us 80% of the benefit with 20% of the cost.

**Counter-evidence:** Look at `ParsedLayer` - it needs the `Layer` reference for validation. Separating them completely would require passing both around everywhere. The coupling is intentional and useful.

### The Architect → The Pragmatist

**Pragmatist, type aliases don't solve the fundamental problem.** Yes, we can alias `SchemaSection = ParameterLayer`, but developers still import `layers` and see `ParameterLayer` in their IDE. The package name matters for discoverability.

**Counter-evidence:** New developers searching for "schema" won't find `layers.ParameterLayer`. They need to know to look in `layers` package. New packages solve this.

### The Migration Engineer → Both

**Both of you are missing the phased approach.** We can have:
- **Phase 1:** Type aliases (Pragmatist wins)
- **Phase 2:** New packages with façades (Architect wins)
- **Phase 3:** Gradual migration (Everyone wins)

This isn't either/or - it's a sequence.

### The Doc Maintainer → All

**All of you are focusing on code, but documentation is the real problem.** Even with perfect code names, if tutorials use confusing terminology, developers will struggle. We need to update docs regardless of code changes.

**Evidence:** Current tutorial uses `layers.ParameterLayers` - this is confusing even if we add type aliases. We need to update tutorials to use clear names.

### The New Developer → All

**As a new developer, I don't care about your migration concerns.** I just want clear, discoverable names. If I search for "schema" and find nothing, that's a problem. New packages solve this immediately.

**Counter to Pragmatist:** Type aliases don't help discoverability. I still need to know to import `layers` and use `SchemaSection` alias.

**Counter to Architect:** I don't need perfect separation - I just need clear names. If `ParsedLayer` contains both schema reference and values, that's fine, but call it `SectionValues` so I understand what it is.

### pkg/cmds/layers → The Architect

**Architect, you want to split me, but my coupling is intentional.** `ParsedLayer` needs its `Layer` reference. Separating completely would make the API more complex, not simpler.

**Compromise:** Keep my structure, but use clearer names. `SchemaSection` for schema, `SectionValues` for values. The coupling is fine - just name it clearly.

### pkg/cmds/parameters → The Architect

**Architect, where should I live?** You propose `schema` package, but I'm field definitions. Should I be:
- `schema.FieldDefinition` (fields are part of schema)
- `fields.FieldDefinition` (fields are separate concept)

**Question:** What's the relationship between schema sections and field definitions? Are fields part of schema, or separate?

### pkg/cmds/middlewares → All

**All of you are ignoring my naming problem.** I'm called "middlewares" but I'm value resolvers. This is confusing regardless of other changes.

**Recommendation:** Rename me to `sources` or `resolvers` as part of this effort. Don't leave me with a misleading name.

## Moderator Summary

### Key Arguments

**1. Type Aliases vs New Packages**
- **Pragmatist:** Type aliases provide clarity without breaking changes
- **Architect:** New packages provide discoverability and clean imports
- **Migration Engineer:** Both - phased approach with aliases first, packages later

**2. Verb Renaming**
- **Pragmatist:** `InitializeStruct` works, renaming is optional
- **Architect:** `DecodeInto` is clearer and more accurate
- **New Developer:** `DecodeInto` is more discoverable and intuitive

**3. Package Organization**
- **Architect:** Separate `schema`, `fields`, `values`, `sources` packages
- **Layer Librarian:** Keep current structure, just rename types
- **Migration Engineer:** Phased migration with compatibility layers

**4. Documentation Priority**
- **Doc Maintainer:** Documentation updates are critical regardless of code changes
- **All:** Agree that tutorials need updates

### Tensions

1. **Compatibility vs Clarity** - Pragmatist prioritizes compatibility, Architect prioritizes clarity
2. **Gradual vs Complete** - Migration Engineer proposes phases, Architect wants complete solution
3. **Coupling vs Separation** - Layer Librarian defends intentional coupling, Architect wants separation

### Interesting Ideas

1. **Phased approach** (Migration Engineer) - Start with aliases, add packages later
2. **Façade pattern** (Architect) - New packages wrap old implementations
3. **Documentation-first** (Doc Maintainer) - Update docs regardless of code changes
4. **Package rename for middlewares** (Source Chain) - "middlewares" → "sources" or "resolvers"

### Open Questions

1. **Where should FieldDefinition live?** `schema` package or `fields` package?
2. **What do we call the collection?** `SchemaSections` vs `Sections` vs `SchemaGroups`?
3. **What do we call resolved values?** `SectionValues` vs `ResolvedValues` vs `Values`?
4. **Should we rename middlewares package?** If yes, to `sources` or `resolvers`?
5. **What's the deprecation timeline?** Keep old names forever or remove after N versions?

### Recommendations for Next Steps

1. **Immediate (Phase 1):**
   - Add type aliases: `SchemaSection`, `SectionValues`, `FieldDefinition`
   - Update documentation and tutorials
   - Keep old names as compatibility layer

2. **Short-term (Phase 2):**
   - Create new packages with façade pattern
   - Gradually migrate internal code
   - Both import paths work

3. **Long-term (Phase 3):**
   - Consider deprecating old names (but keep forever)
   - Evaluate verb renaming (`DecodeInto`)
   - Consider renaming `middlewares` → `sources`

4. **Documentation:**
   - Update glossary with new vocabulary
   - Rewrite tutorials with clear names
   - Add migration guide

### Consensus Areas

- **Schema/Field/Values vocabulary** - All agree this is clearer
- **Type aliases as first step** - All agree this is safe and useful
- **Documentation updates** - All agree this is necessary
- **Backward compatibility** - All agree we must maintain it

### Disagreement Areas

- **Package moves** - Pragmatist opposes, Architect supports, Migration Engineer proposes phases
- **Verb renaming** - Pragmatist says optional, Architect/New Developer say necessary
- **Separation of concerns** - Architect wants separation, Layer Librarian defends coupling

---

**Next Steps:** Review this debate round, identify areas needing deeper analysis, schedule Round 2 if needed.

