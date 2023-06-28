package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetColumnOrder(t *testing.T) {
	table := NewTable()

	table.SetColumnOrder([]string{"a", "b", "c"})
	assert.Equal(t, []string{"a", "b", "c"}, table.Columns)

	table = NewTable()
	table.Columns = []string{"a", "b", "c"}
	table.SetColumnOrder([]string{"a", "b", "c"})
	assert.Equal(t, []string{"a", "b", "c"}, table.Columns)

	table = NewTable()
	table.Columns = []string{"a", "b", "c"}
	table.SetColumnOrder([]string{"a", "c", "b"})
	assert.Equal(t, []string{"a", "c", "b"}, table.Columns)

	table = NewTable()
	table.Columns = []string{"a", "b", "c"}
	table.SetColumnOrder([]string{"a", "c"})
	assert.Equal(t, []string{"a", "c", "b"}, table.Columns)

	table = NewTable()
	table.Columns = []string{"a", "b", "c"}
	table.SetColumnOrder([]string{"a", "c", "d"})
	assert.Equal(t, []string{"a", "c", "d", "b"}, table.Columns)
}
