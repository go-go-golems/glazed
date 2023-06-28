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

func NewRow(initialData ...MapRowPair) Row {
	return orderedmap.New[FieldName, GenericCellValue](
		orderedmap.WithInitialData(initialData...),
	)
}

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

func NewRowFromStruct(i interface{}, lowerCaseKeys bool) Row {
	val := reflect.ValueOf(i)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	t := val.Type()
	row := NewRow()
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

func MRP(key FieldName, value GenericCellValue) MapRowPair {
	return orderedmap.Pair[FieldName, GenericCellValue]{Key: key, Value: value}
}

func GetFields(om Row) []FieldName {
	ret := []FieldName{}
	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		ret = append(ret, pair.Key)
	}
	return ret
}
