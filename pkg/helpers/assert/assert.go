package assert

import (
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func EqualMapRowValue(t *testing.T, expected interface{}, obj types.MapRow, key string) {
	v, ok := obj.Get(key)
	assert.True(t, ok)
	assert.Equal(t, expected, v)
}

func NilMapRowValue(t *testing.T, obj types.MapRow, key string) {
	_, ok := obj.Get(key)
	assert.False(t, ok)
}

func EqualMapRows(t *testing.T, expected types.MapRow, actual types.MapRow) {
	// test one side
	for pair := expected.Oldest(); pair != nil; pair = pair.Next() {
		k, v := pair.Key, pair.Value
		v_, ok := actual.Get(k)
		assert.True(t, ok)
		assert.Equal(t, v, v_)
	}

	// test other way round
	for pair := actual.Oldest(); pair != nil; pair = pair.Next() {
		k, v := pair.Key, pair.Value
		v_, ok := expected.Get(k)
		assert.True(t, ok)
		assert.Equal(t, v, v_)
	}
}

func EqualMapRowValues(t *testing.T, obj types.MapRow, values map[types.FieldName]types.GenericCellValue) {
	for k, v := range values {
		v_, ok := obj.Get(k)
		assert.True(t, ok)
		assert.Equal(t, v, v_)
	}
}

func EqualMapRowMap(t *testing.T, expected map[types.FieldName]types.GenericCellValue, actual types.MapRow) {
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
