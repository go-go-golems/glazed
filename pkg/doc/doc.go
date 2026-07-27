package doc

import (
	"embed"
	"io/fs"

	"github.com/go-go-golems/glazed/pkg/help"
)

//go:embed *
var docFS embed.FS

// FS returns the embedded documentation tree (topics/, tutorials/,
// applications/, examples/ markdown files with help-section frontmatter).
// External consumers such as documentation servers can walk it directly
// instead of going through a HelpSystem.
func FS() fs.FS {
	return docFS
}

func AddDocToHelpSystem(helpSystem *help.HelpSystem) error {
	return helpSystem.LoadSectionsFromFS(docFS, ".")
}
