---
Title: Using the Lua Wrapper for Glazed
Slug: lua
Short: A comprehensive guide to using Glazed commands within Lua scripts through the Glazed Lua wrapper
Topics:
- lua
- integration
- scripting
Commands:
- RegisterGlazedCommand
- CallGlazedCommandFromLua
- CallGlazedBareCommandFromLua
- CallGlazedWriterCommandFromLua
Flags:
- IsTopLevel
- IsTemplate
- ShowPerDefault
- SectionType
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Using the Lua Wrapper for Glazed

The Glazed Lua wrapper provides an interface for executing Glazed commands within Lua scripts. This guide covers the key components, usage patterns, and data conversion utilities available in the wrapper.

## Key Components

### 1. Command Execution Functions

#### CallGlazedCommandFromLua
```go
func CallGlazedCommandFromLua(L *lua.LState, cmd cmds.GlazeCommand, luaTable *lua.LTable) (*types.Table, error)
```
Executes a GlazeCommand with fields from a Lua table.

#### CallGlazedBareCommandFromLua
```go
func CallGlazedBareCommandFromLua(L *lua.LState, cmd cmds.BareCommand, luaTable *lua.LTable) error
```
Executes a BareCommand with fields from a Lua table.

#### CallGlazedWriterCommandFromLua
```go
func CallGlazedWriterCommandFromLua(L *lua.LState, cmd cmds.WriterCommand, luaTable *lua.LTable) (string, error)
```
Executes a WriterCommand with fields from a Lua table.

### 2. Command Registration

#### RegisterGlazedCommand
```go
func RegisterGlazedCommand(L *lua.LState, cmd interface{})
```
Registers a Glazed command (GlazeCommand, BareCommand, or WriterCommand) in the Lua state.

### 3. Middleware Support

#### ParseNestedLuaTableMiddleware
```go
func ParseNestedLuaTableMiddleware(L *lua.LState, luaTable *lua.LTable) middlewares.Middleware
```
Middleware to parse nested Lua tables into Values.

## Data Conversion Functions

### Low-level Conversion Functions

1. **LuaValueToInterface**
```go
func LuaValueToInterface(L *lua.LState, value lua.LValue) interface{}
```
Converts a Lua value to a Go interface{}.

2. **InterfaceToLuaValue**
```go
func InterfaceToLuaValue(L *lua.LState, value interface{}) lua.LValue
```
Converts a Go interface{} to a Lua value.

3. **GlazedTableToLuaTable**
```go
func GlazedTableToLuaTable(L *lua.LState, glazedTable *types.Table) *lua.LTable
```
Converts a Glazed table to a Lua table.

### Middleware Conversion Functions

1. **ParseLuaTableToSection**
```go
func ParseLuaTableToSection(L *lua.LState, luaTable *lua.LTable, section schema.Section) (*values.SectionValues, error)
```
Parses a Lua table into a SectionValues.

2. **ParseNestedLuaTableToValues**
```go
func ParseNestedLuaTableToValues(L *lua.LState, luaTable *lua.LTable, schema_ *schema.Schema) (*values.Values, error)
```
Parses a nested Lua table into Values.

3. **ParseFieldFromLua**
```go
func ParseFieldFromLua(L *lua.LState, value lua.LValue, paramDef *fields.Definition) (interface{}, error)
```
Parses a Lua value into a Go value based on the field definition.

## Usage Guide

### Setting up the Lua State

First, create a new Lua state:
```go
L := lua.NewState()
defer L.Close()
```

### Registering a Glazed Command

Register your Glazed command with the Lua state:
```go
animalListCmd, _ := NewAnimalListCommand()
lua2.RegisterGlazedCommand(L, animalListCmd)
```

This registers the command and creates:
1. A global Lua function with the command's name (replacing hyphens with underscores)
2. A global table containing field information (`command_name_params`)

### Executing Commands from Lua

Once registered, you can call the command from Lua:
```lua
local params = {
    default = {
        count = 3
    },
    glazed = {
        fields = {"animal", "diet"}
    }
}
local result = animal_list(params)
```

### Accessing Command Fields

The registration process creates a global Lua table with field information:
```lua
for section_name, section_fields in pairs(animal_list_params) do
    print("Section: " .. section_name)
    for field_name, field_info in pairs(section_fields) do
        print(string.format("  %s (%s): %s", field_name, field_info.type, field_info.description))
        print(string.format("    Default: %s", tostring(field_info.default)))
        print(string.format("    Required: %s", tostring(field_info.required)))
    end
end
```

### Advanced Usage: Nested Tables

The wrapper supports nested Lua tables for complex field structures:
```lua
local params = {
    default = {
        count = 3
    },
    glazed = {
        fields = {"animal", "diet"},
        format = "table"
    },
    output = {
        file = "animals.txt"
    }
}
local result = animal_list(params)
```

## Data Type Conversion

The wrapper handles automatic conversion between Lua and Go types:

- Lua tables → Go slices or maps
- Lua numbers → Go int or float64
- Lua strings → Go strings
- Lua booleans → Go bools

## Best Practices

1. Always close the Lua state when done:
```go
L := lua.NewState()
defer L.Close()
```

2. Handle errors appropriately when executing commands:
```go
if err := L.DoString(script); err != nil {
    fmt.Printf("Error executing Lua script: %v\n", err)
}
```

3. Use type assertions when handling return values from Lua functions

4. Structure your field tables to match the expected section organization

5. Leverage the field information tables for runtime validation and documentation
