package types

type TableName = string
type FieldName = string
type GenericCellValue = interface{}
type MapRow = map[FieldName]GenericCellValue

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
//
// TODO(manuel, 2023-02-19) This is an ugly ugly method, and really the whole Table/middleware structure needs to be refactored
// See https://github.com/go-go-golems/glazed/issues/146
func (t *Table) Finalize() {
	if t.finalized {
		return
	}
	columnNames := map[FieldName]interface{}{}

	for _, row := range t.Rows {
		for _, field := range row.GetFields() {
			columnNames[field] = nil
		}
	}

	columns := []FieldName{}
	for key := range columnNames {
		columns = append(columns, key)
	}

	t.Columns = columns
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
	for key := range sr.Hash {
		ret = append(ret, key)
	}
	return ret
}

func (sr *SimpleRow) GetValues() MapRow {
	return sr.Hash
}
