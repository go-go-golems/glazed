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

type Table struct {
	Columns   []FieldName
	Rows      []Row
	finalized bool
}

// Finalize is used to "close" a table after processing inputs into it.
// This combines the column names from all the rows with the column names already set to have them all in order.
//
// TODO(manuel, 2023-02-19) This is an ugly ugly method, and really the whole Table/middleware structure needs to be refactored
// See https://github.com/go-go-golems/glazed/issues/146
func (t *Table) Finalize() {
	if t.finalized {
		return
	}

	// create a hash to quickly check if we already have the column
	existingColumns := map[FieldName]interface{}{}
	for _, column := range t.Columns {
		existingColumns[column] = nil
	}

	// WARN(manuel, 2023-06-25) This is really inefficient
	for _, row := range t.Rows {
		for _, field := range row.GetFields() {
			if _, ok := existingColumns[field]; !ok {
				t.Columns = append(t.Columns, field)
				existingColumns[field] = nil
			}
		}
	}

	t.finalized = true
}

func NewTable() *Table {
	return &Table{
		Columns:   []FieldName{},
		Rows:      []Row{},
		finalized: false,
	}
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
