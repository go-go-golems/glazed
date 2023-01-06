package main

import (
	"encoding/json"
	"fmt"
	"github.com/piprate/json-gold/ld"
)

func main() {
	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	expanded, err := proc.Expand("http://json-ld.org/test-suite/tests/expand-0002-in.jsonld", options)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", expanded)

	s, err := json.MarshalIndent(expanded, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", s)

	doc := map[string]interface{}{
		"@context":  "http://schema.org/",
		"@type":     "Person",
		"name":      "Jane Doe",
		"jobTitle":  "Professor",
		"telephone": "(425) 123-4567",
		"url":       "http://www.janedoe.com",
	}

	expanded, err = proc.Expand(doc, options)
	if err != nil {
		panic(err)
	}

	s, err = json.MarshalIndent(expanded, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", s)

	compacted, err := proc.Compact(expanded, "http://schema.org/", options)
	if err != nil {
		panic(err)
	}

	s, err = json.MarshalIndent(compacted, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", s)

	myContext := map[string]interface{}{
		"myJobTitle":  "http://schema.org/jobTitle",
		"myName":      "http://schema.org/name",
		"myUrl":       "http://schema.org/url",
		"myTelephone": "http://schema.org/telephone",
	}
	compacted, err = proc.Compact(expanded, myContext, options)
	if err != nil {
		panic(err)
	}

	s, err = json.MarshalIndent(compacted, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", s)
}
