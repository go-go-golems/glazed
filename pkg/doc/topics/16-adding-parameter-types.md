---
Title: Adding New Parameter Types to Glazed
Slug: adding-parameter-types
Short: Comprehensive guide on implementing new parameter types in the Glazed framework.
Topics:
- parameters
- types
- parsing
- validation
- development
- extension
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Adding New Parameter Types to Glazed

## Overview

Parameter types in glazed are defined in the [`glazed/pkg/cmds/parameters`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/parameters) package. Each parameter type requires modifications to several files to handle:

1. Type definition and metadata
2. Parsing logic
3. Validation
4. Value initialization and assignment
5. Cobra CLI integration
6. Viper configuration support
7. Rendering for display

This guide explains how to add a new parameter type to the glazed command line framework. We'll use the example of adding a `credentials` parameter type to demonstrate the process.

## Files to Modify

When adding a new parameter type, you need to modify these files:

### Core Parameter Files
1. [`parameter-type.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/parameters/parameter-type.go) - Define the type constant and metadata methods
2. [`parse.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/parameters/parse.go) - Add parsing logic
3. [`parameters.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/parameters/parameters.go) - Add validation and value assignment
4. [`cobra.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/parameters/cobra.go) - Add CLI flag support
5. [`viper.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/parameters/viper.go) - Add configuration file support
6. [`render.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/parameters/render.go) - Add display formatting

### Additional Files with Exhaustive Switches
7. [`json-schema.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/json-schema.go) - Add JSON schema type mapping
8. [`glazed.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/codegen/glazed.go) - Add Go type mapping for code generation
9. [`lua.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/lua/lua.go) - Add Lua value parsing support

**Note**: The linter will help you find any additional files with exhaustive switch statements that need updating by running `make lint` or `golangci-lint run`.

## Step 1: Define the Parameter Type

In [`parameter-type.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/parameters/parameter-type.go), add your new type constant:

```go
const (
    // ... existing types ...
    ParameterTypeCredentials ParameterType = "credentials"
)
```

Add your type to the relevant metadata methods. For example, if credentials should be treated as a list:

```go
func (p ParameterType) IsList() bool {
    switch p {
    case ParameterTypeCredentials:
        return true
    // ... existing cases ...
    }
}
```

Add other metadata methods as needed:
- `NeedsFileContent()` - if type can load from files
- `NeedsMultipleFileContent()` - if type loads from multiple files
- `IsFile()` - if type represents file data
- `IsObject()` - if type represents structured objects
- `IsKeyValue()` - if type represents key-value maps

## Step 2: Add Parsing Logic

In [`parse.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/parameters/parse.go), add a case to the `ParseParameter` method:

```go
func (p *ParameterDefinition) ParseParameter(v []string, options ...ParseStepOption) (*ParsedParameter, error) {
    // ... existing code ...
    
    switch p.Type {
    // ... existing cases ...
    
    case ParameterTypeCredentials:
        // Parse credentials from command line arguments
        // Example: expect format "username:password" or load from file with @
        if len(v) == 1 && strings.HasPrefix(v[0], "@") {
            // Load from file
            credFile := v[0][1:]
            content, err := os.ReadFile(credFile)
            if err != nil {
                return nil, errors.Wrapf(err, "Could not read credentials file %s", credFile)
            }
            // Parse JSON/YAML credentials file
            var creds map[string]string
            if strings.HasSuffix(credFile, ".json") {
                err = json.Unmarshal(content, &creds)
            } else {
                err = yaml.Unmarshal(content, &creds)
            }
            if err != nil {
                return nil, errors.Wrapf(err, "Could not parse credentials file %s", credFile)
            }
            v_ = creds
        } else {
            // Parse from command line
            creds := make(map[string]string)
            for _, arg := range v {
                parts := strings.SplitN(arg, ":", 2)
                if len(parts) != 2 {
                    return nil, errors.Errorf("Invalid credentials format: %s (expected username:password)", arg)
                }
                creds[parts[0]] = parts[1]
            }
            v_ = creds
        }
    }
    
    // ... rest of method ...
}
```

If your type supports file loading, add cases to `ParseFromReader`:

```go
func (p *ParameterDefinition) ParseFromReader(f io.Reader, filename string, options ...ParseStepOption) (*ParsedParameter, error) {
    // ... existing code ...
    
    switch p.Type {
    // ... existing cases ...
    
    case ParameterTypeCredentials:
        var creds map[string]string
        if strings.HasSuffix(filename, ".json") {
            err = json.NewDecoder(f).Decode(&creds)
        } else {
            err = yaml.NewDecoder(f).Decode(&creds)
        }
        if err != nil {
            return nil, err
        }
        err = ret.Update(creds, options...)
        return ret, err
    }
}
```

## Step 3: Add Validation and Value Assignment

In [`parameters.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/parameters/parameters.go), add validation to `CheckValueValidity`:

```go
func (p *ParameterDefinition) CheckValueValidity(v interface{}) (interface{}, error) {
    // ... existing code ...
    
    switch p.Type {
    // ... existing cases ...
    
    case ParameterTypeCredentials:
        creds, ok := v.(map[string]string)
        if !ok {
            // Try to convert from map[string]interface{}
            credsIface, ok := v.(map[string]interface{})
            if !ok {
                return nil, errors.Errorf("Value for parameter %s is not credentials (expected map[string]string): %v", p.Name, v)
            }
            creds = make(map[string]string)
            for k, v := range credsIface {
                str, ok := v.(string)
                if !ok {
                    return nil, errors.Errorf("Credentials value for key %s is not a string: %v", k, v)
                }
                creds[k] = str
            }
        }
        
        // Validate required fields
        if _, ok := creds["username"]; !ok {
            return nil, errors.Errorf("Credentials missing required 'username' field")
        }
        if _, ok := creds["password"]; !ok {
            return nil, errors.Errorf("Credentials missing required 'password' field")
        }
        
        return creds, nil
    }
}
```

Add empty value initialization to `InitializeValueToEmptyValue`:

```go
func (p *ParameterDefinition) InitializeValueToEmptyValue(value reflect.Value) error {
    switch p.Type {
    // ... existing cases ...
    
    case ParameterTypeCredentials:
        value.Set(reflect.ValueOf(map[string]string{}))
    }
}
```

Add value assignment to `SetValueFromInterface`:

```go
func (p *ParameterDefinition) SetValueFromInterface(value reflect.Value, v interface{}) error {
    // ... validation ...
    
    switch p.Type {
    // ... existing cases ...
    
    case ParameterTypeCredentials:
        creds, ok := v.(map[string]string)
        if !ok {
            return errors.Errorf("expected credentials for parameter %s, got %T", p.Name, v)
        }
        value.Set(reflect.ValueOf(creds))
    }
}
```

## Step 4: Add Cobra CLI Integration

In [`cobra.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/parameters/cobra.go), add flag creation logic:

```go
func (ps *ParameterDefinitions) AddToCobraCommand(cmd *cobra.Command) error {
    // ... existing code ...
    
    switch parameter.Type {
    // ... existing cases ...
    
    case ParameterTypeCredentials:
        defaultValue := []string{}
        if parameter.Default != nil {
            if creds, ok := (*parameter.Default).(map[string]string); ok {
                for k, v := range creds {
                    defaultValue = append(defaultValue, k+":"+v)
                }
            }
        }
        cmd.Flags().StringSliceVarP(&ps.cobraParameterValues[parameter.Name], 
            parameter.Name, parameter.ShortFlag, defaultValue, parameter.Help)
    }
}
```

Add completion logic if needed:

```go
func (ps *ParameterDefinitions) SetupCobraCompletions(cmd *cobra.Command) error {
    // ... existing code ...
    
    switch parameter.Type {
    case ParameterTypeCredentials:
        err = cmd.RegisterFlagCompletionFunc(parameter.Name, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
            return []string{"username:", "token:", "api_key:"}, cobra.ShellCompDirectiveNoSpace
        })
    }
}
```

## Step 6: Add Rendering Support

In [`render.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/parameters/render.go), add display formatting:

```go
func RenderValue(parameterType ParameterType, value interface{}) (string, error) {
    switch parameterType {
    // ... existing cases ...
    
    case ParameterTypeCredentials:
        if creds, ok := value.(map[string]string); ok {
            // Mask sensitive data for display
            rendered := make([]string, 0, len(creds))
            for k, v := range creds {
                if strings.Contains(strings.ToLower(k), "password") || 
                   strings.Contains(strings.ToLower(k), "secret") ||
                   strings.Contains(strings.ToLower(k), "token") {
                    rendered = append(rendered, k+":***")
                } else {
                    rendered = append(rendered, k+":"+v)
                }
            }
            return strings.Join(rendered, ", "), nil
        }
        return "", errors.Errorf("Invalid credentials value: %v", value)
    }
}
```

## Step 7: Update Additional Files with Exhaustive Switches

After implementing the core parameter functionality, you may need to update additional files that have exhaustive switch statements on parameter types:

### JSON Schema Support (`json-schema.go`)
```go
switch param.Type {
// ... existing cases ...
case fields.TypeCredentials:
    prop.Type = "object"
    prop.Properties = map[string]*JSONSchemaProperty{
        "username": {Type: "string"},
        "password": {Type: "string"},
    }
}
```

### Code Generation Support (`codegen/glazed.go`)
```go
func FlagTypeToGoType(s *jen.Statement, parameterType fields.Type) *jen.Statement {
    switch parameterType {
    // ... existing cases ...
    case fields.TypeCredentials:
        return s.Map(jen.Id("string")).Id("string")
    }
}
```

### Lua Integration Support (`lua/lua.go`)
```go
func ParseParameterFromLua(L *lua.LState, value lua.LValue, paramDef *fields.Definition) (interface{}, error) {
    switch paramDef.Type {
    // ... existing cases ...
    case fields.TypeCredentials:
        if table, ok := value.(*lua.LTable); ok {
            creds := make(map[string]string)
            table.ForEach(func(k, v lua.LValue) {
                if keyStr, ok := k.(lua.LString); ok {
                    if valStr, ok := v.(lua.LString); ok {
                        creds[string(keyStr)] = string(valStr)
                    }
                }
            })
            return creds, nil
        }
        return nil, fmt.Errorf("invalid type for credentials parameter '%s': expected table, got %s", paramDef.Name, value.Type())
    }
}
```

**Pro Tip**: Run `make lint` or `golangci-lint run` after adding your parameter type to discover any additional files with exhaustive switches that need updating.

## Step 8: Update Parameter Types Example

Update the parameter types example command in [`cmd/examples/parameter-types/main.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/cmd/examples/parameter-types/main.go) to showcase your new parameter type.

### Add to Parameter Definitions
Add your parameter to the `cmds.WithFlags()` section:

```go
fields.New(
    "credentials-param",
    fields.TypeCredentials,
    fields.WithHelp("A credentials parameter for username/password pairs"),
    fields.WithDefault(map[string]string{"username": "admin", "password": "secret"}),
),
```

### Add to Settings Struct
Add a field to the `ParameterTypesSettings` struct:

```go
type ParameterTypesSettings struct {
    // ... existing fields ...
    CredentialsParam map[string]string `glazed:"credentials-param"`
}
```

### Add to Parameter Data Array
Add an entry to the `parameterData` slice in `RunIntoGlazeProcessor`:

```go
{"credentials-param", fields.TypeCredentials, s.CredentialsParam, "A credentials parameter for username/password pairs", false, nil, map[string]string{"username": "admin", "password": "secret"}},
```

This ensures that developers and users can easily test and understand how your new parameter type works in practice.

## Step 9: Test the Example

After updating the example, test it to ensure your new parameter type works correctly:

```bash
cd cmd/examples/parameter-types
go build -o parameter-types .

# Test with default values
./parameter-types parameter-types

# Test with custom values for your new type
./parameter-types parameter-types --credentials-param username:admin,password:secret

# Test parameter parsing (useful for debugging)
./parameter-types parameter-types --credentials-param username:test,password:demo --print-parsed-parameters
```

Verify that:
- Your parameter appears in the help output (`--help`)
- Default values work correctly
- Custom values parse and display properly
- The rendered value shows what users should see (e.g., secrets are masked)
- The real value contains the actual parsed data

## Step 10: Add Tests

Create comprehensive tests for your new parameter type:

```go
func TestCredentialsParameter(t *testing.T) {
    tests := []struct {
        name     string
        input    []string
        expected map[string]string
        wantErr  bool
    }{
        {
            name:     "single credential",
            input:    []string{"user:pass"},
            expected: map[string]string{"user": "pass"},
        },
        {
            name:     "multiple credentials",
            input:    []string{"user:pass", "api_key:secret"},
            expected: map[string]string{"user": "pass", "api_key": "secret"},
        },
        {
            name:    "invalid format",
            input:   []string{"invalid"},
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            pd := &ParameterDefinition{
                Name: "credentials",
                Type: ParameterTypeCredentials,
            }
            
            result, err := pd.ParseParameter(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result.Value)
        })
    }
}
```

## Example: Complete Credentials Parameter Type

Here's what the complete implementation would look like for a credentials parameter type:

### In parameter-type.go
```go
const (
    // ... existing types ...
    ParameterTypeCredentials ParameterType = "credentials"
)

func (p ParameterType) IsKeyValue() bool {
    switch p {
    case ParameterTypeKeyValue, ParameterTypeCredentials:
        return true
    default:
        return false
    }
}
```

This implementation would:
- Accept credentials in `username:password` format from command line
- Support loading from JSON/YAML files with `@filename` syntax
- Validate that required fields (username, password) are present
- Mask sensitive data when rendering
- Support both single and multiple credential pairs

## Summary

When adding a new parameter type to glazed, you need to modify these core files and follow these steps:

1. **Define the type constant** in `parameter-type.go`
2. **Add parsing logic** in `parse.go` 
3. **Add validation and value assignment** in `parameters.go`
4. **Add CLI flag support** in `cobra.go`
5. **Add configuration file support** in `viper.go`
6. **Add display formatting** in `render.go`
7. **Update exhaustive switches** in additional files
8. **Update the parameter types example** in `cmd/examples/parameter-types/main.go`
9. **Test the example** to verify functionality
10. **Write comprehensive tests**

## Tips and Best Practices

1. **Consistent naming**: Use the pattern `ParameterType<Name>` for constants
2. **Error handling**: Provide clear, descriptive error messages
3. **Security**: Be careful with sensitive data - mask passwords/tokens in displays
4. **Validation**: Validate input format early and provide helpful error messages
5. **File support**: Consider whether your type should support file loading
6. **Testing**: Write comprehensive tests covering all parsing scenarios
7. **Documentation**: Update parameter type documentation and help text
8. **Backwards compatibility**: Ensure new types don't break existing functionality
9. **Update examples**: Always update the parameter types example to showcase new types

## Common Patterns

- **Simple values**: String, int, float, bool - direct value assignment
- **Lists**: StringList, IntegerList - parse multiple values into slices
- **Files**: File types load content and parse based on extension
- **Key-value**: Maps parsed from colon-separated pairs or files
- **Choices**: Validated against predefined options
- **Objects**: Complex structured data loaded from JSON/YAML

Follow these patterns when implementing your custom parameter type to ensure consistency with the rest of the glazed framework.

**Important**: The parameter types example in `cmd/examples/parameter-types/` serves as both documentation and a testing tool. Always update it when adding new parameter types so users and developers can easily understand and test the new functionality.
