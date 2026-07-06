//go:build ignore
// +build ignore

// 02-visible-fields-probe.go
//
// Purpose: Check whether reflect.VisibleFields is the right tool to fix the
// P2 review comment on PR #599 ("Skip embedded fields that are hidden by outer
// fields"). VisibleFields must (a) INCLUDE promoted fields from unexported
// embedded types (the issue #597 reproduction) and (b) EXCLUDE shadowed fields
// (the reviewer's regression case).
//
// Run: go run 02-visible-fields-probe.go
//
// Key findings (go1.25.5 linux/amd64):
//   - ServeSettings (unexported value embed): VisibleFields returns
//       commonSettings idx=[0]   anonymous=true  (the embed itself)
//       DB             idx=[0 0] tag="db"        (PROMOTED from unexported embed) ✓
//       Listen         idx=[1]   tag="listen"
//     So VisibleFields DOES include promoted fields from unexported embedded
//     types — switching to VisibleFields will NOT regress the issue's case.
//   - Outer (shadowing: direct Port string + embedded Common.Port int):
//       Port   idx=[0] tag="port"  (the outer direct field ONLY) ✓
//       Common idx=[1] anonymous=true
//     VisibleFields EXCLUDES the shadowed Common.Port int — exactly what the
//     reviewer asked for.
//
// Conclusion: iterate reflect.VisibleFields(t) instead of NumField(), and
// access each tagged leaf field via FieldByIndex(field.Index). This honors Go
// promotion/shadowing rules for free. (See 03-... for the FieldByIndex nil
// pointer caveat.)
package main

import (
	"fmt"
	"reflect"
)

// Issue #597 reproduction: unexported embedded type, exported promoted field
type commonSettings struct {
	DB string `glazed:"db"`
}
type ServeSettings struct {
	commonSettings
	Listen string `glazed:"listen"`
}

// Reviewer's shadowing case: outer direct field shadows embedded field
type Common struct {
	Port int `glazed:"port"`
}
type Outer struct {
	Port string `glazed:"port"`
	Common
}

func dump(label string, t reflect.Type) {
	fmt.Println("==", label, "==")
	vf := reflect.VisibleFields(t)
	for _, f := range vf {
		fmt.Printf("  visible: %-12s idx=%v tag=%q exported=%v anonymous=%v\n", f.Name, f.Index, f.Tag.Get("glazed"), f.IsExported(), f.Anonymous)
	}
	// Does FieldByName find the shadowed embedded Port?
	v := reflect.New(t).Elem()
	p := v.FieldByName("Port")
	fmt.Printf("  FieldByName(Port): valid=%v kind=%v settable=%v\n", p.IsValid(), p.Kind(), p.CanSet())
	// Try FieldByName(DB) on ServeSettings
}

func main() {
	dump("ServeSettings (unexported value embed)", reflect.TypeOf(ServeSettings{}))
	dump("Outer (shadowing)", reflect.TypeOf(Outer{}))
}
