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
		if _, ok := existingColumns[column]; ok {
			delete(existingColumns, column)
		}
	}

	for _, column := range t.Columns {
		if _, ok := existingColumns[column]; ok {
			columnsToAppend = append(columnsToAppend, column)
		}
	}

	t.Columns = append(columns, columnsToAppend...)
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
