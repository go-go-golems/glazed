package types

import orderedmap "github.com/wk8/go-ordered-map/v2"

type TableName = string
type FieldName = string
type GenericCellValue = interface{}
type MapRow = *orderedmap.OrderedMap[FieldName, GenericCellValue]
type MapRowPair = orderedmap.Pair[FieldName, GenericCellValue]

func NewMapRow(initialData ...MapRowPair) MapRow {
	return orderedmap.New[FieldName, GenericCellValue](
		orderedmap.WithInitialData(initialData...),
	)
}

func NewMapRowFromMap(hash map[FieldName]GenericCellValue) MapRow {
	ret := NewMapRow()
	for k, v := range hash {
		ret.Set(k, v)
	}
	return ret
}

func NewMapRowFromMapWithColumns(hash map[FieldName]GenericCellValue, columns []FieldName) MapRow {
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

type Row interface {
	GetFields() []FieldName
	GetValues() MapRow
}

type SimpleRow struct {
	Hash MapRow
}

func (sr *SimpleRow) GetFields() []FieldName {
	ret := []FieldName{}
	om := sr.Hash
	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		ret = append(ret, pair.Key)
	}
	return ret
}

func (sr *SimpleRow) GetValues() MapRow {
	return sr.Hash
}
