package helpers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
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
	}
}
