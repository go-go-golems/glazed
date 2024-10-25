Certainly! I'll create a comprehensive reference document and tutorial for the Lua wrapper layer around Glazed, including type signatures and a section about the conversion functions. I'll reference code blocks using the format you specified.

# Reference Document: Lua Wrapper for Glazed

```go
import lua2 "github.com/go-go-golems/glazed/pkg/lua"
```

## Key Components and Type Signatures

### 1. CallGlazedCommandFromLua

```go
func CallGlazedCommandFromLua(L *lua.LState, cmd cmds.GlazeCommand, luaTable *lua.LTable) (*types.Table, error)
```

Executes a GlazeCommand with parameters from a Lua table.

### 2. CallGlazedBareCommandFromLua

```go
func CallGlazedBareCommandFromLua(L *lua.LState, cmd cmds.BareCommand, luaTable *lua.LTable) error
```

Executes a BareCommand with parameters from a Lua table.

### 3. CallGlazedWriterCommandFromLua

```go
func CallGlazedWriterCommandFromLua(L *lua.LState, cmd cmds.WriterCommand, luaTable *lua.LTable) (string, error)
```

Executes a WriterCommand with parameters from a Lua table.

### 4. RegisterGlazedCommand

```go
func RegisterGlazedCommand(L *lua.LState, cmd interface{})
```

Registers a Glazed command (GlazeCommand, BareCommand, or WriterCommand) in the Lua state.

### 5. ParseNestedLuaTableMiddleware

```go
func ParseNestedLuaTableMiddleware(L *lua.LState, luaTable *lua.LTable) middlewares.Middleware
```

Middleware to parse nested Lua tables into ParsedLayers.

## Conversion Functions

### Low-level Conversion Functions

1. LuaValueToInterface

```go
func LuaValueToInterface(L *lua.LState, value lua.LValue) interface{}
```

Converts a Lua value to a Go interface{}.

2. InterfaceToLuaValue

```go
func InterfaceToLuaValue(L *lua.LState, value interface{}) lua.LValue
```

Converts a Go interface{} to a Lua value.

3. GlazedTableToLuaTable

```go
func GlazedTableToLuaTable(L *lua.LState, glazedTable *types.Table) *lua.LTable
```

Converts a Glazed table to a Lua table.

### Middleware Conversion Functions

1. ParseLuaTableToLayer

```go
func ParseLuaTableToLayer(L *lua.LState, luaTable *lua.LTable, layer layers.ParameterLayer) (*layers.ParsedLayer, error)
```

Parses a Lua table into a ParsedLayer.

2. ParseNestedLuaTableToParsedLayers

```go
func ParseNestedLuaTableToParsedLayers(L *lua.LState, luaTable *lua.LTable, parameterLayers *layers.ParameterLayers) (*layers.ParsedLayers, error)
```

Parses a nested Lua table into ParsedLayers.

3. ParseParameterFromLua

```go
func ParseParameterFromLua(L *lua.LState, value lua.LValue, paramDef *parameters.ParameterDefinition) (interface{}, error)
```

Parses a Lua value into a Go value based on the parameter definition.

4. ParsedLayerToLuaTable

```go
func ParsedLayerToLuaTable(L *lua.LState, parsedLayer *layers.ParsedLayer) *lua.LTable
```

Converts a ParsedLayer to a Lua table.

5. ParsedLayersToLuaTable

```go
func ParsedLayersToLuaTable(L *lua.LState, parsedLayers *layers.ParsedLayers) *lua.LTable
```

Converts ParsedLayers to a nested Lua table.

## Usage

### Registering a Glazed Command

```go
animalListCmd, _ := NewAnimalListCommand()
lua2.RegisterGlazedCommand(L, animalListCmd)
```

This registers the command and creates a global Lua function with the command's name (replacing hyphens with underscores).

### Executing a Glazed Command from Lua

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

### Accessing Command Parameters

The registration process also creates a global Lua table with information about the command's parameters:

```lua
for layer_name, layer_params in pairs(animal_list_params) do
    print("Layer: " .. layer_name)
    for param_name, param_info in pairs(layer_params) do
        print(string.format("  %s (%s): %s", param_name, param_info.type, param_info.description))
        print(string.format("    Default: %s", tostring(param_info.default)))
        print(string.format("    Required: %s", tostring(param_info.required)))
    end
end
```

## Data Conversion

The wrapper handles conversion between Lua and Go types:

- Lua tables are converted to Go slices or maps.
- Lua numbers are converted to Go int or float64.
- Lua strings are converted to Go strings.
- Lua booleans are converted to Go bools.

# Tutorial: Using the Lua Wrapper for Glazed

## Step 1: Setting up the Lua State

First, create a new Lua state:

```go
L := lua.NewState()
defer L.Close()
```

## Step 2: Creating a Glazed Command

Define a Glazed command in Go. For this tutorial, we'll use the AnimalListCommand from the provided example:


```23:54:go-go-labs/cmd/experiments/lua-glazed-cmds/main.go

```


## Step 3: Registering the Command

Register the Glazed command with the Lua state:

```go
animalListCmd, _ := NewAnimalListCommand()
lua2.RegisterGlazedCommand(L, animalListCmd)
```

## Step 4: Preparing Lua Script

Create a Lua script that uses the registered command:


```338:362:go-go-labs/cmd/experiments/lua-glazed-cmds/main.go
	// Run a Lua script that uses the registered command
	script := `
		local params = {
			default = {
				count = 3
			},
			glazed = {
				fields = {"animal", "diet"}
			}
		}
		local result = animal_list(params)
		for i, row in ipairs(result) do
			print(string.format("Animal %d: %s, Diet: %s", i, row.animal, row.diet))
		end

		-- Print parameter information for animal_list command
		print("\nParameters for animal_list command:")
		for layer_name, layer_params in pairs(animal_list_params) do
			print("Layer: " .. layer_name)
			for param_name, param_info in pairs(layer_params) do
				print(string.format("  %s (%s): %s", param_name, param_info.type, param_info.description))
				print(string.format("    Default: %s", tostring(param_info.default)))
				print(string.format("    Required: %s", tostring(param_info.required)))
			end
		end
```


## Step 5: Executing the Lua Script

Execute the Lua script:

```go
if err := L.DoString(script); err != nil {
    fmt.Printf("Error executing Lua script: %v\n", err)
}
```

## Step 6: Handling Results

The results of the Glazed command execution will be printed by the Lua script. You can modify the script to return results or perform further processing as needed.

## Advanced Usage: Nested Lua Tables

The Lua wrapper supports nested Lua tables for complex parameter structures. Here's an example of how to use nested tables:

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

In Go, you can handle these nested tables using the `ParseNestedLuaTableMiddleware`:

```go
middlewares_ := []middlewares.Middleware{
    lua2.ParseNestedLuaTableMiddleware(L, luaTable),
    middlewares.SetFromDefaults(parameters.WithParseStepSource("defaults")),
}

err := middlewares.ExecuteMiddlewares(cmd.Description().Layers, parsedLayers, middlewares_...)
```

This middleware will parse the nested Lua table structure into the appropriate `ParsedLayers` for use with your Glazed command.

By following this tutorial and referencing the provided documentation, you should be able to effectively use the Lua wrapper for Glazed in your projects, allowing for flexible integration of Glazed commands within Lua scripts.