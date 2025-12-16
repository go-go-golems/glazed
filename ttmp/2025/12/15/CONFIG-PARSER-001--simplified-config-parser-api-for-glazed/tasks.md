# Tasks

## TODO

- [x] Build a “health inspection” map of Glazed’s parameter parsing architecture (layers/parsed-layers/middlewares/cobra integration)
- [x] Maintain a detailed lab-book diary of searches + findings while researching the new API (including Moments `appconfig` prior art)
- [x] Write a design doc proposing a struct-first ConfigParser API (including modify-vs-layer decision + naming/renaming suggestions + 1–3 implementation designs)
- [x] Superseded (appconfig.Parser redesign): Decide on the internal layering model (single layer vs per-sub-struct layers) for the first implementation
- [x] Superseded (appconfig.Parser redesign): Implement a struct-first ConfigParser package in Glazed (derive schema from structs, generate layers + mapper, parse + hydrate typed config)
- [x] Superseded (appconfig.Parser redesign): Add minimal upstream hooks in Glazed for config-mapper wiring (CobraParserConfig and/or runner.ParseCommandParameters)
- [x] Superseded (appconfig.Parser redesign): Add a purpose-built mapper for struct path → layer/param (avoid forcing patternmapper prefix semantics)
- [x] Superseded (appconfig.Parser redesign): Add examples + docs (one small app; plus a pinocchio migration spike)

- [x] P0: Confirm appconfig.Parser[T] public API + package placement (glazed/pkg/appconfig)
- [x] P0: Decide Parse signature and ergonomics (Parse() vs Parse(ctx), what options belong on NewParser vs Parse)
- [x] P1: Add core type skeleton type Parser[T any] and internal registration model (slug, layer, bind func(*T) any)
- [x] P1: Implement NewParser[T](opts...) and ParserOption types
- [x] P1: Implement option helpers mapping to runner ParseOptions (WithEnv, WithConfigFiles, WithValuesForLayers, WithAdditionalMiddlewares / WithRunnerParseOptions)
- [x] P1: Implement Register(slug, layer, bind) with validation (duplicate slug, nil layer, nil bind)
- [x] P1: Build ParameterLayers collection from registered layers (stable iteration order)
- [x] P1: Implement minimal cmds.Command stub (Description().Layers = registered layers) to reuse runner.ParseCommandParameters
- [x] P1: Implement Parse() to run configured middleware chain (runner.ParseCommandParameters or explicit ExecuteMiddlewares)
- [x] P1: Implement hydration into grouped T via ParsedLayers.InitializeStruct(reg.Slug, reg.Bind(&t)) with per-layer error context
- [x] P1: Decide and document v1 contract: requires glazed.parameter tags; missing params are skipped (zero values)
- [x] P2: Add unit tests for registration invariants and binder failures (bind returns nil or non-pointer)
- [x] P2: Add unit tests for precedence (defaults < config files low→high < env) using temporary YAML config files
- [x] P2: Add unit tests for hydration behavior (tag-required, missing params skipped)
- [ ] P3: Add a minimal example (glazed/cmd/examples or pinocchio/cmd/examples) showing two registered layers hydrating into AppSettings
- [ ] P3: Add docs: quickstart snippet for appconfig.Parser usage and where it fits vs CobraParserConfig
- [ ] P4: Optional: spike a tiny pinocchio integration replacing manual runner+InitializeStruct wiring with appconfig.Parser
