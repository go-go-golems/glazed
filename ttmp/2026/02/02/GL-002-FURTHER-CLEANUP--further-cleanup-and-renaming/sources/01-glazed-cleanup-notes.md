---
Title: Imported cleanup notes (renaming strategy)
Ticket: GL-002-FURTHER-CLEANUP
Status: active
Topics:
    - glazed
    - api-design
    - renaming
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /tmp/glazed-cleanup-notes.md
      Note: Original notes file imported into ticket
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: "Source notes that drive the follow-up renaming plan"
WhenToUse: "When reviewing the rationale for further renames"
---

I read through your rename audit doc  and then sanity‑checked it against the current code in `pkg/cmds/*`. The core problem isn’t really “Schema vs SectionSet” — it’s that the public API still leaks **multiple vocabularies** for the same concepts, so users can’t form a stable mental model.

If the goal is “eliminate confusion,” you need one canonical noun per concept, and you need to stop shipping synonyms.

## What’s actually confusing right now (from an API user’s POV)

### 1) “Layer” is still everywhere

Even after the “Section/Schema” shift, exported identifiers still say **Layer** in key places (e.g., `CommandDescription.Layers`, `AppendLayers`, `AddLayerToCobraCommand`, `GetDefaultParameterLayer`, many `sources.*Layers*` helpers).

That creates immediate ambiguity: **are Layers and Sections different things, or synonyms?** If they’re synonyms, you must not expose both.

### 2) You currently export *two* different “Section” concepts

This one isn’t emphasized in the audit doc but it’s huge in the actual code:

* `schema.Section` = schema definition grouping (the concept users expect)
* `values.Section` = “minimal interface needed by values” (exported, cycle workaround)
* `values.SectionValues.Layer` = points at a section definition but is named **Layer**

So a user sees both `schema.Section` and `values.Section` in GoDoc and has to guess which one they “should” use. That’s guaranteed confusion.

### 3) “ParsedParameters” reads like an intermediate/one-source artifact

But in your model, these objects are the *resolved* values with provenance; they’re not just “raw parsed input.”

If a user sees `ParsedParameters`, they’ll assume:

* it’s tied to a single parse operation, and/or
* it’s not final (so they look for “Resolved…” or “Values…” elsewhere)

### 4) `InitializeStruct` is the wrong verb

You use “initialize” in multiple domains:

* initialize defaults *in schema definitions*
* initialize a struct *from values*

Users don’t know if it’s “set defaults,” “validate,” “decode,” “hydrate,” etc.

“Decode” (or “Bind”) is much clearer and aligns with common Go conventions (json/mapstructure/etc.).

---

## The main principle I’d enforce

Pick one stable vocabulary, then **ban synonyms** in exported names.

Given your intended model, the clean nouns are:

* **Schema**: describes what can be provided
* **Section**: grouping/namespace within the schema
* **Definition**: a field/parameter definition
* **Source**: a step that contributes values (even if implemented as middleware)
* **Values**: resolved values (per section + per command)
* **Decode**: put values into a struct

And then: **the word “Layer” must disappear from exported identifiers**.

That’s the single highest ROI change.

---

## My recommended end-state (concrete rename set)

I’m going to propose a coherent set that (a) matches your mental model, (b) removes “layer”, and (c) removes the “two Section types” trap.

### A) `cmds` (highest priority: entry point)

**Do this first** because it’s what users see first.

* `CommandDescription.Layers` → `CommandDescription.Schema`
* `GetLayer` → `GetSection`
* `GetDefaultLayer` → `GetDefaultSection`
* `Clone(cloneLayers bool, ...)` → `Clone(cloneSchema bool, ...)` (or `cloneSections`)

Options:

* Keep `WithSchema(...)` as the only public option name.
* Rename the others to schema‑specific names (or delete them):

  * `WithLayersList` → `WithSchemaSections(...)` (avoid plain `WithSections` because you already have `layout.Section` in the same struct)
  * `WithLayersMap` → `WithSchemaMap(...)` or `WithSchemaBySlug(...)`
  * `WithReplaceLayers` → `WithReplaceSchema(...)` or `WithReplaceSections(...)`

**Why:** the main struct shouldn’t expose old vocabulary at all. If someone starts here and sees “Layers,” they assume “layers are a thing.”

---

### B) `schema`

I agree with your draft: **keep `Schema` and `Section`**. They match the concept and are familiar.

But clean up the leftovers:

* `AppendLayers` / `PrependLayers` → `AppendSections` / `PrependSections`
* `SectionImpl` → **`SectionDef`** (or `SectionSpec`)

  * “Impl” reads like “internal,” but this is your main building block.
* `ChildLayers` → `ChildSections` (or `Children`)
* Wrapper types still named with “ParameterLayer”:

  * `WhitelistParameterLayer` → `WhitelistSection`
  * `BlacklistParameterLayer` → `BlacklistSection`
  * constructors too (`NewWhitelistSection`, …)

Cobra naming:

* `CobraSection.AddLayerToCobraCommand` → `AddToCobraCommand` (best)
  (or `AddSectionToCobraCommand` if you want verbosity)

**Critical note (design, not just naming):** parsing *from cobra* is a source concern, not a schema concern. If you can move that out of `schema`, you can also fix the `values.Section` cycle workaround (next section).

---

### C) `values` (fix the exported “values.Section” confusion)

This is the biggest “hidden” API confusion right now.

You have two paths:

#### Path 1 (best): break the cycle and delete `values.Section`

If you move “defaults initialization” and “cobra parsing” out of `schema` (into `sources`, where they conceptually belong), then `values` can import `schema` directly and you can:

* remove exported `values.Section` entirely
* make the real association explicit:

  * `SectionValues.Section schema.Section` (instead of `Layer Section`)

This produces the cleanest public API and removes the “which Section?” confusion.

#### Path 2 (minimal): rename the workaround type

If you keep the cycle workaround, **rename the exported interface** to what it actually is:

* `values.Section` → `values.SchemaSection` (or `values.SectionSpec`)
* `SectionValues.Layer` → `SectionValues.Section` (or `Schema`)

That at least stops the API from presenting two different things called “Section.”

Also clean up the “layer” leftovers:

* `GetDefaultParameterLayer` → `DefaultSection()` or `DefaultSectionValues()`

**On `Values` naming:** your audit doc suggests `ResolvedValues` / `CommandValues`. I’d pick **`CommandValues`** if you want to reduce “values.Values” stutter and be explicit about scope. If you keep `Values`, it’s not *wrong*, but it’s less discoverable.

---

### D) `fields` (replace “Parsed*” with “*Value(s)”)

This is where your doc is right to be worried: “Parsed” strongly implies “intermediate.”

I’d avoid the “ResolvedFields” wording unless you want to re-center the whole library around “fields” vocabulary. You consistently say “parameters” elsewhere.

A clean, minimal semantic fix:

* `ParsedParameter` → `ParameterValue`
* `ParsedParameters` → `ParameterValues`
* `InitializeStruct` → `DecodeInto`

Optional (only if you want to reduce “sources vs source” collision):

* `WithSource(...)` → `WithOrigin(...)`
  (not required, but it reads better alongside a package literally named `sources`)

---

### E) `sources`

Make the domain verb obvious: you’re **resolving** values.

* `Execute` → `Resolve`
* `Middleware` → `Source` (domain term)
  (because that’s what users think they’re composing)
* `HandlerFunc` → `ResolveFunc` (or `ResolverFunc`)
* Rename exported `...Layers...` helpers to `...Sections...` everywhere:

  * `WhitelistLayers` → `WhitelistSections`
  * `WhitelistLayerParameters` → `WhitelistSectionParameters` (or `...Fields`, but pick one noun and stick to it)
  * `WrapWithLayerModifyingHandler` → `WrapWithSchemaModifyingHandler` (or `WrapWithSectionModifyingHandler`)

---

## The “happy path” readability test

This is the bar I’d use: can someone read this and not ask “what’s a layer?”

```go
sch := schema.NewSchema(schema.WithSections(defaultSection, loggingSection))

vals := values.New() // -> *values.CommandValues
err := sources.Resolve(
    sch, vals,
    sources.FromDefaults(),
    sources.FromEnv("APP"),
    sources.FromCobra(cmd),
)

var cfg Config
err = values.DecodeSectionInto(vals, schema.DefaultSectionSlug, &cfg)
```

Schema → Resolve → Decode. No “layer.” No “parsed vs values” ambiguity.

---

## Refactor order that gets you the biggest clarity fastest

If you’re unsure where to start, do it in this order:

1. **Eliminate “Layer” from exported API** (cmds/schema/values/sources)
   Biggest immediate confusion reducer.

2. **Fix the `values.Section` situation** (best: break cycle; fallback: rename to `SchemaSection`)
   This prevents users from thinking there are two competing Section concepts.

3. **Rename `InitializeStruct` → Decode** everywhere
   Huge discoverability win.

4. **Rename Parsed* → *Value(s)**
   Aligns names with semantics (“final resolved/typed values with provenance”).

---

If you want, I can also propose a very short package‑doc blurb for `schema`, `sources`, and `values` that encodes the mental model (3–5 lines each). That often matters as much as the names, because users’ first contact is GoDoc, not internal design notes.


--- 

Yeah — **“field” is the better public term here** *if* you’re willing to commit to it consistently.

Right now the API is already half‑telling users “these are fields” (package `fields`, `schema.WithFields(...)`, etc.) but then you surface “parameter” in the most important types (`ParsedParameter(s)`, `glazed` tags, error messages). That mismatch is exactly the kind of “wait… are these different?” confusion you’re trying to eliminate. 

## Why “field” beats “parameter” for *this* API

1. **It matches your conceptual model**
   You have a schema that defines *things you can provide*, sources that fill them, and values that decode into structs. That’s “fields” in almost every Go ecosystem: schema fields → struct fields.

2. **“Parameter” is overloaded in Go + CLI land**
   Users will read “parameter” as:

   * function parameter, or
   * CLI flag parameter, or
   * HTTP query parameter, etc.
     Meanwhile your “parameters” include positional args, env/file inputs, and end up as decoded configuration. “Field” is broader and more accurate.

3. **You already use “field” in the API surface**
   `schema.WithFields(...)` is a big tell. If you keep “parameter” in the core value types, you’re teaching two vocabularies for one concept. 

## The key rule if you switch: don’t mix synonyms

If you decide “field” is the canonical noun, then **remove “parameter” from exported identifiers** (except maybe in very CLI‑specific helper names, but I’d avoid even that).

Here’s the rename set I’d recommend (adjusting my earlier “ParameterValue(s)” suggestion to “FieldValue(s)”):

### `pkg/cmds/fields`

* `Definition` → keep as `Definition` **or** rename to `Field` / `FieldDef`

  * If you keep it: it’s fine because it’s already scoped under package `fields`.
  * If you rename: `fields.Field` is very readable.
* `ParsedParameter` → **`FieldValue`**
  (it’s a value with provenance, not “parsed input”)
* `ParsedParameters` → **`FieldValues`**
* `WithParsedParameter(...)` → **`WithFieldValue(...)`** (or `WithValue(...)` if scoped on `FieldValues`)
* `InitializeStruct` → **`DecodeInto`** (or `Decode`)

### Tag naming (this matters more than it seems)

Your decode tags currently use `glazed:"..."`. If “field” is the term, I’d change the tag to one of:

* **Best ergonomics:** `glazed:"name,from_json"`
  (short, matches common Go patterns like `mapstructure:"..."`)
* Or explicit: `glazed.field:"name,from_json"`

Keeping `glazed` while everything else becomes “field” will reintroduce the “two vocabularies” problem in docs/examples. 

### `pkg/cmds/values`

* `WithParameterValue(...)` → **`WithFieldValue(...)`**
* Any method like `MustUpdateValue(... errors.Errorf("parameter %s not found"...))` should say **field**, not parameter.

### `pkg/cmds/sources`

* Rename helpers that currently say `...Parameters...` to `...Fields...` (e.g., whitelisting, filtering, etc.) so users don’t think “parameters” is a separate concept. 

## One small caution (and how to avoid it)

The only real downside of “field” is it can mean both:

* a *schema field* (your concept), and
* a *Go struct field* (language concept)

But that’s actually workable because the relationship is intentional. You just need docs and GoDoc comments to say **“schema field”** when it matters.

## My recommendation

Adopt **Field** as the universal noun in the public API:

* schema: **Fields** grouped into Sections, collected in a Schema
* resolution: sources fill **FieldValues**
* decoding: **DecodeInto** a struct

And then do a single sweep to remove “parameter” from exported names + tags so there’s only one mental model. 

