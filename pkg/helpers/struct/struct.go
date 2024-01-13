package _struct

import (
	reflect "reflect"
)

func CloneStruct(src interface{}) interface{} {
	srcValue := reflect.ValueOf(src)
	srcType := srcValue.Type()

	// Check if the source is a pointer and get the value it points to
	if srcType.Kind() == reflect.Ptr {
		srcValue = srcValue.Elem()
		srcType = srcValue.Type()
	}

	// Return nil if the source is not a struct
	if srcType.Kind() != reflect.Struct {
		return nil
	}

	// Create a new instance of the source type
	dstValue := reflect.New(srcType).Elem()

	for i := 0; i < srcValue.NumField(); i++ {
		srcField := srcValue.Field(i)
		dstField := dstValue.Field(i)

		// Check if the field can be set
		if dstField.CanSet() {
			// If the field is a pointer and not nil, clone it recursively
			if srcField.Kind() == reflect.Ptr && !srcField.IsNil() {
				// Create a new pointer of the same type as the source field
				newPtr := reflect.New(srcField.Type().Elem())
				// Clone the value pointed to by the source field
				newPtr.Elem().Set(reflect.ValueOf(CloneStruct(srcField.Interface())))
				// Set the cloned pointer to the destination field
				dstField.Set(newPtr)
			} else {
				dstField.Set(srcField)
			}
		}
	}

	// Return the cloned struct as an interface{}
	return dstValue.Interface()
}
