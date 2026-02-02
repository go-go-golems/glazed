package fields

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRenderString(t *testing.T) {
	v, err := RenderValue(TypeString, "foobar")
	require.NoError(t, err)
	assert.Equal(t, "foobar", v)

	_, err = RenderValue(TypeString, 1)
	require.Error(t, err)

	v, err = RenderValue(TypeStringFromFile, "foobar")
	require.NoError(t, err)
	assert.Equal(t, "foobar", v)

	v, err = RenderValue(TypeChoice, "foobar")
	require.NoError(t, err)
	assert.Equal(t, "foobar", v)
}

func TestRenderStringList(t *testing.T) {
	v, err := RenderValue(TypeStringList, []string{"foobar", "barfoo"})
	require.NoError(t, err)
	assert.Equal(t, "foobar,barfoo", v)

	v, err = RenderValue(TypeStringListFromFile, []string{"foobar", "barfoo"})
	require.NoError(t, err)
	assert.Equal(t, "foobar,barfoo", v)

	v, err = RenderValue(TypeStringList, []string{"foobar"})
	require.NoError(t, err)
	assert.Equal(t, "foobar", v)

	_, err = RenderValue(TypeStringList, 1)
	require.Error(t, err)

	_, err = RenderValue(TypeStringList, "foobar")
	require.Error(t, err)
}

func TestRenderInt(t *testing.T) {
	v, err := RenderValue(TypeInteger, 1)
	require.NoError(t, err)
	assert.Equal(t, "1", v)

	_, err = RenderValue(TypeInteger, "foobar")
	require.Error(t, err)

	v, err = RenderValue(TypeIntegerList, []int{1, 2, 3})
	require.NoError(t, err)
	assert.Equal(t, "1,2,3", v)

	v, err = RenderValue(TypeIntegerList, []int{1})
	require.NoError(t, err)
	assert.Equal(t, "1", v)
}

func TestRenderFloat(t *testing.T) {
	v, err := RenderValue(TypeFloat, 1.1)
	require.NoError(t, err)
	assert.Equal(t, "1.100000", v)

	_, err = RenderValue(TypeFloat, "foobar")
	require.Error(t, err)

	v, err = RenderValue(TypeFloatList, []float64{1.1, 2.2, 3.3})
	require.NoError(t, err)
	assert.Equal(t, "1.100000,2.200000,3.300000", v)

	v, err = RenderValue(TypeFloatList, []float64{1.1})
	require.NoError(t, err)
	assert.Equal(t, "1.100000", v)
}

func TestRenderDate(t *testing.T) {
	v, err := RenderValue(TypeDate, "2019-01-01")
	require.NoError(t, err)
	assert.Equal(t, "2019-01-01", v)

	_, err = RenderValue(TypeDate, 1)
	require.Error(t, err)

	d, err := time.Parse("2006-01-02", "2019-01-01")
	require.NoError(t, err)
	v, err = RenderValue(TypeDate, d)
	require.NoError(t, err)
	assert.Equal(t, "2019-01-01T00:00:00Z", v)

	t_, err := time.Parse("2006-01-02 15:04:05", "2019-01-01 12:00:00")
	require.NoError(t, err)
	v, err = RenderValue(TypeDate, t_)
	require.NoError(t, err)
	assert.Equal(t, "2019-01-01T12:00:00Z", v)
}
