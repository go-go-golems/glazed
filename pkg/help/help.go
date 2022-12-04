package help

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"text/template"
)

type Section struct {
	Title       string
	Slug        string
	Content     string
	SubSections []*Section
	// TODO(manuel, 2022-12-03) tags should be a hash map really
	Tags           []string
	IsTemplate     bool
	ShowPerDefault bool
	Order          int
}

func (s *Section) AddSubSection(subSection *Section) {
	s.SubSections = append(s.SubSections, subSection)
}

type RenderContext struct {
	Depth int
	Tags  []string
	Data  interface{}
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
		err := t.Execute(&titleBuffer, rc.Data)
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
		err := t.Execute(w, rc.Data)
		if err != nil {
			return err
		}
	} else {
		_, err := w.Write([]byte(s.Content))
		if err != nil {
			return err
		}
	}

	// TODO(manuel, 2022-12-03) subsections should be sorted by order
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

func (s *Section) IsTagged(tag string) bool {
	for _, t := range s.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (s *Section) IsTaggedWithAny(tags []string) bool {
	for _, t := range tags {
		if s.IsTagged(t) {
			return true
		}
	}
	return false
}

type HelpError int

const (
	ErrSectionNotFound HelpError = iota
)

func (e HelpError) Error() string {
	switch e {
	case ErrSectionNotFound:
		return "Section not found"
	default:
		return "Unknown error"
	}
}

func FindSection(sections []*Section, args []string) (*Section, error) {
	if len(args) == 0 {
		return nil, errors.Wrap(ErrSectionNotFound, "No sections available")
	}

	for _, section := range sections {
		if section.Slug == args[0] {
			if len(args) == 1 {
				return section, nil
			} else {
				return FindSection(section.SubSections, args[1:])
			}
		}
	}

	return nil, errors.Wrap(ErrSectionNotFound, fmt.Sprintf("Section %s not found", args[0]))

}

func FindSectionWithTags(sections []*Section, tags []string) []*Section {
	var result []*Section
	for _, section := range sections {
		if section.IsTaggedWithAny(tags) {
			result = append(result, section)
		}
		for _, subSection := range section.SubSections {
			result = append(result, FindSectionWithTags([]*Section{subSection}, tags)...)
		}
	}
	return result
}
