package types

type Table struct {
	Columns   []FieldName
	Rows      []Row
	finalized bool
}

func (t *Table) AddRows(rows ...Row) {
	// TODO(manuel, 2023-06-27) Update the Columns field when adding rows
	// This might be not very efficient, but it's not like computing it all at the end is great too.
	// This might be something that could be set as an option.
	for _, row := range rows {
		columns := row.GetFields()
		t.SetColumnOrder(columns)
	}
	t.Rows = append(t.Rows, rows...)
}

// SetColumnOrder will set the given columns to be the first one to be output.
// Other columns already present in the order will be appended at the end, preserving the original order.
func (t *Table) SetColumnOrder(columns []FieldName) {
	existingColumns := map[FieldName]interface{}{}
	for _, column := range t.Columns {
		existingColumns[column] = nil
	}

	columnsToAppend := []FieldName{}

	for _, column := range columns {
		delete(existingColumns, column)
	}

	for _, column := range t.Columns {
		if _, ok := existingColumns[column]; ok {
			columnsToAppend = append(columnsToAppend, column)
		}
	}

	t.Columns = append(columns, columnsToAppend...)
}

func NewTable() *Table {
	return &Table{
		Columns:   []FieldName{},
		Rows:      []Row{},
		finalized: false,
	}
}
