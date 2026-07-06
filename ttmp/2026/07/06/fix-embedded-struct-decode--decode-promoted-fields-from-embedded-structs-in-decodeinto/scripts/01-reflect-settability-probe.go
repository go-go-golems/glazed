//go:build ignore
// +build ignore

// 01-reflect-settability-probe.go
//
// Purpose: Determine what reflect allows when decoding into embedded structs of
// (a) unexported and (b) exported type, for both value- and pointer-embeds.
// Run during Step 2 of ticket fix-embedded-struct-decode (issue #597) after the
// first test panicked with "reflect.Value.Interface: cannot return value
// obtained from unexported field".
//
// Run: go run 01-reflect-settability-probe.go
//
// Key findings (go1.25.5 linux/amd64):
//   - Unexported value-embed (type commonSettings): the embedded field itself is
//     CanSet=false / CanAddr=true, BUT v.FieldByName("DB") returns the promoted
//     exported field with CanSet=true and SetString works. So decoding into
//     promoted exported fields is possible even when the embedded TYPE is
//     unexported — as long as we recurse via reflect.Value (not .Interface()).
//   - Exported value-embed (CommonExported): embedded field is CanSet=true and
//     CanInterface=true (the easy case).
//   - Exported pointer-embed, nil (*CommonExported): ptr.CanSet=true, so it can
//     be allocated via ptr.Set(reflect.New(...)) before dereferencing.
//
// Conclusion: recurse via the reflect.Value directly; never call
// .Interface() on an embedded field (panics for unexported embedded types).
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

type CommonExported struct {
	DB string `glazed:"db"`
}

type ServeExported struct {
	CommonExported
	Listen string `glazed:"listen"`
}

type ServePtr struct {
	*CommonExported
	Listen string `glazed:"listen"`
}

func main() {
	// value embed, unexported type
	s := &ServeSettings{}
	v := reflect.ValueOf(s).Elem()
	t := v.Type()
	fmt.Println("== ServeSettings (unexported value embed) ==")
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)
		fmt.Printf("  field %q anonymous=%v settable=%v canaddr=%v\n", f.Name, f.Anonymous, fv.CanSet(), fv.CanAddr())
	}
	db := v.FieldByName("DB")
	fmt.Printf("  FieldByName(DB): valid=%v settable=%v canaddr=%v\n", db.IsValid(), db.CanSet(), db.CanAddr())
	// try set
	if db.CanSet() {
		db.SetString("/tmp/x.db")
		fmt.Println("  set DB via FieldByName OK ->", s.DB)
	}

	// exported value embed
	s2 := &ServeExported{}
	v2 := reflect.ValueOf(s2).Elem()
	emb := v2.Field(0)
	fmt.Println("== ServeExported (exported value embed) ==")
	fmt.Printf("  embedded settable=%v canaddr=%v\n", emb.CanSet(), emb.CanAddr())
	// can we Interface()?
	fmt.Printf("  embedded CanInterface=%v\n", emb.CanInterface())

	// pointer embed exported, nil
	s3 := &ServePtr{}
	v3 := reflect.ValueOf(s3).Elem()
	ptr := v3.Field(0)
	fmt.Println("== ServePtr (exported ptr embed, nil) ==")
	fmt.Printf("  ptr settable=%v canaddr=%v isnil=%v\n", ptr.CanSet(), ptr.CanAddr(), ptr.IsNil())
	if ptr.CanSet() && ptr.IsNil() {
		ptr.Set(reflect.New(ptr.Type().Elem()))
		fmt.Println("  allocated ptr OK")
	}
}
