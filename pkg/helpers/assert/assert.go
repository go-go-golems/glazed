package assert

import (
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func EqualRowValue(t *testing.T, expected interface{}, obj types.Row, key string) {
	v, ok := obj.Get(key)
	assert.True(t, ok)
	assert.Equal(t, expected, v)
}

func NilRowValue(t *testing.T, obj types.Row, key string) {
	_, ok := obj.Get(key)
	assert.False(t, ok)
}

func EqualRow(t *testing.T, expected types.Row, actual types.Row) {
	assert.Equal(t, expected.Len(), actual.Len())

	// test one side
	for pair, actualPair := expected.Oldest(), actual.Oldest(); actualPair != nil && pair != nil; pair, actualPair = pair.Next(), actualPair.Next() {
		k, v := pair.Key, pair.Value
		actualK, actualV := actualPair.Key, actualPair.Value
		assert.Equal(t, k, actualK)
		assert.Equal(t, v, actualV)
	}
}

func EqualRows(t *testing.T, expected []types.Row, actual []types.Row) {
	assert.Equal(t, len(expected), len(actual))

	for i := range expected {
		EqualRow(t, expected[i], actual[i])
	}
}

func EqualRowValues(t *testing.T, obj types.Row, values map[types.FieldName]types.GenericCellValue) {
	for k, v := range values {
		v_, ok := obj.Get(k)
		assert.True(t, ok)
		assert.Equal(t, v, v_)
	}
}

func EqualRowMap(t *testing.T, expected map[types.FieldName]types.GenericCellValue, actual types.Row) {
	// test one side
	for k, v := range expected {
		v_, ok := actual.Get(k)
		assert.True(t, ok)
		assert.Equal(t, v, v_)
	}

	// test other way round
	for pair := actual.Oldest(); pair != nil; pair = pair.Next() {
		k, v := pair.Key, pair.Value
		v_, ok := expected[k]
		assert.True(t, ok)
		assert.Equal(t, v, v_)
	}
}
