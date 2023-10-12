package types

import (
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"reflect"
	"sort"
	"strings"
)

type TableName = string
type FieldName = string
type GenericCellValue = interface{}
type Row = *orderedmap.OrderedMap[FieldName, GenericCellValue]
type MapRowPair = orderedmap.Pair[FieldName, GenericCellValue]

// NewRow creates a new Row instance with the given initial data pairs.
// The order of the pairs is preserved.
func NewRow(initialData ...MapRowPair) Row {
	return orderedmap.New[FieldName, GenericCellValue](
		orderedmap.WithInitialData(initialData...),
	)
}

// NewRowFromMap creates a new Row instance populated with the contents of the provided map.
// The fields in the resulting Row are sorted alphabetically.
func NewRowFromMap(hash map[FieldName]GenericCellValue) Row {
	ret := NewRow()

	// get keys of hash and sorted them
	sortedKeys := []FieldName{}
	for k := range hash {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		ret.Set(k, hash[k])
	}
	return ret
}

// NewRowFromRow creates a new Row instance with a deep copy of the contents from the provided row.
func NewRowFromRow(row Row) Row {
	ret := NewRow()

	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		ret.Set(pair.Key, pair.Value)
	}
	return ret
}

// NewRowFromMapWithColumns creates a new Row instance with values from the provided map,
// but only for the columns specified. If a column doesn't exist in the map, it is skipped.
func NewRowFromMapWithColumns(hash map[FieldName]GenericCellValue, columns []FieldName) Row {
	ret := NewRow()
	for _, column := range columns {
		v, ok := hash[column]
		if !ok {
			continue
		}
		ret.Set(column, v)
	}
	return ret
}

// NewRowFromStruct creates a new Row instance populated with the field values from the provided struct.
// The field names of the struct become the keys in the Row. If lowerCaseKeys is true,
// the field names are converted to lowercase.
func NewRowFromStruct(i interface{}, lowerCaseKeys bool) Row {
	row := NewRow()
	SetFromStruct(row, i, lowerCaseKeys)

	return row
}

// SetFromStruct populates the provided row with values from the given struct.
// The field names of the struct become the keys in the Row. If lowerCaseKeys is true,
// the field names are converted to lowercase. Returns the populated row.
func SetFromStruct(row Row, i interface{}, lowerCaseKeys bool) Row {
	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name := field.Name
		if lowerCaseKeys {
			name = strings.ToLower(name)
		}
		row.Set(name, val.Field(i).Interface())
	}

	return row
}

// MRP creates a new MapRowPair with the given key and value.
// This is useful when creating a Row with NewRow.
func MRP(key FieldName, value GenericCellValue) MapRowPair {
	return orderedmap.Pair[FieldName, GenericCellValue]{Key: key, Value: value}
}

// RowToMap converts a Row to a standard Go map with string keys.
// The keys lose their original order.
func RowToMap(om Row) map[string]interface{} {
	ret := make(map[string]interface{})
	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		ret[pair.Key] = pair.Value
	}
	return ret
}

// GetFields retrieves a list of field names (keys) from the provided Row, in their original order.
func GetFields(om Row) []FieldName {
	ret := []FieldName{}
	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		ret = append(ret, pair.Key)
	}
	return ret
}
