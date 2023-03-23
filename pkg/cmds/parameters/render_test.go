package parameters

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRenderString(t *testing.T) {
	v, err := RenderValue(ParameterTypeString, "foobar")
	require.NoError(t, err)
	assert.Equal(t, "foobar", v)

	_, err = RenderValue(ParameterTypeString, 1)
	require.Error(t, err)

	v, err = RenderValue(ParameterTypeStringFromFile, "foobar")
	require.NoError(t, err)
	assert.Equal(t, "foobar", v)

	v, err = RenderValue(ParameterTypeChoice, "foobar")
	require.NoError(t, err)
	assert.Equal(t, "foobar", v)
}

func TestRenderStringList(t *testing.T) {
	v, err := RenderValue(ParameterTypeStringList, []string{"foobar", "barfoo"})
	require.NoError(t, err)
	assert.Equal(t, "foobar,barfoo", v)

	v, err = RenderValue(ParameterTypeStringListFromFile, []string{"foobar", "barfoo"})
	require.NoError(t, err)
	assert.Equal(t, "foobar,barfoo", v)

	v, err = RenderValue(ParameterTypeStringList, []string{"foobar"})
	require.NoError(t, err)
	assert.Equal(t, "foobar", v)

	_, err = RenderValue(ParameterTypeStringList, 1)
	require.Error(t, err)

	_, err = RenderValue(ParameterTypeStringList, "foobar")
	require.Error(t, err)
}

func TestRenderInt(t *testing.T) {
	v, err := RenderValue(ParameterTypeInteger, 1)
	require.NoError(t, err)
	assert.Equal(t, "1", v)

	_, err = RenderValue(ParameterTypeInteger, "foobar")
	require.Error(t, err)

	v, err = RenderValue(ParameterTypeIntegerList, []int{1, 2, 3})
	require.NoError(t, err)
	assert.Equal(t, "1,2,3", v)

	v, err = RenderValue(ParameterTypeIntegerList, []int{1})
	require.NoError(t, err)
	assert.Equal(t, "1", v)
}

func TestRenderFloat(t *testing.T) {
	v, err := RenderValue(ParameterTypeFloat, 1.1)
	require.NoError(t, err)
	assert.Equal(t, "1.100000", v)

	_, err = RenderValue(ParameterTypeFloat, "foobar")
	require.Error(t, err)

	v, err = RenderValue(ParameterTypeFloatList, []float64{1.1, 2.2, 3.3})
	require.NoError(t, err)
	assert.Equal(t, "1.100000,2.200000,3.300000", v)

	v, err = RenderValue(ParameterTypeFloatList, []float64{1.1})
	require.NoError(t, err)
	assert.Equal(t, "1.100000", v)
}
