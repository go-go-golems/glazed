---
Title: Adding New Field Types to Glazed
Slug: adding-field-types
Short: Comprehensive guide on implementing new field types in the Glazed framework.
Topics:
- fields
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

# Adding New Field Types to Glazed

## Overview

Field types in glazed are defined in the [`glazed/pkg/cmds/fields`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/fields) package. Each field type requires modifications to several files to handle:

1. Type definition and metadata
2. Parsing logic
3. Validation
4. Value initialization and assignment
5. Cobra CLI integration
6. Viper configuration support
7. Rendering for display

This guide explains how to add a new field type to the glazed command line framework. We'll use the example of adding a `credentials` field type to demonstrate the process.

## Files to Modify

When adding a new field type, you need to modify these files:

### Core Field Files
1. [`field-type.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/fields/field-type.go) - Define the type constant and metadata methods
2. [`parse.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/fields/parse.go) - Add parsing logic
3. [`fields.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/fields/fields.go) - Add validation and value assignment
4. [`cobra.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/fields/cobra.go) - Add CLI flag support
5. [`viper.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/fields/viper.go) - Add configuration file support
6. [`render.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/fields/render.go) - Add display formatting

### Additional Files with Exhaustive Switches
7. [`json-schema.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/json-schema.go) - Add JSON schema type mapping
8. [`glazed.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/codegen/glazed.go) - Add Go type mapping for code generation
9. [`lua.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/lua/lua.go) - Add Lua value parsing support

**Note**: The linter will help you find any additional files with exhaustive switch statements that need updating by running `make lint` or `golangci-lint run`.

## Step 1: Define the Field Type

In [`field-type.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/fields/field-type.go), add your new type constant:

```go
const (
    // ... existing types ...
    TypeCredentials Type = "credentials"
)
```

Add your type to the relevant metadata methods. For example, if credentials should be treated as a list:

```go
func (p Type) IsList() bool {
    switch p {
    case TypeCredentials:
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

In [`parse.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/fields/parse.go), add a case to the `ParseField` method:

```go
func (p *Definition) ParseField(v []string, options ...ParseStepOption) (*FieldValue, error) {
    // ... existing code ...
    
    switch p.Type {
    // ... existing cases ...
    
    case TypeCredentials:
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
func (p *Definition) ParseFromReader(f io.Reader, filename string, options ...ParseStepOption) (*FieldValue, error) {
    // ... existing code ...
    
    switch p.Type {
    // ... existing cases ...
    
    case TypeCredentials:
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

In [`fields.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/fields/fields.go), add validation to `CheckValueValidity`:

```go
func (p *Definition) CheckValueValidity(v interface{}) (interface{}, error) {
    // ... existing code ...
    
    switch p.Type {
    // ... existing cases ...
    
    case TypeCredentials:
        creds, ok := v.(map[string]string)
        if !ok {
            // Try to convert from map[string]interface{}
            credsIface, ok := v.(map[string]interface{})
            if !ok {
                return nil, errors.Errorf("Value for field %s is not credentials (expected map[string]string): %v", p.Name, v)
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
func (p *Definition) InitializeValueToEmptyValue(value reflect.Value) error {
    switch p.Type {
    // ... existing cases ...
    
    case TypeCredentials:
        value.Set(reflect.ValueOf(map[string]string{}))
    }
}
```

Add value assignment to `SetValueFromInterface`:

```go
func (p *Definition) SetValueFromInterface(value reflect.Value, v interface{}) error {
    // ... validation ...
    
    switch p.Type {
    // ... existing cases ...
    
    case TypeCredentials:
        creds, ok := v.(map[string]string)
        if !ok {
            return errors.Errorf("expected credentials for field %s, got %T", p.Name, v)
        }
        value.Set(reflect.ValueOf(creds))
    }
}
```

## Step 4: Add Cobra CLI Integration

In [`cobra.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/fields/cobra.go), add flag creation logic:

```go
func (ps *Definitions) AddToCobraCommand(cmd *cobra.Command) error {
    // ... existing code ...
    
    switch field.Type {
    // ... existing cases ...
    
    case TypeCredentials:
        defaultValue := []string{}
        if field.Default != nil {
            if creds, ok := (*field.Default).(map[string]string); ok {
                for k, v := range creds {
                    defaultValue = append(defaultValue, k+":"+v)
                }
            }
        }
        cmd.Flags().StringSliceVarP(&ps.cobraFieldValues[field.Name], 
            field.Name, field.ShortFlag, defaultValue, field.Help)
    }
}
```

Add completion logic if needed:

```go
func (ps *Definitions) SetupCobraCompletions(cmd *cobra.Command) error {
    // ... existing code ...
    
    switch field.Type {
    case TypeCredentials:
        err = cmd.RegisterFlagCompletionFunc(field.Name, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
            return []string{"username:", "token:", "api_key:"}, cobra.ShellCompDirectiveNoSpace
        })
    }
}
```

## Step 6: Add Rendering Support

In [`render.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/pkg/cmds/fields/render.go), add display formatting:

```go
func RenderValue(fieldType Type, value interface{}) (string, error) {
    switch fieldType {
    // ... existing cases ...
    
    case TypeCredentials:
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

After implementing the core field functionality, you may need to update additional files that have exhaustive switch statements on field types:

### JSON Schema Support (`json-schema.go`)
```go
switch field.Type {
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
func FlagTypeToGoType(s *jen.Statement, fieldType fields.Type) *jen.Statement {
    switch fieldType {
    // ... existing cases ...
    case fields.TypeCredentials:
        return s.Map(jen.Id("string")).Id("string")
    }
}
```

### Lua Integration Support (`lua/lua.go`)
```go
func ParseFieldFromLua(L *lua.LState, value lua.LValue, fieldDef *fields.Definition) (interface{}, error) {
    switch fieldDef.Type {
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
        return nil, fmt.Errorf("invalid type for credentials field '%s': expected table, got %s", fieldDef.Name, value.Type())
    }
}
```

**Pro Tip**: Run `make lint` or `golangci-lint run` after adding your field type to discover any additional files with exhaustive switches that need updating.

## Step 8: Update Field Types Example

Update the field types example command in [`cmd/examples/field-types/main.go`](file:///home/manuel/workspaces/2025-06-09/add-geppetto-js/glazed/cmd/examples/field-types/main.go) to showcase your new field type.

### Add to Field Definitions
Add your field to the `cmds.WithFlags()` section:

```go
fields.New(
    "credentials-field",
    fields.TypeCredentials,
    fields.WithHelp("A credentials field for username/password pairs"),
    fields.WithDefault(map[string]string{"username": "admin", "password": "secret"}),
),
```

### Add to Settings Struct
Add a field to the `TypesSettings` struct:

```go
type TypesSettings struct {
    // ... existing fields ...
    CredentialsField map[string]string `glazed:"credentials-field"`
}
```

### Add to Field Data Array
Add an entry to the `fieldData` slice in `RunIntoGlazeProcessor`:

```go
{"credentials-field", fields.TypeCredentials, s.CredentialsField, "A credentials field for username/password pairs", false, nil, map[string]string{"username": "admin", "password": "secret"}},
```

This ensures that developers and users can easily test and understand how your new field type works in practice.

## Step 9: Test the Example

After updating the example, test it to ensure your new field type works correctly:

```bash
cd cmd/examples/field-types
go build -o field-types .

# Test with default values
./field-types field-types

# Test with custom values for your new type
./field-types field-types --credentials-field username:admin,password:secret

# Test field parsing (useful for debugging)
./field-types field-types --credentials-field username:test,password:demo --print-parsed-fields
```

Verify that:
- Your field appears in the help output (`--help`)
- Default values work correctly
- Custom values parse and display properly
- The rendered value shows what users should see (e.g., secrets are masked)
- The real value contains the actual parsed data

## Step 10: Add Tests

Create comprehensive tests for your new field type:

```go
func TestCredentialsField(t *testing.T) {
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
            pd := &Definition{
                Name: "credentials",
                Type: TypeCredentials,
            }
            
            result, err := pd.ParseField(tt.input)
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

## Example: Complete Credentials Field Type

Here's what the complete implementation would look like for a credentials field type:

### In field-type.go
```go
const (
    // ... existing types ...
    TypeCredentials Type = "credentials"
)

func (p Type) IsKeyValue() bool {
    switch p {
    case TypeKeyValue, TypeCredentials:
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

When adding a new field type to glazed, you need to modify these core files and follow these steps:

1. **Define the type constant** in `field-type.go`
2. **Add parsing logic** in `parse.go` 
3. **Add validation and value assignment** in `fields.go`
4. **Add CLI flag support** in `cobra.go`
5. **Add configuration file support** in `viper.go`
6. **Add display formatting** in `render.go`
7. **Update exhaustive switches** in additional files
8. **Update the field types example** in `cmd/examples/field-types/main.go`
9. **Test the example** to verify functionality
10. **Write comprehensive tests**

## Tips and Best Practices

1. **Consistent naming**: Use the pattern `Type<Name>` for constants
2. **Error handling**: Provide clear, descriptive error messages
3. **Security**: Be careful with sensitive data - mask passwords/tokens in displays
4. **Validation**: Validate input format early and provide helpful error messages
5. **File support**: Consider whether your type should support file loading
6. **Testing**: Write comprehensive tests covering all parsing scenarios
7. **Documentation**: Update field type documentation and help text
8. **Backwards compatibility**: Ensure new types don't break existing functionality
9. **Update examples**: Always update the field types example to showcase new types

## Common Patterns

- **Simple values**: String, int, float, bool - direct value assignment
- **Lists**: StringList, IntegerList - parse multiple values into slices
- **Files**: File types load content and parse based on extension
- **Key-value**: Maps parsed from colon-separated pairs or files
- **Choices**: Validated against predefined options
- **Objects**: Complex structured data loaded from JSON/YAML

Follow these patterns when implementing your custom field type to ensure consistency with the rest of the glazed framework.

**Important**: The field types example in `cmd/examples/field-types/` serves as both documentation and a testing tool. Always update it when adding new field types so users and developers can easily understand and test the new functionality.
