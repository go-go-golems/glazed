package dsl_test

import (
	"fmt"
	"log"

	"github.com/go-go-golems/glazed/pkg/help/dsl"
)

// ExampleParseQuery demonstrates basic usage of the DSL parser
func ExampleParseQuery() {
	// Simple field query
	predicate, err := dsl.ParseQuery("type:example")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parsed: type:example\n")
	_ = predicate

	// Boolean operations
	predicate, err = dsl.ParseQuery("examples AND topic:database")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parsed: examples AND topic:database\n")
	_ = predicate

	// Complex query with grouping
	predicate, err = dsl.ParseQuery("(examples OR tutorials) AND topic:database")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parsed: (examples OR tutorials) AND topic:database\n")
	_ = predicate

	// Text search
	predicate, err = dsl.ParseQuery("\"full text search\"")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Parsed: \"full text search\"\n")
	_ = predicate

	// Output:
	// Parsed: type:example
	// Parsed: examples AND topic:database
	// Parsed: (examples OR tutorials) AND topic:database
	// Parsed: "full text search"
}

// ExampleGetQueryInfo demonstrates getting information about the DSL
func ExampleGetQueryInfo() {
	info := dsl.GetQueryInfo()

	fmt.Printf("Valid fields count: %d\n", len(info.ValidFields))
	fmt.Printf("Valid types: %v\n", info.ValidTypes)
	fmt.Printf("Valid shortcuts: %v\n", info.ValidShortcuts)
	fmt.Printf("Example queries: %v\n", info.Examples[:3]) // Show first 3 examples

	// Output:
	// Valid fields count: 8
	// Valid types: [example tutorial topic application]
	// Valid shortcuts: [examples tutorials topics applications toplevel defaults templates]
	// Example queries: [examples type:example topic:database]
}
