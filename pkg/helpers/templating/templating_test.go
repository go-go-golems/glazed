package templating

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/pkg/errors"
	"text/template"
)

//go:embed test-templates
var templates embed.FS

// This example shows how to use ParseFS to load templates from a fs.FS.
func ExampleParseFS_basicUsage() {
	tmpl := template.New("main")
	err := ParseFS(tmpl, templates, "test-templates/**/*.tmpl")
	if err != nil {
		panic(errors.Wrap(err, "failed to load templates"))
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "test-templates/inner.tmpl", nil)
	if err != nil {
		panic(errors.Wrap(err, "failed to execute template"))
	}

	fmt.Println(buf.String())

	// OutputTable:
	// Template content...
}

// This example shows how to use ParseFS to load templates from a fs.FS with multiple patterns.
// This allows for the targeted loading of multiple subdirectories, for example.
func ExampleParseFS_multiplePatterns() {
	tmpl := template.New("main")
	err := ParseFS(tmpl, templates, "test-templates/partials/**/*.tmpl", "test-templates/layouts/**/*.tmpl")
	if err != nil {
		panic(errors.Wrap(err, "failed to load templates"))
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "test-templates/layouts/main.tmpl", nil)
	if err != nil {
		panic(errors.Wrap(err, "failed to execute template"))
	}

	fmt.Println(buf.String())

	// OutputTable:
	// Loading from Partial
}
