package lua

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/glazed/pkg/types"
	lua "github.com/yuin/gopher-lua"
)

// ParseNestedLuaTableToValues parses a nested Lua table into Values.
func ParseNestedLuaTableToValues(L *lua.LState, luaTable *lua.LTable, sectionSchema *schema.Schema) (*values.Values, error) {
	parsedValues := values.New()
	var conversionErrors []string

	luaTable.ForEach(func(key, value lua.LValue) {
		if keyStr, ok := key.(lua.LString); ok {
			sectionName := string(keyStr)
			section, ok := sectionSchema.Get(sectionName)
			if !ok {
				conversionErrors = append(conversionErrors, fmt.Sprintf("section '%s' not found", sectionName))
				return
			}

			if nestedTable, ok := value.(*lua.LTable); ok {
				sectionValues, err := ParseLuaTableToSection(L, nestedTable, section)
				if err != nil {
					conversionErrors = append(conversionErrors, err.Error())
				} else {
					parsedValues.Set(sectionName, sectionValues)
				}
			} else {
				conversionErrors = append(conversionErrors, fmt.Sprintf("invalid value for section '%s': expected table, got %s", sectionName, value.Type()))
			}
		}
	})

	if len(conversionErrors) > 0 {
		return nil, fmt.Errorf("field conversion errors: %s", strings.Join(conversionErrors, "; "))
	}

	return parsedValues, nil
}

// ParseLuaTableToSection parses a Lua table into SectionValues.
func ParseLuaTableToSection(L *lua.LState, luaTable *lua.LTable, section schema.Section) (*values.SectionValues, error) {
	fieldValues := make(map[string]interface{})
	var conversionErrors []string

	luaTable.ForEach(func(key, value lua.LValue) {
		if keyStr, ok := key.(lua.LString); ok {
			fieldDef, _ := section.GetDefinitions().Get(string(keyStr))
			if fieldDef != nil {
				convertedValue, err := ParseFieldFromLua(L, value, fieldDef)
				if err != nil {
					conversionErrors = append(conversionErrors, err.Error())
				} else {
					fieldValues[string(keyStr)] = convertedValue
				}
			}
		}
	})

	if len(conversionErrors) > 0 {
		return nil, fmt.Errorf("field conversion errors: %s", strings.Join(conversionErrors, "; "))
	}

	// Parse fields using the section definitions.
	parsedFields, err := section.GetDefinitions().GatherFieldsFromMap(fieldValues, true, fields.WithSource("lua"))
	if err != nil {
		return nil, err
	}

	return values.NewSectionValues(section, values.WithFields(parsedFields))
}

// ParseLuaTableMiddleware parses a Lua table into SectionValues.
func ParseLuaTableMiddleware(L *lua.LState, luaTable *lua.LTable, sectionName string) sources.Middleware {
	return func(next sources.HandlerFunc) sources.HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			section, ok := schema_.Get(sectionName)
			if !ok {
				return fmt.Errorf("section '%s' not found", sectionName)
			}

			parsedSectionValues, err := ParseLuaTableToSection(L, luaTable, section)
			if err != nil {
				return err
			}

			err = parsedValues.GetOrCreate(section).MergeFields(parsedSectionValues)
			if err != nil {
				return err
			}

			return next(schema_, parsedValues)
		}
	}
}

// ParseNestedLuaTableMiddleware parses nested Lua tables into Values.
func ParseNestedLuaTableMiddleware(L *lua.LState, luaTable *lua.LTable) sources.Middleware {
	return func(next sources.HandlerFunc) sources.HandlerFunc {
		return func(schema_ *schema.Schema, parsedValues *values.Values) error {
			newValues, err := ParseNestedLuaTableToValues(L, luaTable, schema_)
			if err != nil {
				return err
			}

			// Merge the new parsed values with the existing ones.
			err = parsedValues.Merge(newValues)
			if err != nil {
				return err
			}

			return next(schema_, parsedValues)
		}
	}
}

// ParseFieldFromLua parses a Lua value into a Go value based on a field definition.
func ParseFieldFromLua(L *lua.LState, value lua.LValue, fieldDef *fields.Definition) (interface{}, error) {
	switch fieldDef.Type {
	case fields.TypeString, fields.TypeSecret:
		if v, ok := value.(lua.LString); ok {
			return string(v), nil
		}
		return nil, fmt.Errorf("invalid type for field '%s': expected string, got %s", fieldDef.Name, value.Type())
	case fields.TypeInteger:
		if v, ok := value.(lua.LNumber); ok {
			return int(v), nil
		}
		return nil, fmt.Errorf("invalid type for field '%s': expected integer, got %s", fieldDef.Name, value.Type())
	case fields.TypeFloat:
		if v, ok := value.(lua.LNumber); ok {
			return float64(v), nil
		}
		return nil, fmt.Errorf("invalid type for field '%s': expected float, got %s", fieldDef.Name, value.Type())
	case fields.TypeBool:
		if v, ok := value.(lua.LBool); ok {
			return bool(v), nil
		}
		return nil, fmt.Errorf("invalid type for field '%s': expected boolean, got %s", fieldDef.Name, value.Type())
	case fields.TypeStringList:
		if tbl, ok := value.(*lua.LTable); ok {
			var list []string
			var invalidTypes []string
			tbl.ForEach(func(_, v lua.LValue) {
				if str, ok := v.(lua.LString); ok {
					list = append(list, string(str))
				} else {
					invalidTypes = append(invalidTypes, v.Type().String())
				}
			})
			if len(invalidTypes) > 0 {
				return nil, fmt.Errorf("invalid types in string list for field '%s': %v", fieldDef.Name, invalidTypes)
			}
			return list, nil
		}
		return nil, fmt.Errorf("invalid type for field '%s': expected table (string list), got %s", fieldDef.Name, value.Type())
	case fields.TypeIntegerList:
		if tbl, ok := value.(*lua.LTable); ok {
			var list []int
			var invalidTypes []string
			tbl.ForEach(func(_, v lua.LValue) {
				if num, ok := v.(lua.LNumber); ok {
					list = append(list, int(num))
				} else {
					invalidTypes = append(invalidTypes, v.Type().String())
				}
			})
			if len(invalidTypes) > 0 {
				return nil, fmt.Errorf("invalid types in integer list for field '%s': %v", fieldDef.Name, invalidTypes)
			}
			return list, nil
		}
		return nil, fmt.Errorf("invalid type for field '%s': expected table (integer list), got %s", fieldDef.Name, value.Type())
	case fields.TypeFloatList:
		if tbl, ok := value.(*lua.LTable); ok {
			var list []float64
			var invalidTypes []string
			tbl.ForEach(func(_, v lua.LValue) {
				if num, ok := v.(lua.LNumber); ok {
					list = append(list, float64(num))
				} else {
					invalidTypes = append(invalidTypes, v.Type().String())
				}
			})
			if len(invalidTypes) > 0 {
				return nil, fmt.Errorf("invalid types in float list for field '%s': %v", fieldDef.Name, invalidTypes)
			}
			return list, nil
		}
		return nil, fmt.Errorf("invalid type for field '%s': expected table (float list), got %s", fieldDef.Name, value.Type())
	case fields.TypeChoice:
		if v, ok := value.(lua.LString); ok {
			choice := string(v)
			for _, c := range fieldDef.Choices {
				if c == choice {
					return choice, nil
				}
			}
			return nil, fmt.Errorf("invalid choice '%s' for field '%s'", choice, fieldDef.Name)
		}
		return nil, fmt.Errorf("invalid type for field '%s': expected string (choice), got %s", fieldDef.Name, value.Type())
	case fields.TypeChoiceList:
		if tbl, ok := value.(*lua.LTable); ok {
			var choices []string
			var invalidChoices []string
			var invalidTypes []string
			tbl.ForEach(func(_, v lua.LValue) {
				if str, ok := v.(lua.LString); ok {
					choice := string(str)
					valid := false
					for _, c := range fieldDef.Choices {
						if c == choice {
							choices = append(choices, choice)
							valid = true
							break
						}
					}
					if !valid {
						invalidChoices = append(invalidChoices, choice)
					}
				} else {
					invalidTypes = append(invalidTypes, v.Type().String())
				}
			})
			if len(invalidTypes) > 0 {
				return nil, fmt.Errorf("invalid types in choice list for field '%s': %v", fieldDef.Name, invalidTypes)
			}
			if len(invalidChoices) > 0 {
				return nil, fmt.Errorf("invalid choices %v for field '%s'", invalidChoices, fieldDef.Name)
			}
			return choices, nil
		}
		return nil, fmt.Errorf("invalid type for field '%s': expected table (choice list), got %s", fieldDef.Name, value.Type())
	case fields.TypeDate:
		if v, ok := value.(lua.LString); ok {
			parsedDate, err := fields.ParseDate(string(v))
			if err == nil {
				return parsedDate, nil
			}
			return nil, fmt.Errorf("invalid date '%s' for field '%s': %v", v, fieldDef.Name, err)
		}
		return nil, fmt.Errorf("invalid type for field '%s': expected string (date), got %s", fieldDef.Name, value.Type())
	case fields.TypeKeyValue:
		if tbl, ok := value.(*lua.LTable); ok {
			keyValue := make(map[string]interface{})
			tbl.ForEach(func(k, v lua.LValue) {
				if key, ok := k.(lua.LString); ok {
					keyValue[string(key)] = LuaValueToInterface(L, v)
				}
			})
			return keyValue, nil
		}
		return nil, fmt.Errorf("invalid type for field '%s': expected table (key-value), got %s", fieldDef.Name, value.Type())
	case fields.TypeFile,
		fields.TypeFileList,
		fields.TypeObjectListFromFile,
		fields.TypeObjectListFromFiles,
		fields.TypeObjectFromFile,
		fields.TypeStringFromFile,
		fields.TypeStringFromFiles,
		fields.TypeStringListFromFile,
		fields.TypeStringListFromFiles:
		return nil, fmt.Errorf("field type '%s' for '%s' is not implemented for Lua conversion", fieldDef.Type, fieldDef.Name)
	}
	return nil, fmt.Errorf("unsupported field type '%s' for '%s'", fieldDef.Type, fieldDef.Name)
}

// LuaValueToInterface converts a Lua value to a Go interface{}

func LuaValueToInterface(L *lua.LState, value lua.LValue) interface{} {
	switch v := value.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		maxn := v.MaxN()
		if maxn == 0 { // Table is a map
			result := make(map[string]interface{})
			v.ForEach(func(key, value lua.LValue) {
				result[key.String()] = LuaValueToInterface(L, value)
			})
			return result
		} else { // Table is an array
			result := make([]interface{}, 0, maxn)
			for i := 1; i <= maxn; i++ {
				result = append(result, LuaValueToInterface(L, v.RawGetInt(i)))
			}
			return result
		}
	default:
		return v.String()
	}
}

// GlazedTableToLuaTable converts a Glazed table to a Lua table
func GlazedTableToLuaTable(L *lua.LState, glazedTable *types.Table) *lua.LTable {
	luaTable := L.CreateTable(len(glazedTable.Rows), 0)

	for i, row := range glazedTable.Rows {
		rowTable := L.CreateTable(0, len(glazedTable.Columns))
		for _, col := range glazedTable.Columns {
			value, ok := row.Get(col)
			if !ok {
				continue
			}
			rowTable.RawSetString(col, InterfaceToLuaValue(L, value))
		}
		luaTable.RawSetInt(i+1, rowTable)
	}

	return luaTable
}

// InterfaceToLuaValue converts a Go interface{} to a Lua value
func InterfaceToLuaValue(L *lua.LState, value interface{}) lua.LValue {
	// Dereference pointers
	if reflect.ValueOf(value).Kind() == reflect.Ptr {
		if reflect.ValueOf(value).IsNil() {
			return lua.LNil
		}
		return InterfaceToLuaValue(L, reflect.ValueOf(value).Elem().Interface())
	}

	// Unwrap interfaces
	if v := reflect.ValueOf(value); v.Kind() == reflect.Interface && !v.IsNil() {
		return InterfaceToLuaValue(L, v.Elem().Interface())
	}

	switch v := value.(type) {
	case nil:
		return lua.LNil
	case bool:
		return lua.LBool(v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		if n, ok := cast.CastNumberInterfaceToFloat[float64](v); ok {
			return lua.LNumber(n)
		}
	case float32, float64:
		if n, ok := cast.CastFloatInterfaceToFloat[float64](v); ok {
			return lua.LNumber(n)
		}
	case string:
		return lua.LString(v)
	case []byte:
		return lua.LString(string(v))
	case time.Time:
		return lua.LString(v.Format(time.RFC3339))
	case []interface{}:
		table := L.CreateTable(len(v), 0)
		for i, item := range v {
			table.RawSetInt(i+1, InterfaceToLuaValue(L, item))
		}
		return table
	case map[string]interface{}:
		table := L.CreateTable(0, len(v))
		for key, item := range v {
			table.RawSetString(key, InterfaceToLuaValue(L, item))
		}
		return table
	}

	// Handle slices and arrays
	if reflect.TypeOf(value).Kind() == reflect.Slice || reflect.TypeOf(value).Kind() == reflect.Array {
		v := reflect.ValueOf(value)
		table := L.CreateTable(v.Len(), 0)
		for i := 0; i < v.Len(); i++ {
			table.RawSetInt(i+1, InterfaceToLuaValue(L, v.Index(i).Interface()))
		}
		return table
	}

	// Handle maps
	if reflect.TypeOf(value).Kind() == reflect.Map {
		v := reflect.ValueOf(value)
		table := L.CreateTable(0, v.Len())
		for _, key := range v.MapKeys() {
			table.RawSet(InterfaceToLuaValue(L, key.Interface()), InterfaceToLuaValue(L, v.MapIndex(key).Interface()))
		}
		return table
	}

	// Handle structs
	if reflect.TypeOf(value).Kind() == reflect.Struct {
		v := reflect.ValueOf(value)
		t := v.Type()
		table := L.CreateTable(0, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.IsExported() {
				table.RawSetString(field.Name, InterfaceToLuaValue(L, v.Field(i).Interface()))
			}
		}
		return table
	}

	// Default: convert to string
	return lua.LString(fmt.Sprintf("%v", value))
}

// SectionValuesToLuaTable converts SectionValues to a Lua table.
func SectionValuesToLuaTable(L *lua.LState, sectionValues *values.SectionValues) *lua.LTable {
	luaTable := L.CreateTable(0, len(sectionValues.Fields.ToMap()))

	sectionValues.Fields.ForEach(func(name string, fieldValue *fields.FieldValue) {
		luaTable.RawSetString(name, InterfaceToLuaValue(L, fieldValue.Value))
	})

	return luaTable
}

// ValuesToLuaTable converts Values to a nested Lua table.
func ValuesToLuaTable(L *lua.LState, parsedValues *values.Values) *lua.LTable {
	luaTable := L.CreateTable(0, parsedValues.Len())

	parsedValues.ForEach(func(sectionName string, sectionValues *values.SectionValues) {
		sectionTable := SectionValuesToLuaTable(L, sectionValues)
		luaTable.RawSetString(sectionName, sectionTable)
	})

	return luaTable
}
