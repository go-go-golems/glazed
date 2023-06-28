package types

import orderedmap "github.com/wk8/go-ordered-map/v2"

type TableName = string
type FieldName = string
type GenericCellValue = interface{}
type Row = *orderedmap.OrderedMap[FieldName, GenericCellValue]
type MapRowPair = orderedmap.Pair[FieldName, GenericCellValue]

func NewMapRow(initialData ...MapRowPair) Row {
	return orderedmap.New[FieldName, GenericCellValue](
		orderedmap.WithInitialData(initialData...),
	)
}

func NewMapRowFromMap(hash map[FieldName]GenericCellValue) Row {
	ret := NewMapRow()
	for k, v := range hash {
		ret.Set(k, v)
	}
	return ret
}

func NewMapRowFromMapWithColumns(hash map[FieldName]GenericCellValue, columns []FieldName) Row {
	ret := NewMapRow()
	for _, column := range columns {
		v, ok := hash[column]
		if !ok {
			continue
		}
		ret.Set(column, v)
	}
	return ret
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
