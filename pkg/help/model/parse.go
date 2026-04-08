package model

import (
	"bytes"

	"github.com/adrg/frontmatter"
	strings2 "github.com/go-go-golems/glazed/pkg/helpers/strings"
	"github.com/pkg/errors"
)

// ParseSectionFromMarkdown parses YAML frontmatter + markdown body into a Section.
// This is the canonical parser — all callers should use this instead of rolling their own.
func ParseSectionFromMarkdown(markdownBytes []byte) (*Section, error) {
	var metaData map[string]interface{}

	inputReader := bytes.NewReader(markdownBytes)
	rest, err := frontmatter.Parse(inputReader, &metaData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse frontmatter")
	}

	section := &Section{}

	// Core text fields
	if title, ok := metaData["Title"]; ok {
		section.Title = title.(string)
	}
	if subTitle, ok := metaData["SubTitle"]; ok {
		section.SubTitle = subTitle.(string)
	}
	if short, ok := metaData["Short"]; ok {
		section.Short = short.(string)
	}

	// Section type — defaults to GeneralTopic
	if sectionType, ok := metaData["SectionType"]; ok {
		section.SectionType, err = SectionTypeFromString(sectionType.(string))
		if err != nil {
			return nil, errors.Wrap(err, "invalid section type")
		}
	} else {
		section.SectionType = SectionGeneralTopic
	}

	// Slug
	if slug, ok := metaData["Slug"]; ok {
		section.Slug = slug.(string)
	}

	// Content (body after frontmatter)
	section.Content = string(rest)

	// Arrays — always initialize to empty slices
	if topics, ok := metaData["Topics"]; ok {
		section.Topics = strings2.InterfaceToStringList(topics)
	} else {
		section.Topics = []string{}
	}
	if flags, ok := metaData["Flags"]; ok {
		section.Flags = strings2.InterfaceToStringList(flags)
	} else {
		section.Flags = []string{}
	}
	if commands, ok := metaData["Commands"]; ok {
		section.Commands = strings2.InterfaceToStringList(commands)
	} else {
		section.Commands = []string{}
	}

	// Boolean flags
	if isTopLevel, ok := metaData["IsTopLevel"]; ok {
		section.IsTopLevel = isTopLevel.(bool)
	}
	if isTemplate, ok := metaData["IsTemplate"]; ok {
		section.IsTemplate = isTemplate.(bool)
	}
	if showPerDefault, ok := metaData["ShowPerDefault"]; ok {
		section.ShowPerDefault = showPerDefault.(bool)
	}

	// Order — YAML may parse integers as float64
	if order, ok := metaData["Order"]; ok {
		switch v := order.(type) {
		case int:
			section.Order = v
		case float64:
			section.Order = int(v)
		}
	}

	// Validate required fields
	if section.Slug == "" || section.Title == "" {
		return nil, errors.New("missing required fields: slug and title")
	}

	return section, nil
}
