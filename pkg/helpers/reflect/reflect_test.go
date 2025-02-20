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

func TestStripInterfaceValue_PlainString(t *testing.T) {
	result := StripInterfaceValue("hello")
	assert.Equal(t, "hello", result)
}

func TestStripInterfaceValue_PointerToString(t *testing.T) {
	s := "hello"
	result := StripInterfaceValue(&s)
	assert.Equal(t, &s, result)
}

func TestStripInterfaceValue_EmptyInterface(t *testing.T) {
	var i interface{}
	result := StripInterfaceValue(i)
	assert.Nil(t, result)
}

func TestStripInterfaceValue_InterfaceContainingString(t *testing.T) {
	var i interface{} = "hello"
	result := StripInterfaceValue(i)
	assert.Equal(t, "hello", result)
}

func TestStripInterfaceValue_PointerToInterfaceContainingString(t *testing.T) {
	var s interface{} = "test"
	ps := &s
	result := StripInterfaceValue(ps)
	assert.IsType(t, (*string)(nil), result)
	assert.Equal(t, "test", *result.(*string))
}

func TestStripInterfaceValue_CustomString(t *testing.T) {
	type CustomString string
	result := StripInterfaceValue(CustomString("hello"))
	assert.Equal(t, CustomString("hello"), result)
}

func TestStripInterfaceValue_InterfaceContainingCustomString(t *testing.T) {
	type CustomString string
	var i interface{} = CustomString("hello")
	result := StripInterfaceValue(i)
	assert.Equal(t, CustomString("hello"), result)
}

func TestStripInterfaceValue_PointerToInterfaceContainingCustomString(t *testing.T) {
	type CustomString string
	var i interface{} = CustomString("hello")
	pi := &i
	result := StripInterfaceValue(pi)
	assert.IsType(t, (*CustomString)(nil), result)
	assert.Equal(t, CustomString("hello"), *result.(*CustomString))
}

func TestStripInterfaceValue_Int(t *testing.T) {
	result := StripInterfaceValue(42)
	assert.Equal(t, 42, result)
}

func TestStripInterfaceValue_CustomInt(t *testing.T) {
	type CustomInt int
	result := StripInterfaceValue(CustomInt(42))
	assert.Equal(t, CustomInt(42), result)
}

func TestStripInterfaceValue_DeeplyNestedInterfacesWithPointers(t *testing.T) {
	s := "test"
	i1 := interface{}(s)
	i2 := interface{}(&i1)
	i3 := interface{}(&i2)
	pi3 := &i3

	result := StripInterfaceValue(pi3)
	assert.IsType(t, (***string)(nil), result)
	assert.Equal(t, "test", *(*(*(result.(***string)))))
}

func TestStripInterfaceValue_StructContainingInterface(t *testing.T) {
	type Wrapper struct {
		Value interface{}
	}
	w := Wrapper{Value: "hello"}
	result := StripInterfaceValue(w)
	assert.Equal(t, w, result)
}

func TestStripInterfaceValue_NilPointerToString(t *testing.T) {
	var s *string
	result := StripInterfaceValue(s)
	assert.Nil(t, result)
}

func TestStripInterfaceValue_InterfaceContainingNilPointer(t *testing.T) {
	var s *string
	var i interface{} = s
	result := StripInterfaceValue(i)
	assert.Nil(t, result)
}

func TestStripInterfaceValue_SliceOfStrings(t *testing.T) {
	slice := []string{"hello", "world"}
	result := StripInterfaceValue(slice)
	assert.Equal(t, slice, result)
}

func TestStripInterfaceValue_InterfaceContainingSlice(t *testing.T) {
	var i interface{} = []string{"hello", "world"}
	result := StripInterfaceValue(i)
	assert.Equal(t, []string{"hello", "world"}, result)
}

func TestStripInterfaceValue_MapOfStrings(t *testing.T) {
	m := map[string]string{"hello": "world"}
	result := StripInterfaceValue(m)
	assert.Equal(t, m, result)
}

func TestStripInterfaceValue_InterfaceContainingMap(t *testing.T) {
	var i interface{} = map[string]string{"hello": "world"}
	result := StripInterfaceValue(i)
	assert.Equal(t, map[string]string{"hello": "world"}, result)
}

func TestStripInterface_PointerToString(t *testing.T) {
	s := "hello"
	result := StripInterface(reflect.ValueOf(&s))
	assert.Equal(t, reflect.TypeOf((*string)(nil)), result)
}

func TestStripInterface_EmptyInterface(t *testing.T) {
	var i interface{}
	result := StripInterface(reflect.ValueOf(i))
	assert.Nil(t, result)
}

func TestStripInterface_InterfaceContainingString(t *testing.T) {
	var i interface{} = "hello"
	result := StripInterface(reflect.ValueOf(i))
	assert.Equal(t, reflect.TypeOf(""), result)
}

func TestStripInterface_PointerToInterfaceContainingString(t *testing.T) {
	var s interface{} = "test"
	ps := &s
	result := StripInterface(reflect.ValueOf(ps))
	assert.Equal(t, reflect.TypeOf((*string)(nil)), result)
}

func TestStripInterface_CustomString(t *testing.T) {
	type CustomString string
	result := StripInterface(reflect.ValueOf(CustomString("hello")))
	assert.Equal(t, reflect.TypeOf(CustomString("")), result)
}

func TestStripInterface_InterfaceContainingCustomString(t *testing.T) {
	type CustomString string
	var i interface{} = CustomString("hello")
	result := StripInterface(reflect.ValueOf(i))
	assert.Equal(t, reflect.TypeOf(CustomString("")), result)
}

func TestStripInterface_PointerToInterfaceContainingCustomString(t *testing.T) {
	type CustomString string
	var i interface{} = CustomString("hello")
	pi := &i
	result := StripInterface(reflect.ValueOf(pi))
	assert.Equal(t, reflect.TypeOf((*CustomString)(nil)), result)
}

func TestStripInterface_Int(t *testing.T) {
	result := StripInterface(reflect.ValueOf(42))
	assert.Equal(t, reflect.TypeOf(0), result)
}

func TestStripInterface_CustomInt(t *testing.T) {
	type CustomInt int
	result := StripInterface(reflect.ValueOf(CustomInt(42)))
	assert.Equal(t, reflect.TypeOf(CustomInt(0)), result)
}

func TestStripInterface_DeeplyNestedInterfacesWithPointers(t *testing.T) {
	s := "test"
	i1 := interface{}(s)
	i2 := interface{}(&i1)
	i3 := interface{}(&i2)
	pi3 := &i3

	result := StripInterface(reflect.ValueOf(pi3))
	expectedType := result.String()
	assert.Equal(t, "***string", expectedType)
}

func TestStripInterface_StructContainingInterface(t *testing.T) {
	type Wrapper struct {
		Value interface{}
	}
	w := Wrapper{Value: "hello"}
	result := StripInterface(reflect.ValueOf(w))
	assert.Equal(t, reflect.TypeOf(Wrapper{}), result)
}

func TestStripInterface_NilPointerToString(t *testing.T) {
	var s *string
	result := StripInterface(reflect.ValueOf(s))
	assert.Nil(t, result)
}

func TestStripInterface_InterfaceContainingNilPointer(t *testing.T) {
	var s *string
	var i interface{} = s
	result := StripInterface(reflect.ValueOf(i))
	assert.Nil(t, result)
}

func TestStripInterface_SliceOfStrings(t *testing.T) {
	slice := []string{"hello", "world"}
	result := StripInterface(reflect.ValueOf(slice))
	assert.Equal(t, reflect.TypeOf([]string{}), result)
}

func TestStripInterface_InterfaceContainingSlice(t *testing.T) {
	var i interface{} = []string{"hello", "world"}
	result := StripInterface(reflect.ValueOf(i))
	assert.Equal(t, reflect.TypeOf([]string{}), result)
}

func TestStripInterface_MapOfStrings(t *testing.T) {
	m := map[string]string{"hello": "world"}
	result := StripInterface(reflect.ValueOf(m))
	assert.Equal(t, reflect.TypeOf(map[string]string{}), result)
}

func TestStripInterface_InterfaceContainingMap(t *testing.T) {
	var i interface{} = map[string]string{"hello": "world"}
	result := StripInterface(reflect.ValueOf(i))
	assert.Equal(t, reflect.TypeOf(map[string]string{}), result)
}

func TestStripInterface_DoublePointerToString(t *testing.T) {
	s := "test"
	ps := &s
	pps := &ps
	result := StripInterfaceValue(pps)
	assert.IsType(t, (**string)(nil), result)
	assert.Equal(t, &s, *result.(**string))
}

func TestStripInterface_PointerToInterfaceToString(t *testing.T) {
	var s interface{} = "test"
	ps := &s
	result := StripInterfaceValue(ps)
	assert.IsType(t, (*string)(nil), result)
	assert.Equal(t, "test", *result.(*string))
}

func TestStripInterface_PointerToInterfaceToCustomType(t *testing.T) {
	type CustomString string
	var s interface{} = CustomString("test")
	ps := &s
	result := StripInterfaceValue(ps)
	assert.IsType(t, (*CustomString)(nil), result)
	assert.Equal(t, CustomString("test"), *result.(*CustomString))
}

func TestStripInterface_DoublePointerToInterfaceToString(t *testing.T) {
	var s interface{} = "test"
	ps := &s
	pps := &ps
	result := StripInterfaceValue(pps)
	assert.IsType(t, (**string)(nil), result)
	assert.Equal(t, s, **result.(**string))
}

func TestStripInterface_PointerToStructWithInterfaceField(t *testing.T) {
	type Wrapper struct {
		Value interface{}
	}
	w := &Wrapper{Value: "hello"}
	result := StripInterfaceValue(w)
	assert.Equal(t, w, result)
}

func TestStripInterfaceValue_InterfaceContainingInterfaceContainingString(t *testing.T) {
	s := interface{}("hello")
	i := interface{}(s)
	result := StripInterfaceValue(i)
	assert.Equal(t, "hello", result)
}

func TestStripInterfaceValue_DoublePointerToString(t *testing.T) {
	s := "test"
	ps := &s
	pps := &ps
	result := StripInterfaceValue(pps)
	assert.IsType(t, (**string)(nil), result)
	assert.Equal(t, &s, *result.(**string))
}

func TestStripInterfaceValue_PointerToStructWithInterfaceField(t *testing.T) {
	type Wrapper struct {
		Value interface{}
	}
	w := &Wrapper{Value: "hello"}
	result := StripInterfaceValue(w)
	assert.Equal(t, w, result)
}

func TestStripInterfaceValue_DoublePointerToInterfaceToString(t *testing.T) {
	var s interface{} = "test"
	ps := &s
	pps := &ps
	result := StripInterfaceValue(pps)
	assert.IsType(t, (**string)(nil), result)
	assert.Equal(t, s, **result.(**string))
}

func TestStripInterface_PlainString(t *testing.T) {
	result := StripInterface(reflect.ValueOf("hello"))
	assert.Equal(t, reflect.TypeOf(""), result)
}
