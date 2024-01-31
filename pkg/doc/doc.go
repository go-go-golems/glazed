package doc

import (
	"embed"
	"github.com/go-go-golems/glazed/pkg/help"
)

//go:embed **/*.md
var docFS embed.FS

func AddDocToHelpSystem(helpSystem *help.HelpSystem) error {
	return helpSystem.LoadSectionsFromFS(docFS, ".")
}
