package lua

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	middlewares2 "github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	lua2 "github.com/yuin/gopher-lua"
)

// CallGlazedCommandFromLua executes a GlazeCommand with parameters from a Lua table
func CallGlazedCommandFromLua(L *lua2.LState, cmd cmds.GlazeCommand, luaTable *lua2.LTable) (*types.Table, error) {
	// Create parsed layers
	parsedLayers := values.New()

	// Define middlewares
	middlewares_ := []middlewares.Middleware{
		// Parse from Lua table (highest priority)
		ParseNestedLuaTableMiddleware(L, luaTable),
		// Set defaults (lowest priority)
		sources.FromDefaults(sources.WithSource(sources.SourceDefaults)),
	}

	// Execute middlewares
	err := sources.Execute(cmd.Description().Layers, parsedLayers, middlewares_...)
	if err != nil {
		return nil, fmt.Errorf("error executing middlewares: %v", err)
	}

	glazedLayer, ok := parsedLayers.Get(settings.GlazedSlug)
	if !ok {
		return nil, fmt.Errorf("glazed layer not found")
	}
	gp, err := settings.SetupTableProcessor(glazedLayer, middlewares2.WithTableMiddleware(&table.NullTableMiddleware{}))
	if err != nil {
		return nil, fmt.Errorf("error setting up table processor: %v", err)
	}

	ctx := context.Background()

	// Run the command with the parsed layers
	err = cmd.RunIntoGlazeProcessor(ctx, parsedLayers, gp)
	if err != nil {
		return nil, fmt.Errorf("error running command: %v", err)
	}
	err = gp.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("error closing processor: %v", err)
	}

	return gp.Table, nil
}

// LuaCallGlazedCommand is a Lua-callable wrapper for CallGlazedCommandFromLua
func LuaCallGlazedCommand(L *lua2.LState) int {
	// Get the GlazeCommand from the first argument (userdata)
	cmdUD := L.CheckUserData(1)
	cmd, ok := cmdUD.Value.(cmds.GlazeCommand)
	if !ok {
		L.ArgError(1, "GlazeCommand expected")
		return 0
	}

	// Get the Lua table from the second argument
	luaTable := L.CheckTable(2)

	// Call the Go function
	result, err := CallGlazedCommandFromLua(L, cmd, luaTable)
	if err != nil {
		L.Push(lua2.LNil)
		L.Push(lua2.LString(err.Error()))
		return 2
	}

	// Convert the result to a Lua table
	luaResult := GlazedTableToLuaTable(L, result)
	L.Push(luaResult)
	return 1
}

// CallGlazedBareCommandFromLua executes a BareCcommand with parameters from a Lua table
func CallGlazedBareCommandFromLua(L *lua2.LState, cmd cmds.BareCommand, luaTable *lua2.LTable) error {
	parsedLayers := values.New()

	middlewares_ := []middlewares.Middleware{
		ParseNestedLuaTableMiddleware(L, luaTable),
		sources.FromDefaults(sources.WithSource(sources.SourceDefaults)),
	}

	err := sources.Execute(cmd.Description().Layers, parsedLayers, middlewares_...)
	if err != nil {
		return fmt.Errorf("error executing middlewares: %v", err)
	}

	ctx := context.Background()

	// Run the command with the parsed layers
	err = cmd.Run(ctx, parsedLayers)
	if err != nil {
		return fmt.Errorf("error running command: %v", err)
	}

	return nil
}

// CallGlazedWriterCommandFromLua executes a WriterCommand with parameters from a Lua table
func CallGlazedWriterCommandFromLua(L *lua2.LState, cmd cmds.WriterCommand, luaTable *lua2.LTable) (string, error) {
	parsedLayers := values.New()

	middlewares_ := []middlewares.Middleware{
		ParseNestedLuaTableMiddleware(L, luaTable),
		sources.FromDefaults(sources.WithSource(sources.SourceDefaults)),
	}

	err := sources.Execute(cmd.Description().Layers, parsedLayers, middlewares_...)
	if err != nil {
		return "", fmt.Errorf("error executing middlewares: %v", err)
	}

	ctx := context.Background()

	// Create a buffer to capture the output
	var buf bytes.Buffer

	// Run the command with the parsed layers
	err = cmd.RunIntoWriter(ctx, parsedLayers, &buf)
	if err != nil {
		return "", fmt.Errorf("error running command: %v", err)
	}

	return buf.String(), nil
}

// LuaCallGlazedBareCommand is a Lua-callable wrapper for CallGlazedBareCommandFromLua
func LuaCallGlazedBareCommand(L *lua2.LState) int {
	cmdUD := L.CheckUserData(1)
	cmd, ok := cmdUD.Value.(cmds.BareCommand)
	if !ok {
		L.ArgError(1, "BareCommand expected")
		return 0
	}

	luaTable := L.CheckTable(2)

	err := CallGlazedBareCommandFromLua(L, cmd, luaTable)
	if err != nil {
		L.Push(lua2.LBool(false))
		L.Push(lua2.LString(err.Error()))
		return 2
	}

	L.Push(lua2.LBool(true))
	return 1
}

// LuaCallGlazedWriterCommand is a Lua-callable wrapper for CallGlazedWriterCommandFromLua
func LuaCallGlazedWriterCommand(L *lua2.LState) int {
	cmdUD := L.CheckUserData(1)
	cmd, ok := cmdUD.Value.(cmds.WriterCommand)
	if !ok {
		L.ArgError(1, "WriterCommand expected")
		return 0
	}

	luaTable := L.CheckTable(2)

	result, err := CallGlazedWriterCommandFromLua(L, cmd, luaTable)
	if err != nil {
		L.Push(lua2.LNil)
		L.Push(lua2.LString(err.Error()))
		return 2
	}

	L.Push(lua2.LString(result))
	return 1
}

// RegisterGlazedCommand registers a GlazeCommand, BareCommand, or WriterCommand in the Lua state
func RegisterGlazedCommand(L *lua2.LState, cmd interface{}) {
	var name string
	var fn *lua2.LFunction

	switch c := cmd.(type) {
	case cmds.GlazeCommand:
		name = c.Description().Name
		fn = L.NewFunction(func(L *lua2.LState) int {
			luaTable := L.CheckTable(1)
			result, err := CallGlazedCommandFromLua(L, c, luaTable)
			if err != nil {
				L.Push(lua2.LNil)
				L.Push(lua2.LString(err.Error()))
				return 2
			}
			luaResult := GlazedTableToLuaTable(L, result)
			L.Push(luaResult)
			return 1
		})
	case cmds.BareCommand:
		name = c.Description().Name
		fn = L.NewFunction(func(L *lua2.LState) int {
			luaTable := L.CheckTable(1)
			err := CallGlazedBareCommandFromLua(L, c, luaTable)
			if err != nil {
				L.Push(lua2.LBool(false))
				L.Push(lua2.LString(err.Error()))
				return 2
			}
			L.Push(lua2.LBool(true))
			return 1
		})
	case cmds.WriterCommand:
		name = c.Description().Name
		fn = L.NewFunction(func(L *lua2.LState) int {
			luaTable := L.CheckTable(1)
			result, err := CallGlazedWriterCommandFromLua(L, c, luaTable)
			if err != nil {
				L.Push(lua2.LNil)
				L.Push(lua2.LString(err.Error()))
				return 2
			}
			L.Push(lua2.LString(result))
			return 1
		})
	default:
		panic(fmt.Sprintf("Unsupported command type: %T", cmd))
	}

	// Convert command name to a valid Lua identifier
	luaName := strings.ReplaceAll(name, "-", "_")

	// Register the function in the global environment
	L.SetGlobal(luaName, fn)

	// Update the parameter information global name
	paramsGlobalName := luaName + "_params"

	// Get the command description
	desc := cmd.(cmds.Command).Description()

	// Create a table to hold all layers and their parameters
	layersTable := L.CreateTable(0, desc.Layers.Len())

	// Iterate through all layers
	desc.Layers.ForEach(func(layerName string, layer schema.Section) {
		layerTable := L.CreateTable(0, layer.GetParameterDefinitions().Len())

		// Add parameters for this layer
		layer.GetParameterDefinitions().ForEach(func(param *fields.Definition) {
			paramInfo := L.CreateTable(0, 5)
			paramInfo.RawSetString("name", lua2.LString(param.Name))
			paramInfo.RawSetString("type", lua2.LString(string(param.Type)))
			paramInfo.RawSetString("description", lua2.LString(param.Help))
			defaultValue := InterfaceToLuaValue(L, param.Default)
			paramInfo.RawSetString("default", defaultValue)
			paramInfo.RawSetString("required", lua2.LBool(param.Required))
			layerTable.RawSetString(param.Name, paramInfo)
		})

		layersTable.RawSetString(layerName, layerTable)
	})

	// Set the global variable with the layers table
	L.SetGlobal(paramsGlobalName, layersTable)
}
