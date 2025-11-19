---
Title: 'Analysis: Environment parsing in glazed'
Ticket: MEN-20251119
Status: active
Topics:
    - glazed
    - parameters
    - middleware
    - env
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Investigate parsing env strings to typed params; plan updates to UpdateFromEnv
LastUpdated: 2025-11-18T21:17:08.213582239-05:00
---

# Analysis: Environment parsing in glazed

<!-- Add your content here -->

### Context

- Glazed env middleware currently writes raw strings via UpdateValue, which validates types but expects typed inputs for non-strings.
- We must parse env strings per ParameterType (bool/int/float/date/choice/lists/key-value).

### Proposed Approach

- For each parameter definition in a layer, when ENV var is set, map value string -> []string and call p.ParseParameter([]string{value}) to reuse existing parsing semantics.
- Update parsedLayer.Parameters with parsed.Value and append parse step metadata {env_key: ..., source: env}.

### Acceptance Criteria

- Bool/int/float/date/choice and list types parse correctly from ENV strings.
- Unit tests cover representative types and failure cases.
- Docs describe ENV name formation: [PREFIX_]LAYERPREFIX + NAME (hyphens->underscores, uppercased).
