---
Title: Getting Started with the Glazed Lua Wrapper
Slug: lua-tutorial
Short: A step-by-step tutorial for implementing and using the Glazed Lua wrapper in your applications
Topics:
- lua
- tutorial
- integration
- scripting
Commands:
- RegisterGlazedCommand
- animal_list
Flags:
- count
- fields
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

# Tutorial: Getting Started with the Glazed Lua Wrapper

This tutorial will guide you through the process of integrating and using the Glazed Lua wrapper in your applications. We'll create a simple animal list command and execute it from Lua.

## Prerequisites

- Basic understanding of Go programming
- Basic understanding of Lua
- Glazed library installed
- `github.com/yuin/gopher-lua` package installed

## Step 1: Setting up the Lua State

First, we need to create a new Lua state that will serve as our execution environment:

```go
L := lua.NewState()
defer L.Close()
```

Make sure to always close the Lua state when you're done to prevent memory leaks.

## Step 2: Creating a Glazed Command

Let's create a simple command that we'll use throughout this tutorial. We'll implement an AnimalListCommand that returns information about different animals:

```go
type AnimalListCommand struct {
    // Command implementation
}

func NewAnimalListCommand() (*AnimalListCommand, error) {
    // Command initialization
    return &AnimalListCommand{}, nil
}
```

## Step 3: Registering the Command

Now we'll register our Glazed command with the Lua state:

```go
animalListCmd, _ := NewAnimalListCommand()
lua2.RegisterGlazedCommand(L, animalListCmd)
```

This registration process:
1. Creates a global Lua function named after your command (with hyphens replaced by underscores)
2. Creates a global table containing parameter information (`animal_list_params`)

## Step 4: Creating a Lua Script

Let's create a Lua script that uses our registered command. This script will:
- Set up parameters for the command
- Execute the command
- Print the results
- Display parameter information

```lua
local params = {
    default = {
        count = 3
    },
    glazed = {
        fields = {"animal", "diet"}
    }
}

-- Execute the command
local result = animal_list(params)

-- Print results
for i, row in ipairs(result) do
    print(string.format("Animal %d: %s, Diet: %s", i, row.animal, row.diet))
end

-- Print parameter information
print("\nParameters for animal_list command:")
for layer_name, layer_params in pairs(animal_list_params) do
    print("Layer: " .. layer_name)
    for param_name, param_info in pairs(layer_params) do
        print(string.format("  %s (%s): %s", 
            param_name, 
            param_info.type, 
            param_info.description))
        print(string.format("    Default: %s", 
            tostring(param_info.default)))
        print(string.format("    Required: %s", 
            tostring(param_info.required)))
    end
end
```

## Step 5: Executing the Script

Execute the Lua script in your Go application:

```go
if err := L.DoString(script); err != nil {
    fmt.Printf("Error executing Lua script: %v\n", err)
}
```

## Step 6: Advanced Usage - Working with Nested Tables

For more complex commands, you might need to work with nested parameter structures. Here's how to use nested tables:

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

In Go, handle these nested tables using the `ParseNestedLuaTableMiddleware`:

```go
middlewares_ := []middlewares.Middleware{
    lua2.ParseNestedLuaTableMiddleware(L, luaTable),
    sources.FromDefaults(sources.WithSource("defaults")),
}

err := sources.Execute(cmd.Description().Layers, 
    parsedLayers, 
    middlewares_...)
```


## Conclusion

This tutorial covered the basics of implementing the Glazed Lua wrapper in your applications. You learned how to:
- Set up the Lua environment
- Register Glazed commands
- Create and execute Lua scripts
- Work with parameters and nested tables
- Handle common implementation challenges

For more detailed information about specific features and advanced usage, refer to the main Glazed Lua Wrapper documentation using `glaze help lua`.