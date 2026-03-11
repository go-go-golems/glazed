package help

import (
	"context"
	"errors"
	"sort"
	"strings"
)

var ErrPageNotComposed = errors.New("help page not composed")

// PageComposer can assemble a richer runtime page for a slug using the underlying HelpSystem.
type PageComposer interface {
	ComposePage(ctx context.Context, slug string, hs *HelpSystem) (*ComposedPage, error)
}

// PagePart is one ordered portion of a composed help page.
type PagePart interface {
	Kind() string
	Order() int
	RenderMarkdown(ctx context.Context) (string, error)
}

// ComposedPage is a runtime-only help page assembled from ordered parts.
type ComposedPage struct {
	Slug  string
	Title string
	Parts []PagePart
}

// RenderMarkdown renders all parts into a single markdown document.
func (p *ComposedPage) RenderMarkdown(ctx context.Context) (string, error) {
	if p == nil {
		return "", ErrPageNotComposed
	}

	parts := append([]PagePart(nil), p.Parts...)
	sort.SliceStable(parts, func(i, j int) bool {
		return parts[i].Order() < parts[j].Order()
	})

	var sb strings.Builder
	for _, part := range parts {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		markdown, err := part.RenderMarkdown(ctx)
		if err != nil {
			return "", err
		}
		markdown = strings.TrimSpace(markdown)
		if markdown == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(markdown)
	}
	return sb.String(), nil
}
