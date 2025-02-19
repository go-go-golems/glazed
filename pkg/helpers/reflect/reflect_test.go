package reflect

import (
	"reflect"
	"testing"

	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetReflectInt(t *testing.T) {
	ints := &struct {
		I   int
		I8  int8
		I16 int16
		I32 int32
		I64 int64
	}{}

	uints := &struct {
		Ui   uint
		Ui8  uint8
		Ui16 uint16
		Ui32 uint32
		Ui64 uint64
	}{}

	st := reflect.TypeOf(ints).Elem()

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		iValue := reflect.ValueOf(ints).Elem().FieldByName(field.Name)

		err := SetReflectValue(iValue, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), iValue.Int())

		// test setting i with an i8
		err = SetReflectValue(iValue, int8(2))
		require.NoError(t, err)
		assert.Equal(t, int64(2), iValue.Int())

		// test setting i with an i16
		err = SetReflectValue(iValue, int16(3))
		require.NoError(t, err)
		assert.Equal(t, int64(3), iValue.Int())

		// test setting i with an i32
		err = SetReflectValue(iValue, int32(4))
		require.NoError(t, err)
		assert.Equal(t, int64(4), iValue.Int())

		// test setting i with an i64
		err = SetReflectValue(iValue, int64(5))
		require.NoError(t, err)
		assert.Equal(t, int64(5), iValue.Int())

		// test setting i with an ui
		err = SetReflectValue(iValue, uint(6))
		require.NoError(t, err)
		assert.Equal(t, int64(6), iValue.Int())

		// test setting i with an ui8
		err = SetReflectValue(iValue, uint8(7))
		require.NoError(t, err)
		assert.Equal(t, int64(7), iValue.Int())

		// test setting i with an ui16
		err = SetReflectValue(iValue, uint16(8))
		require.NoError(t, err)
		assert.Equal(t, int64(8), iValue.Int())

		// test setting i with an ui32
		err = SetReflectValue(iValue, uint32(9))
		require.NoError(t, err)
		assert.Equal(t, int64(9), iValue.Int())

		// test setting i with an ui64
		err = SetReflectValue(iValue, uint64(10))
		require.NoError(t, err)
		assert.Equal(t, int64(10), iValue.Int())

		// try to set a string
		err = SetReflectValue(iValue, "11")
		require.NoError(t, err)
		assert.Equal(t, int64(11), iValue.Int())

		err = SetReflectValue(iValue, "abc")
		require.Error(t, err)

		// try to set a float
		err = SetReflectValue(iValue, 12.0)
		require.Error(t, err)

		// try to set a bool
		err = SetReflectValue(iValue, true)
		require.Error(t, err)
	}

	st = reflect.TypeOf(uints).Elem()

	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		iValue := reflect.ValueOf(uints).Elem().FieldByName(field.Name)

		err := SetReflectValue(iValue, 1)
		require.NoError(t, err)
		assert.Equal(t, uint64(1), iValue.Uint())

		// test setting i with an i8
		err = SetReflectValue(iValue, int8(2))
		require.NoError(t, err)
		assert.Equal(t, uint64(2), iValue.Uint())

		// test setting i with an i16
		err = SetReflectValue(iValue, int16(3))
		require.NoError(t, err)
		assert.Equal(t, uint64(3), iValue.Uint())

		// test setting i with an i32
		err = SetReflectValue(iValue, int32(4))
		require.NoError(t, err)
		assert.Equal(t, uint64(4), iValue.Uint())

		// test setting i with an i64
		err = SetReflectValue(iValue, int64(5))
		require.NoError(t, err)
		assert.Equal(t, uint64(5), iValue.Uint())

		// test setting i with an ui
		err = SetReflectValue(iValue, uint(6))
		require.NoError(t, err)
		assert.Equal(t, uint64(6), iValue.Uint())

		// test setting i with an ui8
		err = SetReflectValue(iValue, uint8(7))
		require.NoError(t, err)
		assert.Equal(t, uint64(7), iValue.Uint())

		// test setting i with an ui16
		err = SetReflectValue(iValue, uint16(8))
		require.NoError(t, err)
		assert.Equal(t, uint64(8), iValue.Uint())

		// test setting i with an ui32
		err = SetReflectValue(iValue, uint32(9))
		require.NoError(t, err)
		assert.Equal(t, uint64(9), iValue.Uint())

		// test setting i with an ui64
		err = SetReflectValue(iValue, uint64(10))
		require.NoError(t, err)
		assert.Equal(t, uint64(10), iValue.Uint())

		// try to set a string
		err = SetReflectValue(iValue, "11")
		require.NoError(t, err)
		assert.Equal(t, uint64(11), iValue.Uint())

		err = SetReflectValue(iValue, "foobar")
		require.Error(t, err)

		// try to set a float
		err = SetReflectValue(iValue, 12.0)
		require.Error(t, err)

		// try to set a bool
		err = SetReflectValue(iValue, true)
		require.Error(t, err)
	}
}

func TestReflectString(t *testing.T) {
	s := ""
	sValue := reflect.ValueOf(&s).Elem()

	err := SetReflectValue(sValue, "hello")
	require.NoError(t, err)
	assert.Equal(t, "hello", sValue.String())

	err = SetReflectValue(sValue, 1)
	require.Error(t, err)

	err = SetReflectValue(sValue, true)
	require.Error(t, err)
}

func TestReflectStringMap(t *testing.T) {
	mValue := reflect.ValueOf(&map[string]string{}).Elem()

	err := SetReflectValue(mValue, map[string]string{"hello": "world"})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"hello": "world"}, mValue.Interface())

	// try with map[string]interface{}
	err = SetReflectValue(mValue, map[string]interface{}{"hello2": "world2"})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"hello2": "world2"}, mValue.Interface())

	err = SetReflectValue(mValue, map[string]int{"hello": 1})
	require.Error(t, err)

	// fail when convert an int
	err = SetReflectValue(mValue, 1)
	require.Error(t, err)

	// fail when convert a bool
	err = SetReflectValue(mValue, true)
	require.Error(t, err)

	// fail when convert a string
	err = SetReflectValue(mValue, "hello")
	require.Error(t, err)

	err = SetReflectValue(mValue, map[string]interface{}{"hello2": 2})
	require.Error(t, err)
}

func TestReflectMap(t *testing.T) {
	i := &map[string]interface{}{}
	mValue := reflect.ValueOf(i).Elem()

	err := SetReflectValue(mValue, map[string]interface{}{"hello": "world"})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"hello": "world"}, mValue.Interface())

	// try with map[string]string
	err = SetReflectValue(mValue, map[string]string{"hello2": "world2"})
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"hello2": "world2"}, mValue.Interface())

	// we only allow map[string]interface{} and map[string]string

	// try with map[string]int
	err = SetReflectValue(mValue, map[string]int{"hello": 1})
	require.Error(t, err)

	// try with map[string][]string
	err = SetReflectValue(mValue, map[string][]string{
		"hello": {"world"},
	})
	require.Error(t, err)

	// fail when convert an int
	err = SetReflectValue(mValue, 1)
	require.Error(t, err)

	// fail when convert a bool
	err = SetReflectValue(mValue, true)
	require.Error(t, err)

	// fail when convert a string
	err = SetReflectValue(mValue, "hello")
	require.Error(t, err)
}

func TestReflectFloat(t *testing.T) {
	f := 0.0
	fValue := reflect.ValueOf(&f).Elem()

	err := SetReflectValue(fValue, 1)
	require.NoError(t, err)
	assert.Equal(t, 1.0, fValue.Float())

	err = SetReflectValue(fValue, 2.0)
	require.NoError(t, err)
	assert.Equal(t, 2.0, fValue.Float())

	// try float32
	err = SetReflectValue(fValue, float32(3.0))
	require.NoError(t, err)
	assert.Equal(t, 3.0, fValue.Float())

	// try int64
	err = SetReflectValue(fValue, int64(4))
	require.NoError(t, err)
	assert.Equal(t, 4.0, fValue.Float())

	// try uint64
	err = SetReflectValue(fValue, uint64(5))
	require.NoError(t, err)
	assert.Equal(t, 5.0, fValue.Float())

	// try int
	err = SetReflectValue(fValue, int(6))
	require.NoError(t, err)
	assert.Equal(t, 6.0, fValue.Float())

	// try uint
	err = SetReflectValue(fValue, uint(7))
	require.NoError(t, err)
	assert.Equal(t, 7.0, fValue.Float())

	// try int32
	err = SetReflectValue(fValue, int32(8))
	require.NoError(t, err)
	assert.Equal(t, 8.0, fValue.Float())

	// try uint32
	err = SetReflectValue(fValue, uint32(9))
	require.NoError(t, err)
	assert.Equal(t, 9.0, fValue.Float())

	// try int16
	err = SetReflectValue(fValue, int16(10))
	require.NoError(t, err)
	assert.Equal(t, 10.0, fValue.Float())

	// try uint16
	err = SetReflectValue(fValue, uint16(11))
	require.NoError(t, err)
	assert.Equal(t, 11.0, fValue.Float())

	// try int8
	err = SetReflectValue(fValue, int8(12))
	require.NoError(t, err)
	assert.Equal(t, 12.0, fValue.Float())

	// try uint8
	err = SetReflectValue(fValue, uint8(13))
	require.NoError(t, err)
	assert.Equal(t, 13.0, fValue.Float())

	err = SetReflectValue(fValue, "1")
	require.Error(t, err)

	err = SetReflectValue(fValue, true)
	require.Error(t, err)
}

func TestReflectBool(t *testing.T) {
	b := false
	bValue := reflect.ValueOf(&b).Elem()

	err := SetReflectValue(bValue, true)
	require.NoError(t, err)
	assert.Equal(t, true, bValue.Bool())

	err = SetReflectValue(bValue, false)
	require.NoError(t, err)
	assert.Equal(t, false, bValue.Bool())

	err = SetReflectValue(bValue, 1)
	require.Error(t, err)

	err = SetReflectValue(bValue, "true")
	require.Error(t, err)
}

func TestReflectStringSlice(t *testing.T) {
	s := []string{}
	sValue := reflect.ValueOf(&s).Elem()

	err := SetReflectValue(sValue, []string{"hello", "world"})
	require.NoError(t, err)
	assert.Equal(t, []string{"hello", "world"}, sValue.Interface())

	err = SetReflectValue(sValue, []interface{}{"hello", "world"})
	require.NoError(t, err)
	assert.Equal(t, []string{"hello", "world"}, sValue.Interface())

	err = SetReflectValue(sValue, []int{1, 2})
	require.Error(t, err)

	err = SetReflectValue(sValue, []interface{}{"hello", 1})
	require.Error(t, err)
}

func TestReflectIntegerSlice(t *testing.T) {
	s := []int{}
	s2 := []int64{}
	s4 := []int32{}
	s6 := []int16{}
	s8 := []int8{}

	s10 := []uint{}
	s3 := []uint64{}
	s5 := []uint32{}
	s7 := []uint16{}
	s9 := []uint8{}

	testSetReflectIntList[int](t, reflect.ValueOf(&s).Elem())
	testSetReflectIntList[int64](t, reflect.ValueOf(&s2).Elem())
	testSetReflectIntList[int32](t, reflect.ValueOf(&s4).Elem())
	testSetReflectIntList[int16](t, reflect.ValueOf(&s6).Elem())
	testSetReflectIntList[int8](t, reflect.ValueOf(&s8).Elem())

	testSetReflectIntList[uint](t, reflect.ValueOf(&s10).Elem())
	testSetReflectIntList[uint64](t, reflect.ValueOf(&s3).Elem())
	testSetReflectIntList[uint32](t, reflect.ValueOf(&s5).Elem())
	testSetReflectIntList[uint16](t, reflect.ValueOf(&s7).Elem())
	testSetReflectIntList[uint8](t, reflect.ValueOf(&s9).Elem())

}

func TestReflectFloatSlice(t *testing.T) {
	s := []float64{}
	s2 := []float32{}

	testSetReflectFloatList[float64](t, reflect.ValueOf(&s).Elem())
	testSetReflectFloatList[float32](t, reflect.ValueOf(&s2).Elem())
}

func testSetReflectIntList[T cast.Number](t *testing.T, sValue reflect.Value) {
	err := SetReflectValue(sValue, []int{1, 2})
	require.NoError(t, err)
	assert.Equal(t, []T{1, 2}, sValue.Interface())

	err = SetReflectValue(sValue, []interface{}{3, 4})
	require.NoError(t, err)
	assert.Equal(t, []T{3, 4}, sValue.Interface())

	// try list of int64
	err = SetReflectValue(sValue, []int64{5, 6})
	require.NoError(t, err)
	assert.Equal(t, []T{5, 6}, sValue.Interface())

	// try list of uint64
	err = SetReflectValue(sValue, []uint64{7, 8})
	require.NoError(t, err)
	assert.Equal(t, []T{7, 8}, sValue.Interface())

	// try list of int32
	err = SetReflectValue(sValue, []int32{9, 10})
	require.NoError(t, err)
	assert.Equal(t, []T{9, 10}, sValue.Interface())

	// try list of uint32
	err = SetReflectValue(sValue, []uint32{11, 12})
	require.NoError(t, err)
	assert.Equal(t, []T{11, 12}, sValue.Interface())

	// try list of int16
	err = SetReflectValue(sValue, []int16{13, 14})
	require.NoError(t, err)
	assert.Equal(t, []T{13, 14}, sValue.Interface())

	// try list of uint16
	err = SetReflectValue(sValue, []uint16{15, 16})
	require.NoError(t, err)
	assert.Equal(t, []T{15, 16}, sValue.Interface())

	// try list of int8
	err = SetReflectValue(sValue, []int8{17, 18})
	require.NoError(t, err)
	assert.Equal(t, []T{17, 18}, sValue.Interface())

	// try list of uint8
	err = SetReflectValue(sValue, []uint8{19, 20})
	require.NoError(t, err)
	assert.Equal(t, []T{19, 20}, sValue.Interface())

	err = SetReflectValue(sValue, []string{"hello", "world"})
	require.Error(t, err)

	err = SetReflectValue(sValue, []interface{}{1, "world"})
	require.Error(t, err)
}

func testSetReflectFloatList[T cast.FloatNumber](t *testing.T, sValue reflect.Value) {
	err := SetReflectValue(sValue, []float64{1.1, 2.2})
	require.NoError(t, err)
	assert.InDeltaSlice(t, []T{1.1, 2.2}, sValue.Interface(), 0.01)

	err = SetReflectValue(sValue, []interface{}{3.3, 4.4})
	require.NoError(t, err)
	assert.InDeltaSlice(t, []T{3.3, 4.4}, sValue.Interface(), 0.01)

	err = SetReflectValue(sValue, []float32{5.5, 6.6})
	require.NoError(t, err)
	assert.InDeltaSlice(t, []T{5.5, 6.6}, sValue.Interface(), 0.01)

	// now integers
	err = SetReflectValue(sValue, []int{7, 8})
	require.NoError(t, err)
	assert.InDeltaSlice(t, []T{7, 8}, sValue.Interface(), 0.01)

	// now uints
	err = SetReflectValue(sValue, []uint{9, 10})
	require.NoError(t, err)
	assert.InDeltaSlice(t, []T{9, 10}, sValue.Interface(), 0.01)

	// now int64
	err = SetReflectValue(sValue, []int64{11, 12})
	require.NoError(t, err)
	assert.InDeltaSlice(t, []T{11, 12}, sValue.Interface(), 0.01)

	// now uint64
	err = SetReflectValue(sValue, []uint64{13, 14})
	require.NoError(t, err)
	assert.InDeltaSlice(t, []T{13, 14}, sValue.Interface(), 0.01)

	// now int32
	err = SetReflectValue(sValue, []int32{15, 16})
	require.NoError(t, err)
	assert.InDeltaSlice(t, []T{15, 16}, sValue.Interface(), 0.01)

	// now uint32
	err = SetReflectValue(sValue, []uint32{17, 18})
	require.NoError(t, err)
	assert.InDeltaSlice(t, []T{17, 18}, sValue.Interface(), 0.01)

	// now int16
	err = SetReflectValue(sValue, []int16{19, 20})
	require.NoError(t, err)
	assert.InDeltaSlice(t, []T{19, 20}, sValue.Interface(), 0.01)

	// now uint16
	err = SetReflectValue(sValue, []uint16{21, 22})
	require.NoError(t, err)
	assert.InDeltaSlice(t, []T{21, 22}, sValue.Interface(), 0.01)

	// now int8
	err = SetReflectValue(sValue, []int8{23, 24})
	require.NoError(t, err)
	assert.InDeltaSlice(t, []T{23, 24}, sValue.Interface(), 0.01)

	// now uint8
	err = SetReflectValue(sValue, []uint8{25, 26})
	require.NoError(t, err)
	assert.InDeltaSlice(t, []T{25, 26}, sValue.Interface(), 0.01)

	err = SetReflectValue(sValue, []string{"hello", "world"})
	require.Error(t, err)

	err = SetReflectValue(sValue, []interface{}{1.1, "world"})
	require.Error(t, err)
}

// Test that passing a list of strings as an interface{} gets unwrapped properly
func TestSetReflectStringListInterface(t *testing.T) {
	s := []string{}
	sValue := reflect.ValueOf(&s).Elem()

	err := SetReflectValue(sValue, []interface{}{"hello", "world"})
	require.NoError(t, err)
	assert.Equal(t, []string{"hello", "world"}, sValue.Interface())

	err = SetReflectValue(sValue, []interface{}{"hello", 1})
	require.Error(t, err)
}

type StringAlias string
type StringDeclaration string

func TestSetReflectString(t *testing.T) {
	var s string
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), "test")
	assert.NoError(t, err)
	assert.Equal(t, "test", s)
}

func TestSetReflectStringFromAlias(t *testing.T) {
	var s string
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), StringAlias("test"))
	assert.NoError(t, err)
	assert.Equal(t, "test", s)
}

func TestSetReflectStringFromDeclaration(t *testing.T) {
	var s string
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), StringDeclaration("test"))
	assert.NoError(t, err)
	assert.Equal(t, "test", s)
}

func TestSetReflectStringAlias(t *testing.T) {
	var s StringAlias
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), "test")
	assert.NoError(t, err)
	assert.Equal(t, StringAlias("test"), s)
}

func TestSetReflectStringAliasFromAlias(t *testing.T) {
	var s StringAlias
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), StringAlias("test"))
	assert.NoError(t, err)
	assert.Equal(t, StringAlias("test"), s)
}

func TestSetReflectStringAliasFromDeclaration(t *testing.T) {
	var s StringAlias
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), StringDeclaration("test"))
	assert.NoError(t, err)
	assert.Equal(t, StringAlias("test"), s)
}

func TestSetReflectStringDeclaration(t *testing.T) {
	var s StringDeclaration
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), "test")
	assert.NoError(t, err)
	assert.Equal(t, StringDeclaration("test"), s)
}

func TestSetReflectStringDeclarationFromAlias(t *testing.T) {
	var s StringDeclaration
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), StringAlias("test"))
	assert.NoError(t, err)
	assert.Equal(t, StringDeclaration("test"), s)
}

func TestSetReflectStringDeclarationFromDeclaration(t *testing.T) {
	var s StringDeclaration
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), StringDeclaration("test"))
	assert.NoError(t, err)
	assert.Equal(t, StringDeclaration("test"), s)
}

func TestSetReflectStringSlice(t *testing.T) {
	var s []string
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), []string{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, s)
}

func TestSetReflectStringSliceFromAlias(t *testing.T) {
	var s []string
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), []StringAlias{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, s)
}

func TestSetReflectStringSliceFromDeclaration(t *testing.T) {
	var s []string
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), []StringDeclaration{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, s)
}

func TestSetReflectStringAliasSlice(t *testing.T) {
	var s []StringAlias
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), []string{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, []StringAlias{"a", "b"}, s)
}

func TestSetReflectStringAliasSliceFromAlias(t *testing.T) {
	var s []StringAlias
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), []StringAlias{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, []StringAlias{"a", "b"}, s)
}

func TestSetReflectStringAliasSliceFromDeclaration(t *testing.T) {
	var s []StringAlias
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), []StringDeclaration{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, []StringAlias{"a", "b"}, s)
}

func TestSetReflectStringDeclarationSlice(t *testing.T) {
	var s []StringDeclaration
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), []string{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, []StringDeclaration{"a", "b"}, s)
}

func TestSetReflectStringDeclarationSliceFromAlias(t *testing.T) {
	var s []StringDeclaration
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), []StringAlias{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, []StringDeclaration{"a", "b"}, s)
}

func TestSetReflectStringDeclarationSliceFromDeclaration(t *testing.T) {
	var s []StringDeclaration
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), []StringDeclaration{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, []StringDeclaration{"a", "b"}, s)
}

func TestSetReflectStringSliceFromInterface(t *testing.T) {
	var s []string
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), []interface{}{"a", "b"})
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, s)
}

func TestSetReflectStringError(t *testing.T) {
	var s string
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), 42)
	assert.Error(t, err)
}

func TestSetReflectStringSliceError(t *testing.T) {
	var s []string
	err := SetReflectValue(reflect.ValueOf(&s).Elem(), []int{1, 2})
	assert.Error(t, err)
}

func TestStripInterface(t *testing.T) {
	type CustomString string
	type CustomInt int
	type Wrapper struct {
		Value interface{}
	}

	tests := []struct {
		name     string
		input    interface{}
		expected reflect.Type
	}{
		{
			name:     "plain string",
			input:    "hello",
			expected: reflect.TypeOf(""),
		},
		{
			name:     "pointer to string",
			input:    func() interface{} { s := "hello"; return &s }(),
			expected: reflect.TypeOf(""),
		},
		{
			name:     "empty interface",
			input:    interface{}(nil),
			expected: nil,
		},
		{
			name:     "interface containing string",
			input:    interface{}("hello"),
			expected: reflect.TypeOf(""),
		},
		{
			name: "nested interfaces and pointers",
			input: func() interface{} {
				s := interface{}("test")
				return &s
			}(),
			expected: reflect.TypeOf(""),
		},
		{
			name:     "custom string type",
			input:    CustomString("hello"),
			expected: reflect.TypeOf(CustomString("")),
		},
		{
			name:     "interface containing custom string",
			input:    interface{}(CustomString("hello")),
			expected: reflect.TypeOf(CustomString("")),
		},
		{
			name: "pointer to interface containing custom string",
			input: func() interface{} {
				i := interface{}(CustomString("hello"))
				return &i
			}(),
			expected: reflect.TypeOf(CustomString("")),
		},
		{
			name:     "int",
			input:    42,
			expected: reflect.TypeOf(0),
		},
		{
			name:     "custom int",
			input:    CustomInt(42),
			expected: reflect.TypeOf(CustomInt(0)),
		},
		{
			name: "deeply nested interfaces",
			input: func() interface{} {
				s := "test"
				i1 := interface{}(s)
				i2 := interface{}(&i1)
				i3 := interface{}(&i2)
				return &i3
			}(),
			expected: reflect.TypeOf(""),
		},
		{
			name: "struct containing interface",
			input: Wrapper{
				Value: "hello",
			},
			expected: reflect.TypeOf(Wrapper{}),
		},
		{
			name:     "nil pointer to string",
			input:    (*string)(nil),
			expected: nil,
		},
		{
			name: "interface containing nil pointer",
			input: func() interface{} {
				var s *string
				return s
			}(),
			expected: nil,
		},
		{
			name:     "slice of strings",
			input:    []string{"hello", "world"},
			expected: reflect.TypeOf([]string{}),
		},
		{
			name:     "interface containing slice",
			input:    interface{}([]string{"hello", "world"}),
			expected: reflect.TypeOf([]string{}),
		},
		{
			name:     "map of strings",
			input:    map[string]string{"hello": "world"},
			expected: reflect.TypeOf(map[string]string{}),
		},
		{
			name: "interface containing map",
			input: interface{}(map[string]string{
				"hello": "world",
			}),
			expected: reflect.TypeOf(map[string]string{}),
		},
		{
			name: "interface containing interface containing string",
			input: func() interface{} {
				s := interface{}("hello")
				return interface{}(s)
			}(),
			expected: reflect.TypeOf(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripInterface(reflect.ValueOf(tt.input))
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStripInterfaceFromValue(t *testing.T) {
	type CustomString string
	type CustomInt int
	type Wrapper struct {
		Value interface{}
	}

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "plain string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "pointer to string",
			input:    func() interface{} { s := "hello"; return &s }(),
			expected: "hello",
		},
		{
			name:     "empty interface",
			input:    interface{}(nil),
			expected: nil,
		},
		{
			name:     "interface containing string",
			input:    interface{}("hello"),
			expected: "hello",
		},
		{
			name: "nested interfaces and pointers",
			input: func() interface{} {
				s := interface{}("test")
				return &s
			}(),
			expected: "test",
		},
		{
			name:     "custom string type",
			input:    CustomString("hello"),
			expected: CustomString("hello"),
		},
		{
			name:     "interface containing custom string",
			input:    interface{}(CustomString("hello")),
			expected: CustomString("hello"),
		},
		{
			name: "pointer to interface containing custom string",
			input: func() interface{} {
				i := interface{}(CustomString("hello"))
				return &i
			}(),
			expected: CustomString("hello"),
		},
		{
			name:     "int",
			input:    42,
			expected: 42,
		},
		{
			name:     "custom int",
			input:    CustomInt(42),
			expected: CustomInt(42),
		},
		{
			name: "deeply nested interfaces",
			input: func() interface{} {
				s := "test"
				i1 := interface{}(s)
				i2 := interface{}(&i1)
				i3 := interface{}(&i2)
				return &i3
			}(),
			expected: "test",
		},
		{
			name: "struct containing interface",
			input: Wrapper{
				Value: "hello",
			},
			expected: Wrapper{Value: "hello"},
		},
		{
			name:     "nil pointer to string",
			input:    (*string)(nil),
			expected: nil,
		},
		{
			name: "interface containing nil pointer",
			input: func() interface{} {
				var s *string
				return s
			}(),
			expected: nil,
		},
		{
			name:     "slice of strings",
			input:    []string{"hello", "world"},
			expected: []string{"hello", "world"},
		},
		{
			name:     "interface containing slice",
			input:    interface{}([]string{"hello", "world"}),
			expected: []string{"hello", "world"},
		},
		{
			name:     "map of strings",
			input:    map[string]string{"hello": "world"},
			expected: map[string]string{"hello": "world"},
		},
		{
			name: "interface containing map",
			input: interface{}(map[string]string{
				"hello": "world",
			}),
			expected: map[string]string{"hello": "world"},
		},
		{
			name: "interface containing interface containing string",
			input: func() interface{} {
				s := interface{}("hello")
				return interface{}(s)
			}(),
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripInterfaceFromValue(reflect.ValueOf(tt.input))
			if tt.expected == nil {
				assert.False(t, result.IsValid())
			} else {
				assert.Equal(t, tt.expected, result.Interface())
			}
		})
	}
}

func TestStripInterfaceValue(t *testing.T) {
	type CustomString string
	type CustomInt int
	type Wrapper struct {
		Value interface{}
	}

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "plain string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "pointer to string",
			input:    func() interface{} { s := "hello"; return &s }(),
			expected: "hello",
		},
		{
			name:     "empty interface",
			input:    interface{}(nil),
			expected: nil,
		},
		{
			name:     "interface containing string",
			input:    interface{}("hello"),
			expected: "hello",
		},
		{
			name: "nested interfaces and pointers",
			input: func() interface{} {
				s := interface{}("test")
				return &s
			}(),
			expected: "test",
		},
		{
			name:     "custom string type",
			input:    CustomString("hello"),
			expected: CustomString("hello"),
		},
		{
			name:     "interface containing custom string",
			input:    interface{}(CustomString("hello")),
			expected: CustomString("hello"),
		},
		{
			name: "pointer to interface containing custom string",
			input: func() interface{} {
				i := interface{}(CustomString("hello"))
				return &i
			}(),
			expected: CustomString("hello"),
		},
		{
			name:     "int",
			input:    42,
			expected: 42,
		},
		{
			name:     "custom int",
			input:    CustomInt(42),
			expected: CustomInt(42),
		},
		{
			name: "deeply nested interfaces",
			input: func() interface{} {
				s := "test"
				i1 := interface{}(s)
				i2 := interface{}(&i1)
				i3 := interface{}(&i2)
				return &i3
			}(),
			expected: "test",
		},
		{
			name: "struct containing interface",
			input: Wrapper{
				Value: "hello",
			},
			expected: Wrapper{Value: "hello"},
		},
		{
			name:     "nil pointer to string",
			input:    (*string)(nil),
			expected: nil,
		},
		{
			name: "interface containing nil pointer",
			input: func() interface{} {
				var s *string
				return s
			}(),
			expected: nil,
		},
		{
			name:     "slice of strings",
			input:    []string{"hello", "world"},
			expected: []string{"hello", "world"},
		},
		{
			name:     "interface containing slice",
			input:    interface{}([]string{"hello", "world"}),
			expected: []string{"hello", "world"},
		},
		{
			name:     "map of strings",
			input:    map[string]string{"hello": "world"},
			expected: map[string]string{"hello": "world"},
		},
		{
			name: "interface containing map",
			input: interface{}(map[string]string{
				"hello": "world",
			}),
			expected: map[string]string{"hello": "world"},
		},
		{
			name: "interface containing interface containing string",
			input: func() interface{} {
				s := interface{}("hello")
				return interface{}(s)
			}(),
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripInterfaceValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
