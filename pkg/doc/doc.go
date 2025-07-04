package doc

import (
	"context"
	"embed"
	"github.com/go-go-golems/glazed/pkg/help"
)

//go:embed *
var docFS embed.FS

func AddDocToHelpSystem(ctx context.Context, helpSystem *help.HelpSystem) error {
	return helpSystem.LoadSectionsFromFS(ctx, docFS, ".")
}
