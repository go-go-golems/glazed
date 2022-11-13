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
	Columns []FieldName
	Rows    []Row
}

func NewTable() *Table {
	return &Table{
		Columns: []FieldName{},
		Rows:    []Row{},
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
