//go:build ignore
// +build ignore

// 03-fieldbyindex-nil-pointer-probe.go
//
// Purpose: Determine how reflect.Value.FieldByIndex behaves when the index
// chain crosses a NIL embedded pointer-to-struct (the pointer-embed allocation
// case). This decides whether a VisibleFields-based rewrite can rely on
// FieldByIndex, or needs a custom allocator for nil pointer intermediates.
//
// Run: go run 03-fieldbyindex-nil-pointer-probe.go
//
// Key findings (go1.25.5 linux/amd64):
//   - VisibleFields for ServePtr{*CommonSettings; Listen} returns:
//       CommonSettings idx=[0]   anonymous=true
//       DB             idx=[0 0] tag="db"
//       Listen         idx=[1]   tag="listen"
//   - v.FieldByIndex([]int{0, 0}) on a ServePtr whose *CommonSettings is nil
//     PANICS: "reflect: indirection through nil pointer to embedded struct".
//
// Conclusion: FieldByIndex does NOT allocate nil pointer intermediates, so a
// VisibleFields-based DecodeInto must walk field.Index manually, allocating
// settable (exported) nil pointers before descending. Unexported nil
// pointer-embeds still cannot be allocated and must be skipped (as in the
// current decodeEmbedded).
package main

import (
	"fmt"
	"reflect"
)

type CommonSettings struct {
	DB string `glazed:"db"`
}
type ServePtr struct {
	*CommonSettings
	Listen string `glazed:"listen"`
}

func main() {
	s := &ServePtr{}
	v := reflect.ValueOf(s).Elem()
	vf := reflect.VisibleFields(v.Type())
	for _, f := range vf {
		fmt.Printf("visible %-14s idx=%v anon=%v tag=%q\n", f.Name, f.Index, f.Anonymous, f.Tag.Get("glazed"))
	}
	// try FieldByIndex on DB ([0 0]) with nil pointer embed
	db := v.FieldByIndex([]int{0, 0})
	fmt.Printf("FieldByIndex([0,0]) DB: valid=%v kind=%v settable=%v\n", db.IsValid(), db.Kind(), db.CanSet())
}
