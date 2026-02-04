# Layer/Parameter Inventory

Root: `/home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed`

Total files with matches: **161**
- go: 103
- doc: 44
- data: 9
- other: 5

## GO files

### `cmd/examples/appconfig-parser/main.go`
Identifiers:

- LayerSlug
- WithValuesForLayers
- dbLayer
- layer
- redisLayer

### `cmd/examples/appconfig-profiles/main.go`
Identifiers:

- NewProfileSettingsLayer
- addLayer
- cobraLayer
- layer
- psLayer
- redisLayer

### `cmd/examples/config-custom-mapper/main.go`
Identifiers:

- Layers
- SkipCommandSettingsLayer
- WithLayersList
- demoLayer
- layer
- layerSlug
- layers
- parameter
- parameters
- parsedCommandLayers

### `cmd/examples/config-overlay/main.go`
Identifiers:

- Layers
- SkipCommandSettingsLayer
- WithLayersList
- layer
- layerSlug
- parameter

### `cmd/examples/config-pattern-mapper/main.go`
Identifiers:

- Layer
- LoadParametersFromFile
- TargetLayer
- TargetParameter
- demoLayer
- paramLayers

### `cmd/examples/config-single/main.go`
Identifiers:

- Layers
- SkipCommandSettingsLayer
- WithLayersList
- layer
- layerSlug
- layers
- parameter
- parameters

### `cmd/examples/middlewares-config-env/main.go`
Identifiers:

- WithLayersList
- layer
- parameters

### `cmd/examples/parameter-types/main.go`
Identifiers:

- NewParameterTypesCommand
- Parameter
- ParameterTypesCommand
- ParameterTypesSettings
- WithLayersList
- layer
- parameter
- parameterData
- parameter_name
- parameter_type
- parameters

### `cmd/examples/refactor-new-packages/main.go`
Identifiers:

- parameters

### `cmd/examples/sources-example/main.go`
Identifiers:

- Layers

### `cmd/glaze/cmds/csv.go`
Identifiers:

- WithLayersList
- glazedLayer
- parameters

### `cmd/glaze/cmds/docs.go`
Identifiers:

- WithFieldsFiltersParameterLayerOptions
- cobraLayer
- glazedLayer
- layer
- parameter

### `cmd/glaze/cmds/example.go`
Identifiers:

- WithLayersList
- glazedLayer
- layer
- layers
- parameter
- parameters
- parsedLayers

### `cmd/glaze/cmds/html/cmds.go`
Identifiers:

- cobraLayer
- glazedLayer
- layer

### `cmd/glaze/cmds/json.go`
Identifiers:

- WithLayersList
- glazedLayer
- layer
- parameter
- parameters

### `cmd/glaze/cmds/markdown.go`
Identifiers:

- cobraLayer
- glazedLayer
- layer

### `cmd/glaze/cmds/yaml.go`
Identifiers:

- WithLayersList
- glazedLayer
- layer
- parameter
- parameters

### `pkg/appconfig/doc.go`
Identifiers:

- ParameterLayers
- layers

### `pkg/appconfig/options.go`
Identifiers:

- NewProfileSettingsLayer
- WithValuesForLayers
- bootstrapLayers
- layer
- layers
- layers_
- parsedLayers
- psLayer

### `pkg/appconfig/parser.go`
Identifiers:

- LayerSlug
- ParameterLayer
- layer
- layers
- paramLayers
- parameters
- parsedLayers

### `pkg/appconfig/parser_test.go`
Identifiers:

- LayerSlug
- WithValuesForLayers
- cobraLayer
- layer
- newTestRedisLayer

### `pkg/appconfig/profile_test.go`
Identifiers:

- LayerSlug
- NewProfileSettingsLayer
- layer
- newTestRedisLayer
- psLayer

### `pkg/cli/cli.go`
Identifiers:

- LoadParametersFromFile
- NewCommandSettingsLayer
- NewCreateCommandSettingsLayer
- NewProfileSettingsLayer
- PrintParsedParameters
- createCommandSettingsLayer
- glazedMinimalCommandLayer
- parameter
- parameters
- profileSettingsLayer

### `pkg/cli/cliopatra/capture.go`
Identifiers:

- Layers
- Parameter
- ParameterDefinition
- getCliopatraParameters
- layer
- layered
- layers
- parameter
- parameters_
- parsedLayer
- parsedLayers

### `pkg/cli/cliopatra/capture_test.go`
Identifiers:

- GetLayer
- TestSingleLayer
- WithLayersList
- layer
- makeParsedDefaultLayer

### `pkg/cli/cliopatra/program.go`
Identifiers:

- Parameter
- Parameters
- parameter
- parsedLayers

### `pkg/cli/cobra-parser.go`
Identifiers:

- CommandSettingsLayer
- CreateCommandSettingsLayer
- EnableCreateCommandSettingsLayer
- EnableProfileSettingsLayer
- Layer
- Layers
- LoadParametersFromResolvedFilesForCobra
- NewCobraParserFromLayers
- NewCommandSettingsLayer
- NewCreateCommandSettingsLayer
- NewProfileSettingsLayer
- ParseCommandSettingsLayer
- ParseGlazedCommandLayer
- ProfileSettingsLayer
- ShortHelpLayers
- SkipCommandSettingsLayer
- cobraLayer
- commandSettingsLayer
- commandSettingsLayers
- createCommandSettingsLayer
- enableCreateCommandSettingsLayer
- enableProfileSettingsLayer
- layer
- layers
- paramLayers
- parameters
- parsedCommandLayers
- parsedLayers
- profileSettingsLayer
- shortHelpLayers
- shortHelperLayer
- skipCommandSettingsLayer

### `pkg/cli/cobra.go`
Identifiers:

- EnableCreateCommandSettingsLayer
- EnableProfileSettingsLayer
- Layers
- NewCobraParserFromLayers
- NewGlazedParameterLayers
- PrintParsedParameters
- ShortHelpLayers
- SkipCommandSettingsLayer
- WithCobraShortHelpLayers
- WithCreateCommandSettingsLayer
- WithProfileSettingsLayer
- WithSkipCommandSettingsLayer
- createLayer
- glLayer
- glazedLayer
- glazedLayers
- layer
- layers
- layers_
- minimalLayer
- originalLayers
- parameter
- parsedLayers
- printParsedParameters
- printParsedParameters_

### `pkg/cli/helpers.go`
Identifiers:

- GlazeParameterLayerOption
- Layers
- NewCobraParserFromLayers
- NewGlazedParameterLayers
- layer
- layerName
- layersMap
- layers_
- parameter
- parsedLayer
- parsedLayers
- printParsedParameters

### `pkg/cmds/alias/alias.go`
Identifiers:

- Layers
- newLayers
- parsedLayers

### `pkg/cmds/cmds.go`
Identifiers:

- GetDefaultLayer
- GetLayer
- GlazeLayer
- Layers
- ParameterLayerImpl
- SetLayers
- WithLayers
- WithLayersList
- WithLayersMap
- WithReplaceLayers
- cloneLayers
- layer
- layers
- layers_
- parameter
- parsedLayers

### `pkg/cmds/cobra_test.go`
Identifiers:

- ExpectedArgumentParameters
- ExpectedFlagParameters
- GetDefaultLayer
- Layers
- WithLayersList
- argumentParameters
- defaultLayer
- defaultLayer_
- flagParameters
- layer
- parameters

### `pkg/cmds/fields/cobra.go`
Identifiers:

- Parameter
- Parameters
- parameter
- parameters

### `pkg/cmds/fields/errors.go`
Identifiers:

- parameter

### `pkg/cmds/fields/gather-arguments.go`
Identifiers:

- parameter

### `pkg/cmds/fields/gather-arguments_test.go`
Identifiers:

- TestGatherArguments_ListParameterParsing
- TestSingleParametersFollowedByListDefaults
- TestThreeSingleParametersFollowedByListDefaults
- TestThreeSingleParametersFollowedByListDefaultsOnlyTwoValues
- parameter

### `pkg/cmds/fields/gather-parameters.go`
Identifiers:

- parameter

### `pkg/cmds/fields/gather-parameters_test.go`
Identifiers:

- ParameterDefs
- TestGatherParametersFromMap
- gatherParametersYAML
- parameterDefs
- parameters

### `pkg/cmds/fields/initialize-struct.go`
Identifiers:

- Parameters
- parameter
- parameters

### `pkg/cmds/fields/initialize-struct_test.go`
Identifiers:

- TestInitializeStructWithMissingParameters
- parameters

### `pkg/cmds/fields/parameter-type.go`
Identifiers:

- parameter

### `pkg/cmds/fields/parameters.go`
Identifiers:

- Parameter
- parameter
- parameterDefinitions
- parameterType
- parameters
- parameters_

### `pkg/cmds/fields/parameters_from_defaults_test.go`
Identifiers:

- TestParsedParametersFromDefaults_BasicTypes
- TestParsedParametersFromDefaults_EdgeCases
- TestParsedParametersFromDefaults_EmptyCollections
- TestParsedParametersFromDefaults_FileLoadingTypes
- TestParsedParametersFromDefaults_ListTypes
- TestParsedParametersFromDefaults_MapTypes
- TestParsedParametersFromDefaults_NilComplexTypes
- parameter
- parameterDefinitions
- parameters

### `pkg/cmds/fields/parse.go`
Identifiers:

- parameter
- parameters

### `pkg/cmds/fields/parse_test.go`
Identifiers:

- ParameterBool
- ParameterChoice
- ParameterChoiceList
- ParameterFloat
- ParameterFloatList
- ParameterInt
- ParameterIntegerList
- ParameterString
- ParameterStringList
- ParameterTest
- ParameterTestCase
- TestParameterDate
- TestParseParameter
- parameter

### `pkg/cmds/fields/parsed-parameter.go`
Identifiers:

- parameter
- parameterType

### `pkg/cmds/fields/serialize.go`
Identifiers:

- parameter
- parameters

### `pkg/cmds/fields/strings.go`
Identifiers:

- parameter
- parameters

### `pkg/cmds/fields/strings_test.go`
Identifiers:

- EmptyParameters
- MixOfValidAndInvalidParameters
- ParametersWithDifferentNameSameShortFlag
- ParametersWithEmptyShortFlag
- TestGatherFlagsFromStringList_ValidArgumentsAndParameters
- parameter
- parameters

### `pkg/cmds/fields/viper.go`
Identifiers:

- LoadParametersFromFiles
- parameter

### `pkg/cmds/helpers/test-helpers.go`
Identifiers:

- BlacklistLayerParameters
- BlacklistLayerParametersFirst
- BlacklistLayers
- BlacklistLayersFirst
- Layers
- NewTestParameterLayer
- NewTestParameterLayers
- ParameterLayer
- ParameterLayers
- TestBlacklistLayerParameters
- TestBlacklistLayerParametersFirst
- TestBlacklistLayers
- TestBlacklistLayersFirst
- TestExpectedLayer
- TestParameterLayer
- TestParsedParameter
- TestWhitelistLayerParameters
- TestWhitelistLayerParametersFirst
- TestWhitelistLayers
- TestWhitelistLayersFirst
- WhitelistLayerParameters
- WhitelistLayerParametersFirst
- WhitelistLayers
- WhitelistLayersFirst
- blacklistLayerParameters
- blacklistLayerParametersFirst
- blacklistLayers
- blacklistLayersFirst
- expectedLayers
- expectedLayers_
- layer
- layer1
- layer2
- layers
- layers_
- parameter
- parameterDefinition
- parameterLayers
- parameters
- parsedLayers
- whitelistLayerParameters
- whitelistLayerParametersFirst
- whitelistLayers
- whitelistLayersFirst

### `pkg/cmds/json-schema.go`
Identifiers:

- parameter
- parameterTypeToJsonSchema

### `pkg/cmds/layout/layout.go`
Identifiers:

- layer
- parameter

### `pkg/cmds/logging/init-early.go`
Identifiers:

- AddLoggingLayerToRootCommand
- layer

### `pkg/cmds/logging/init.go`
Identifiers:

- AddLoggingLayerToRootCommand
- layers

### `pkg/cmds/logging/layer.go`
Identifiers:

- AddLoggingLayerToCommand
- AddLoggingLayerToRootCommand
- Layers
- LoggingLayerSlug
- NewLoggingLayer
- layer
- loggingLayer
- parameter
- parameters
- parsedLayers

### `pkg/cmds/runner/run.go`
Identifiers:

- Layers
- ParseCommandParameters
- ValuesForLayers
- WithValuesForLayers
- glazedLayer
- layer
- layers
- parameter
- parameters
- parsedLayers

### `pkg/cmds/schema/cobra_flag_groups.go`
Identifiers:

- ParameterDefinition
- Parameters
- layers

### `pkg/cmds/schema/errors.go`
Identifiers:

- layer
- parameter

### `pkg/cmds/schema/layer-impl.go`
Identifiers:

- ChildLayers
- InitializeDefaultsFromParameters
- InitializeStructFromParameterDefaults
- ParameterDefinition
- ParseLayerFromCobraCommand
- childLayers
- layer
- parameter
- parameterDefinitions
- parameters

### `pkg/cmds/schema/layer-impl_test.go`
Identifiers:

- createSimpleParameterLayer
- layer

### `pkg/cmds/schema/layer.go`
Identifiers:

- AppendLayers
- PrependLayers
- layer
- layers
- parameter
- parsedLayer
- parsedLayers

### `pkg/cmds/schema/layer_test.go`
Identifiers:

- AppendLayers
- Layer
- PrependLayers
- TestNewParameterLayers
- TestParameterLayersAppendLayers
- TestParameterLayersAsList
- TestParameterLayersClone
- TestParameterLayersForEach
- TestParameterLayersForEachE
- TestParameterLayersGetAllDefinitions
- TestParameterLayersMerge
- TestParameterLayersMergeWithOverlappingLayers
- TestParameterLayersPrependLayers
- TestParameterLayersSubset
- TestParameterLayersSubsetWithNonExistentLayers
- TestParameterLayersWithDuplicateSlugs
- TestParameterLayersWithLargeNumberOfLayers
- TestParameterLayersWithLayers
- TestParameterLayersWithUnicodeLayerNames
- createParameterLayer
- layer
- layer0
- layer1
- layer1Duplicate
- layer2
- layer3
- layers
- layers1
- layers2
- numLayers
- parameter

### `pkg/cmds/sources/cobra.go`
Identifiers:

- LoadParametersFromResolvedFilesForCobra
- ParseLayerFromCobraCommand
- cobraLayer
- defaultLayer
- layer
- layer_prefix
- layer_slug
- layers_
- parameter
- parameterDefinitions
- parameters
- parsedCommandLayers
- parsedLayer
- parsedLayers

### `pkg/cmds/sources/config-mapper-interface.go`
Identifiers:

- layer

### `pkg/cmds/sources/custom-profiles_test.go`
Identifiers:

- layer
- layers
- parameter
- parameterLayers
- parsedLayer
- parsedLayers

### `pkg/cmds/sources/layers.go`
Identifiers:

- layer
- layerSlug
- layerToMerge
- layers
- layersToMerge
- layers_
- newLayer
- newLayers
- parsedLayers
- targetLayer

### `pkg/cmds/sources/load-parameters-from-json.go`
Identifiers:

- layer
- layerData
- layerMap
- layerSlug
- layers
- layers_
- parameter
- parameterName
- parameters
- parsedLayers
- readConfigFileToLayerMap

### `pkg/cmds/sources/middlewares.go`
Identifiers:

- ParameterLayers
- clonedLayers
- layer
- layers
- layers_
- parameter
- parameters
- parsedLayers

### `pkg/cmds/sources/middlewares_test.go`
Identifiers:

- ExpectedLayers
- NewTestParameterLayers
- ParameterLayers
- TestExpectedLayer
- TestParameterLayer
- TestWrapWithRestrictedLayers
- WrapWithBlacklistedLayers
- WrapWithWhitelistedLayers
- expectedLayers
- layers
- layers_
- parameterLayers
- parsedLayers
- wrapWithRestrictedLayersTest
- wrapWithRestrictedLayersTestsYAML

### `pkg/cmds/sources/patternmapper/loader.go`
Identifiers:

- TargetLayer
- TargetParameter
- layers
- layers_
- target_layer
- target_parameter

### `pkg/cmds/sources/patternmapper/pattern_mapper.go`
Identifiers:

- TargetLayer
- TargetParameter
- layer
- layers
- parameter
- parameters
- resolveCanonicalParameterName
- resolveTargetParameter
- targetParameter

### `pkg/cmds/sources/patternmapper/pattern_mapper_builder.go`
Identifiers:

- TargetLayer
- TargetParameter
- layers
- targetLayer
- targetParameter

### `pkg/cmds/sources/patternmapper/pattern_mapper_edge_cases_test.go`
Identifiers:

- Layer
- TargetLayer
- TargetParameter
- TestLayerPrefix
- createTestLayers
- layer
- layers
- parameter
- parameters
- setupLayers
- testLayers

### `pkg/cmds/sources/patternmapper/pattern_mapper_loader_test.go`
Identifiers:

- Layers
- TargetLayer
- buildTestLayers
- layer
- target_layer
- target_parameter

### `pkg/cmds/sources/patternmapper/pattern_mapper_orderedmap_test.go`
Identifiers:

- Layer
- TargetLayer
- TargetParameter
- layer

### `pkg/cmds/sources/patternmapper/pattern_mapper_proposals_test.go`
Identifiers:

- Layer
- TargetLayer
- TargetParameter
- layer
- layerMulti
- layers
- parameter
- parameters
- testLayers
- testLayersMulti

### `pkg/cmds/sources/patternmapper/pattern_mapper_test.go`
Identifiers:

- Layer
- ResolveTargetParameter
- TargetLayer
- TargetParameter
- TestIntegrationWithLoadParametersFromFile
- TestResolveTargetParameter
- createTestLayers
- demoLayer
- layer
- layers_
- parameter
- testLayers

### `pkg/cmds/sources/update.go`
Identifiers:

- layer
- layerPrefix
- layers
- layers_
- parameter
- parsedLayer
- parsedLayers

### `pkg/cmds/sources/update_test.go`
Identifiers:

- cfgLayer
- layer
- parameter

### `pkg/cmds/sources/whitelist.go`
Identifiers:

- BlacklistLayerParameters
- BlacklistLayerParametersFirst
- BlacklistLayerParametersHandler
- BlacklistLayers
- BlacklistLayersFirst
- BlacklistLayersHandler
- ParameterLayers
- WhitelistLayerParameters
- WhitelistLayerParametersFirst
- WhitelistLayerParametersHandler
- WhitelistLayers
- WhitelistLayersFirst
- WhitelistLayersHandler
- WrapWithBlacklistedLayers
- WrapWithBlacklistedParameterLayers
- WrapWithLayerModifyingHandler
- WrapWithWhitelistedLayers
- WrapWithWhitelistedParameterLayers
- clonedLayers
- layer
- layers
- layersToDelete
- layersToUpdate
- layers_
- parameters
- parametersToKeep
- parameters_
- parsedLayers

### `pkg/cmds/template.go`
Identifiers:

- Layers
- WithLayersList
- defaultLayer
- layers
- parsedLayers

### `pkg/cmds/values/parsed-layer.go`
Identifiers:

- layer
- parameter

### `pkg/cmds/values/parsed-layer_test.go`
Identifiers:

- Layer
- TestValuesGetOrCreateNilLayer
- layer
- layer1
- layer2
- parameter
- parsedLayer
- parsedLayer1
- parsedLayer2
- parsedLayers
- sameLayer

### `pkg/config/editor.go`
Identifiers:

- parameter

### `pkg/help/help.go`
Identifiers:

- Parameters

### `pkg/help/store/compat.go`
Identifiers:

- layer

### `pkg/helpers/maps/maps.go`
Identifiers:

- parameter
- parameterName

### `pkg/helpers/markdown/markdown.go`
Identifiers:

- parameter

### `pkg/helpers/templating/templating.go`
Identifiers:

- parameter
- toUrlParameter

### `pkg/lua/cmds.go`
Identifiers:

- Layers
- glazedLayer
- layer
- layerName
- layerTable
- layers
- layersTable
- parameter
- parameters
- parsedLayers

### `pkg/lua/lua.go`
Identifiers:

- ParseLuaTableToLayer
- ParseParameterFromLua
- layer
- layerName
- layerTable
- layers
- layers_
- parameter
- parameterLayers
- parameters
- parsedLayer
- parsedLayers

### `pkg/settings/glazed_layer.go`
Identifiers:

- ChildLayers
- FieldsFiltersParameterLayer
- GlazeParameterLayerOption
- GlazedParameterLayers
- JqParameterLayer
- NewFieldsFiltersParameterLayer
- NewGlazedParameterLayers
- NewJqParameterLayer
- NewJqSettingsFromParameters
- NewOutputParameterLayer
- NewRenameParameterLayer
- NewRenameSettingsFromParameters
- NewReplaceParameterLayer
- NewReplaceSettingsFromParameters
- NewSelectParameterLayer
- NewSelectSettingsFromParameters
- NewSkipLimitParameterLayer
- NewSkipLimitSettingsFromParameters
- NewSortParameterLayer
- NewSortSettingsFromParameters
- NewTemplateParameterLayer
- OutputParameterLayer
- ParseLayerFromCobraCommand
- RenameParameterLayer
- ReplaceParameterLayer
- SelectParameterLayer
- SkipLimitParameterLayer
- SortParameterLayer
- TemplateParameterLayer
- WithFieldsFiltersParameterLayerOptions
- WithJqParameterLayerOptions
- WithOutputParameterLayerOptions
- WithRenameParameterLayerOptions
- WithReplaceParameterLayerOptions
- WithSelectParameterLayerOptions
- WithSkipLimitParameterLayerOptions
- WithSortParameterLayerOptions
- WithTemplateParameterLayerOptions
- fieldsFiltersParameterLayer
- glazedLayer
- jqParameterLayer
- layer
- layers
- outputParameterLayer
- parsedLayer
- renameParameterLayer
- replaceParameterLayer
- selectParameterLayer
- skipLimitParameterLayer
- sortParameterLayer
- templateParameterLayer

### `pkg/settings/settings_fields-filters.go`
Identifiers:

- FieldsFiltersParameterLayer
- InitializeStructFromParameterDefaults
- NewFieldsFiltersParameterLayer
- ParseLayerFromCobraCommand
- glazedLayer
- layer
- parameter

### `pkg/settings/settings_jq.go`
Identifiers:

- JqParameterLayer
- NewJqParameterLayer
- NewJqSettingsFromParameters
- glazedLayer
- layer
- parameter
- parameters

### `pkg/settings/settings_output.go`
Identifiers:

- NewOutputParameterLayer
- OutputParameterLayer
- glazedLayer
- layer

### `pkg/settings/settings_rename.go`
Identifiers:

- NewRenameParameterLayer
- NewRenameSettingsFromParameters
- RenameParameterLayer
- glazedLayer
- layer
- parameter

### `pkg/settings/settings_replace.go`
Identifiers:

- NewReplaceParameterLayer
- NewReplaceSettingsFromParameters
- ReplaceParameterLayer
- glazedLayer
- layer
- parameters

### `pkg/settings/settings_select.go`
Identifiers:

- NewSelectParameterLayer
- NewSelectSettingsFromParameters
- SelectParameterLayer
- glazedLayer
- layer
- parameter
- parameters

### `pkg/settings/settings_skip_limit.go`
Identifiers:

- NewSkipLimitParameterLayer
- NewSkipLimitSettingsFromParameters
- SkipLimitParameterLayer
- glazedLayer
- layer
- parameter
- parameters

### `pkg/settings/settings_sort.go`
Identifiers:

- NewSortParameterLayer
- NewSortSettingsFromParameters
- SortParameterLayer
- glazedLayer
- layer
- parameter
- parameters

### `pkg/settings/settings_template.go`
Identifiers:

- GlazedTemplateLayerSlug
- NewTemplateParameterLayer
- TemplateParameterLayer
- layer
- parameter

### `pkg/settings/settings_template_test.go`
Identifiers:

- GlazedTemplateLayerSlug
- NewTemplateParameterLayer
- layer
- layers_
- parsedLayers

## DOC files

### `AGENT.md`
Snippets:

- Don't add backwards compatibility layers unless explicitly asked.

### `README.md`
Snippets:

- Glazed is a comprehensive Go framework for building command-line applications that handle structured data elegantly. It provides a rich command system, flexible parameter management, multiple output formats, and an integrated help system.
- ### Parameter Layer System
- Organize command parameters into reusable, composable layers:
- - Type-safe parameter extraction
- ## Parameter Layers

### `changelog.md`
Snippets:

- # Parameter Layer Serialization
- Added ability to serialize parameter layers to YAML/JSON format for better interoperability and configuration management.
- - Added SerializableLayers struct for serializing collections of layers as a map keyed by slug
- - Updated serialization to maintain layer order while providing slug-based access
- # Parsed Parameters Serialization

### `cmd/examples/config-pattern-mapper/README.md`
Snippets:

- The pattern mapper allows you to declaratively map config file structures to layer parameters using pattern matching rules, without writing custom Go functions.
- 2. **Named Captures**: Extract values from config paths and use them in parameter names
- mapper, err := patternmapper.NewConfigMapper(layers,
- b := patternmapper.NewConfigMapperBuilder(layers).
- mapper, err := patternmapper.LoadMapperFromFile(layers, "mappings.yaml")

### `cmd/examples/parameter-types/README.md`
Snippets:

- # Parameter Types Example
- This example demonstrates all parameter types available in the glazed framework.
- go build -o parameter-types .
- # Show help to see all available parameters
- ./parameter-types --help

### `cmd/examples/parameter-types/sample-text.txt`
Snippets:

- Perfect for testing string-from-file parameters.

### `pkg/cmds/logging/README.md`
Snippets:

- # Clay Logging Layer
- This package provides a Glazed parameter layer for configuring logging in Clay applications.
- **📖 For API reference and detailed usage**, see: [Logging Layer API Reference](../../doc/reference/logging-layer.md)
- **🎓 To learn how to create custom layers**, see: [Custom Layer Tutorial](../../doc/tutorials/custom-layer.md)
- The logging layer provides:

### `pkg/doc/applications/03-user-store-command.md`
Snippets:

- - **Configuration via YAML**: Define commands and their parameters using YAML files for easy customization.
- "github.com/go-go-golems/glazed/pkg/cmds/layers"
- "github.com/go-go-golems/glazed/pkg/cmds/parameters"
- "github.com/go-go-golems/glazed/pkg/cmds/layers"
- "github.com/go-go-golems/glazed/pkg/cmds/parameters"

### `pkg/doc/topics/01-help-system.md`
Snippets:

- //   Parameters: [1 database]

### `pkg/doc/topics/03-templates.md`
Snippets:

- - `toUrlParameter(v interface{}) string` - Convert value to URL parameter format

### `pkg/doc/topics/06-usage-string.md`
Snippets:

- In contrast to parameter flags, which are preceded by `--` for `-`, arguments are
- - Parameters accepting list inputs should not directly follow each other.
- ### Required Parameters
- ### Optional Parameters
- ### List Parameters

### `pkg/doc/topics/07-load-parameters-from-json.md`
Snippets:

- Title: Loading Parameters from JSON
- Slug: load-parameters-json
- Short: Explains how to load parameters from a JSON file.
- - Parameters
- - load-parameters-from-json

### `pkg/doc/topics/08-file-parameter-type.md`
Snippets:

- Slug: file-parameters
- Short: Describes how to work with file inputs in command parameters.
- - Parameters
- Glazed provides two new parameter types `file` and `fileList` that allow passing file paths which will be automatically loaded and parsed.
- File parameters are parsed into a single or a list of `FileData` structures which can then be accessed from within a template.

### `pkg/doc/topics/09-gather-flags-from-string-list.md`
Snippets:

- ## Parameters
- - `params`: a slice of `*ParameterDefinition` representing the parameter definitions.
- The function returns a map where the keys are the parameter names and the values are the parsed values. If a flag is not recognized or its value cannot be parsed, an error is returned.
- In this example, the function parses the `--verbose` and `-o` flags according to the provided parameter definitions. The `--verbose` flag is a boolean flag and is set to "true". The `-o` flag is a string flag and its value is "file.txt".

### `pkg/doc/topics/10-template-command.md`
Snippets:

- A TemplateCommand allows you to define commands that render Go template text using command-line parameters as template variables. This enables rapid prototyping of text generation tools without writing Go code—simply define parameters in YAML and write a template that uses those parameters.
- Template commands are defined in YAML files with a `template` field containing Go template syntax. The template receives all parsed parameters as variables accessible through the standard `{{.variable}}` syntax.
- // Provide parameter values for the "default" layer
- Template commands use Go's `text/template` package syntax. All parsed parameters are available as variables in the template context.
- # Using defaults for optional parameters

### `pkg/doc/topics/12-profiles-use-code.md`
Snippets:

- Profile middleware in Pinocchio is responsible for loading and applying configuration parameters from a specified
- The middleware will then load the configuration parameters from the `development` profile and apply them to the command.

### `pkg/doc/topics/13-layers-and-parsed-layers.md`
Snippets:

- Title: Parameter Layers and Parsed Layers
- Slug: parameter-layers-and-parsed-layers
- Learn how to use parameter layers and parsed layers in Glazed to organize and manage parameter definitions.
- - layers
- ## Parameter Layers

### `pkg/doc/topics/15-profiles.md`
Snippets:

- Use profiles.yaml to apply named configuration bundles across parameter layers, with predictable precedence and debugging.
- Profiles are a **named bundle of parameter overrides** stored in a YAML file (typically `profiles.yaml`).
- - **Second level**: layer slug
- - **Third level**: parameter name/value pairs for that layer
- In a Cobra CLI built with Glazed, profile selection typically comes from the **ProfileSettings layer**:

### `pkg/doc/topics/16-adding-parameter-types.md`
Snippets:

- Title: Adding New Parameter Types to Glazed
- Slug: adding-parameter-types
- Short: Comprehensive guide on implementing new parameter types in the Glazed framework.
- - parameters
- # Adding New Parameter Types to Glazed

### `pkg/doc/topics/16-parsing-parameters.md`
Snippets:

- Title: Parsing Parameters
- Slug: parsing-parameters
- Short: Learn how to define and parse parameters in Go applications using the Parameter API.
- - Parameter API
- The **Parameter API** facilitates parsing and managing parameters in Go applications. It's ideal for applications requiring flexible parameter handling.

### `pkg/doc/topics/18-lua.md`
Snippets:

- Executes a GlazeCommand with parameters from a Lua table.
- Executes a BareCommand with parameters from a Lua table.
- Executes a WriterCommand with parameters from a Lua table.
- func ParseLuaTableToLayer(L *lua.LState, luaTable *lua.LTable, layer schema.Section) (*values.SectionValues, error)
- Parses a Lua value into a Go value based on the parameter definition.

### `pkg/doc/topics/19-writing-yaml-commands.md`
Snippets:

- ## Parameter Types
- Glazed supports these parameter types for both flags and arguments:
- Flags are optional parameters that modify command behavior. Here's a comprehensive example:
- - `type`: Parameter type (required)
- Arguments are positional parameters. They're defined similarly to flags:

### `pkg/doc/topics/21-cmds-middlewares.md`
Snippets:

- Short: Learn how to use Glazed's middleware system to load parameter values from various sources
- - parameters
- # Glazed Middlewares Guide: Loading Parameter Values
- Glazed provides a flexible middleware system for loading parameter values from various sources. This guide explains how to use these middlewares effectively to populate your command parameters from different locations like environment variables, config files, and command line arguments.
- type HandlerFunc func(layers *schema.Schema, parsedLayers *values.Values) error

### `pkg/doc/topics/22-command-loaders.md`
Snippets:

- "github.com/go-go-golems/glazed/pkg/cmds/parameters"
- -   **`SqlCommandLoader` (`github.com/go-go-golems/sqleton/pkg/cmds`)**: Loads SQL execution commands for the `sqleton` tool from YAML files containing SQL queries and parameter definitions. Uses `loaders.CheckYamlFileType(f, fileName, "sqleton")` in `IsFileSupported`.

### `pkg/doc/topics/22-templating-helpers.md`
Snippets:

- ps map[string]interface{}, // Parameters for the template

### `pkg/doc/topics/23-pattern-based-config-mapping.md`
Snippets:

- Short: Declarative mapping of config files to parameter layers using pattern matching rules
- The pattern-based config mapping system provides a declarative way to map arbitrary config file structures to Glazed's layer-based parameter system without writing custom Go functions. Instead of implementing `ConfigFileMapper` functions with manual config traversal, you define mapping rules that specify patterns to match in config files and how to map matched values to parameters. This keeps configuration logic concise, testable, and consistent across commands.
- - Verify target layers exist and static target parameters are valid (prefix-aware)
- - For each pattern, collect matches; resolve `{captures}` into parameter names
- - Write values to the target layer/parameter; error on ambiguity or collisions

### `pkg/doc/topics/24-config-files.md`
Snippets:

- - Traceability: Each config file write is logged with `source: config` and `{ config_file, index }` metadata and can be inspected with `--print-parsed-parameters`.
- "github.com/go-go-golems/glazed/pkg/cmds/layers"
- "github.com/go-go-golems/glazed/pkg/cmds/parameters"
- // Define layers
- pls := schema.NewSchema(layers.WithLayers(demo))

### `pkg/doc/topics/commands-reference.md`
Snippets:

- - layers
- The Glazed command system provides a structured approach to building CLI applications that handle multiple output formats, complex parameter validation, and reusable components. This reference covers the complete command system architecture, interfaces, and implementation patterns.
- Building CLI tools typically involves handling parameter parsing, validation, output formatting, and configuration management. Glazed addresses these concerns through a layered architecture that separates command logic from presentation and parameter management.
- The core principle is separation of concerns: commands focus on business logic while Glazed handles parameter parsing, validation, and output formatting. This approach enables automatic support for multiple output formats, consistent parameter handling across commands, and reusable parameter groups.
- │  (name, flags, arguments, layers, etc.)     │

### `pkg/doc/topics/how-to-write-good-documentation-pages.md`
Snippets:

- - **Options/Config:** When documenting flags, parameters, or settings

### `pkg/doc/topics/layers-guide.md`
Snippets:

- Title: Glazed Command Layers Guide
- Slug: layers-guide
- Short: Complete guide to understanding and working with command parameter layers in Glazed
- - layers
- - parameters

### `pkg/doc/topics/logging-layer.md`
Snippets:

- Title: Logging Layer API Reference
- Slug: logging-layer-reference
- # Logging Layer API Reference
- The Glazed logging layer provides comprehensive logging configuration for CLI applications through command-line parameters, environment variables, and configuration files. The layer handles setup for console output, file logging, and centralized log aggregation while supporting multiple output formats and verbosity levels.
- A[CLI Parameters] --> B[Logging Layer]

### `pkg/doc/topics/using-the-query-api.md`
Snippets:

- ├── dsl_bridge.go          # Integration layer
- fmt.Printf("Parameters: %v\n", debugInfo.Parameters)
- http.Error(w, "Missing query parameter", http.StatusBadRequest)

### `pkg/doc/tutorials/01-a-simple-table-cli.md`
Snippets:

- - `ParameterDefinition`: This struct is used to define the parameters (flags or arguments) that the command takes. It
- includes the name of the parameter, the type, and any default value.
- parsedLayers map[string]*layers.ParsedParameterLayer,

### `pkg/doc/tutorials/04-lua.md`
Snippets:

- 2. Creates a global table containing parameter information (`animal_list_params`)
- - Set up parameters for the command
- - Display parameter information
- -- Print parameter information
- print("Layer: " .. layer_name)

### `pkg/doc/tutorials/05-build-first-command.md`
Snippets:

- - Understand command configuration and parameter handling
- Every Glazed command follows a consistent pattern: a command struct embeds `*cmds.CommandDescription` for metadata, and a settings struct maps command-line flags to Go fields using struct tags for type-safe parameter access.
- // Step 2.2: Define settings for type-safe parameter access
- 1. **Command Struct**: `ListUsersCommand` embeds `*cmds.CommandDescription`, which contains command metadata (name, help text, parameters)
- ### Command Configuration and Parameters

### `pkg/doc/tutorials/config-files-quickstart.md`
Snippets:

- This tutorial shows how to load configuration from one or more files using Glazed middlewares. You’ll see a simple single-file setup and a multi-file overlay with deterministic precedence. We’ll also show how to inspect parse steps using `--print-parsed-parameters`.
- - Familiarity with Cobra commands and Glazed layers
- Create a minimal command with a single custom layer and an explicit config file path:
- layers.WithPrefix("demo-"),
- Add `--print-parsed-parameters` to see each config file applied in sequence:

### `pkg/doc/tutorials/custom-layer.md`
Snippets:

- Title: Creating Custom Parameter Layers
- Slug: custom-layer-tutorial
- Short: Step-by-step tutorial for creating reusable custom parameter layers in Glazed
- - layers
- - parameters

### `pkg/doc/tutorials/migrating-from-viper-to-config-files.md`
Snippets:

- The new system replaces Viper's automatic config discovery and merging with explicit file loading middlewares that record each parse step. This makes it clear where each parameter value originated and enables better debugging with `--print-parsed-parameters`.
- ### 2. Config File Format Must Match Layer Structure
- **After:** Config must match layer names and parameters:
- # Layer names as top-level keys
- - Group parameters under layer names

### `pkg/doc/tutorials/migrating-to-facade-packages.md`
Snippets:

- Short: Step-by-step guide to migrate Glazed code from layers/parameters/middlewares vocabulary to the new facade packages (schema/fields/values/sources)
- - `schema` — schema sections (previously “layers”)
- - `fields` — field definitions and field types (previously “parameters”)
- - `values` — resolved values + decoding helpers (previously “parsed layers”)
- - `pkg/cmds/layers.ParameterLayer` → `pkg/cmds/schema.Section`

### `pkg/help/store/README.md`
Snippets:

- - **Compat**: Compatibility layer for existing help system interface
- ## Compatibility Layer
- The predicate system is more powerful than the existing `SectionQuery`, but the compatibility layer ensures existing code continues to work.

### `prompto/glazed/command-description.md`
Snippets:

- - **Layers** (contains parameter definitions, i.e. your command’s flags/arguments)
- Your command’s parameters (both flags and positional arguments) are grouped in a default “layer.” You typically add them via the convenience functions:
- ### 3.1 Defining Parameter Definitions
- Parameters themselves are described by `fields.Definition` from the `glazed/pkg/cmds/parameters` package. For example:
- **Common parameter definition functions**:

### `prompto/glazed/create-application-tutorial.md`
Snippets:

- ### 1.1 Layer Design
- - Each layer should handle one aspect of configuration (e.g., authentication, database, output formatting)
- - Keep parameter definitions focused and cohesive
- - Avoid mixing unrelated parameters in the same layer
- - Use descriptive slugs that indicate the layer's purpose (e.g., "auth", "db", "output")

### `prompto/glazed/create-yaml-command.md`
Snippets:

- Parameters can be defined either as flags (with -- prefix) or positional arguments. Both use the same parameter definition structure, just with different usage patterns. Each entry describes one parameter with fields such as:
- - `name` (required): The parameter name.
- - `type` (required): The parameter type.
- - `required` (optional, boolean): Indicates if this parameter must be supplied.
- - For single numeric parameters, `type` can be `int` or `float`.

### `prompto/glazed/main.md`
Snippets:

- ### Custom Help Layers
- // Implement your custom help layer

## DATA files

### `cmd/examples/config-custom-mapper/config.yaml`
Snippets:

- # Flat config structure - different from the default layer-based structure

### `pinocchio/glazed/create-template-command.yaml`
Snippets:

- The types of parameters that can be used for flags are:

### `pkg/cmds/fields/test-data/gather-fields.yaml`
Snippets:

- expectedError: "unknown parameter type foobar"
- - title: "Test with choice parameters"
- - title: "Test with valid choice parameter"
- - title: "Test with empty choice parameter"
- description: "Ensure that providing an empty string for a choice parameter results in an error."

### `pkg/cmds/sources/tests/middlewares.yaml`
Snippets:

- description: "Empty middlewares should result in empty parsed layers"
- description: "Only set from defaults middlewares, with empty parsed layers"
- - name: "Single parameter, set from defaults"
- description: "Only set from defaults middlewares, with single parameter"
- - name: "Single parameter, set from defaults and then update from map"

### `pkg/cmds/sources/tests/multi-update-from-map.yaml`
Snippets:

- # Test 5: Non-Existent Layers
- - name: "Non-Existent Layers"
- description: "Updates that reference non-existent layers should be ignored."
- - layer2: # This layer does not exist in parameterLayers
- # Test 6: New Layers

### `pkg/cmds/sources/tests/set-from-defaults.yaml`
Snippets:

- - name: "Empty layers and parsedLayers"
- description: "Empty layers should result in empty parsed layers"
- - name: "Single layer with default"
- description: "Single layer with default values should result in a single layer with these values"
- - name: "Single layer with list type default"

### `pkg/cmds/sources/tests/update-from-map-as-default.yaml`
Snippets:

- description: "Confirm that the middleware skips any layers not present in layers_ and does not throw an error."
- layer2:  # This layer is not defined in parameterLayers
- description: "When a parameter has specific choices, verify that defaults not in the choices do not get set and result in an error."
- - name: "Test with multiple layers"
- description: "Verify that the middleware correctly updates defaults across multiple layers without affecting already set parameters in any layer."

### `pkg/cmds/sources/tests/update-from-map.yaml`
Snippets:

- - name: "Update single layer with valid map"
- description: "Updating a single layer with valid values should correctly merge these values"
- - name: "Update non-existent layer"
- description: "Updating a non-existent layer should be ignored and no error should be thrown"
- - name: "Invalid parameter type in update map"

### `pkg/cmds/sources/tests/wrap-with-restricted-layers.yaml`
Snippets:

- - name: "Blacklist Single Layer"
- description: "A single layer is blacklisted and should be removed from ParameterLayers."
- - name: "Blacklist Multiple Layers"
- description: "Multiple layers are blacklisted and should be removed from ParameterLayers."
- - name: "Whitelist single layer"

## OTHER files

### `CHANGELOG`
Snippets:

- - Parse Lua tables into Glazed parameter layers

### `prompto/glazed/definitions`
Snippets:

- prompto get glazed/parameter-types

### `prompto/glazed/parameter-types`
Snippets:

- echo "// Here are the types that can be used to define parameters in glazed:"
- echo "package github.com/go-go-golems/glazed/pkg/cmds/parameters"
- oak go consts pkg/cmds/parameters/parameter-type.go
- oak go definitions pkg/cmds/parameters/file.go --name "FileData" --definition-type struct,interface

### `prompto/glazed/parameters`
Snippets:

- prompto get glazed/parameter-types
- echo "Here are all the types and method signatures for manipulating parameters and parsed parameters in glazed (github.com/go-go-golems/glazed is the base package):"
- echo "package github.com/go-go-golems/glazed/pkg/cmds/parameters"
- oak go definitions --only-public pkg/cmds/parameters/parameters.go
- oak go definitions --only-public pkg/cmds/parameters/parsed-parameter.go

### `prompto/glazed/parameters-verbose`
Snippets:

- glaze help parameter-layers-and-parsed-layers 2>&1
- glaze help parsing-parameters 2>&1
- prompto get glazed/parameters
