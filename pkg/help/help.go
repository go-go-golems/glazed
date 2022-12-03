package help

import (
	"bytes"
	"fmt"
	"io"
	"text/template"
)

type Section struct {
	Title          string
	Slug           string
	Content        string
	SubSections    []*Section
	Tags           []string
	IsTemplate     bool
	ShowPerDefault bool
}

func (s *Section) AddSubSection(subSection *Section) {
	s.SubSections = append(s.SubSections, subSection)
}

type RenderContext struct {
	Depth int
	Tags  []string
	Date  interface{}
}

func (rc *RenderContext) AddTag(tag string) *RenderContext {
	newrc := &RenderContext{
		Depth: rc.Depth,
		Tags:  append(rc.Tags, tag),
	}
	return newrc
}

func (rc *RenderContext) IsTaggedWithAny(tags []string) bool {
	for _, t := range tags {
		if rc.IsTagged(t) {
			return true
		}
	}
	return false
}

func (rc *RenderContext) IsTagged(tag string) bool {
	return rc.IsTaggedWithAny([]string{tag})
}

func (rc *RenderContext) IncreaseDepth() *RenderContext {
	newrc := &RenderContext{
		Depth: rc.Depth + 1,
		Tags:  rc.Tags,
	}
	return newrc
}

func makeMarkdownHeader(depth int, title string) string {
	return fmt.Sprintf("%s %s\n\n", bytes.Repeat([]byte("#"), depth), title)
}

type ContentSection interface {
	Render(w io.Writer, rc *RenderContext) error
}

func (s *Section) Render(w io.Writer, rc *RenderContext) error {
	renderedTitle := s.Title
	if s.IsTemplate {
		t := template.New("title")
		template.Must(t.Parse(s.Title))
		var titleBuffer bytes.Buffer
		err := t.Execute(&titleBuffer, rc.Date)
		if err != nil {
			return err
		}
		renderedTitle = titleBuffer.String()
	}
	_, err := w.Write([]byte(makeMarkdownHeader(rc.Depth, renderedTitle)))
	if err != nil {
		return err
	}

	if s.IsTemplate {
		t := template.New("content")
		template.Must(t.Parse(s.Content))
		err := t.Execute(w, rc.Date)
		if err != nil {
			return err
		}
	} else {
		_, err := w.Write([]byte(s.Content))
		if err != nil {
			return err
		}
	}

	for _, subSection := range s.SubSections {
		if subSection.ShowPerDefault || rc.IsTaggedWithAny(subSection.Tags) {
			err := subSection.Render(w, rc.IncreaseDepth())
			if err != nil {
				return err
			}
		}
	}

	return nil
}
