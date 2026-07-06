//go:build ignore
// +build ignore

// 04-interface-through-unexported-embed-probe.go
//
// Purpose: Confirm that reflect.Value.Interface() can be called on a promoted
// field reached THROUGH an unexported embedded struct (via both FieldByIndex and
// FieldByName). This decides whether a VisibleFields-based StructToDataMap can
// use fieldByIndex + .Interface() to read promoted fields, or whether it panics
// ("value obtained from unexported field").
//
// Run: go run 04-interface-through-unexported-embed-probe.go
//
// Key findings (go1.25.5 linux/amd64):
//   - FieldByIndex([0 0]) DB: valid, kind=string, CanInterface=true;
//     Interface() returns "/tmp/x.db". OK.
//   - FieldByName("DB"): valid, CanInterface=true; Interface() returns
//     "/tmp/x.db". OK.
//
// Conclusion: a promoted EXPORTED field is interfacable even when reached
// through an unexported embedded type, so StructToDataMap can safely use
// VisibleFields + fieldByIndex(alloc=false) + .Interface(). (Nil embedded
// pointers are skipped before .Interface() is reached.)
package main

import (
	"fmt"
	"reflect"
)

type commonSettings struct {
	DB string `glazed:"db"`
}
type ServeSettings struct {
	commonSettings
	Listen string `glazed:"listen"`
}

func main() {
	s := &ServeSettings{commonSettings: commonSettings{DB: "/tmp/x.db"}, Listen: ":8080"}
	v := reflect.ValueOf(s).Elem()

	// via FieldByIndex through the unexported embed
	db1 := v.FieldByIndex([]int{0, 0})
	fmt.Printf("FieldByIndex DB: valid=%v kind=%v canInterface=%v\n", db1.IsValid(), db1.Kind(), db1.CanInterface())
	if db1.CanInterface() {
		fmt.Printf("  Interface() = %v\n", db1.Interface())
	}

	// via FieldByName
	db2 := v.FieldByName("DB")
	fmt.Printf("FieldByName DB: valid=%v canInterface=%v\n", db2.IsValid(), db2.CanInterface())
	if db2.CanInterface() {
		fmt.Printf("  Interface() = %v\n", db2.Interface())
	}
}
