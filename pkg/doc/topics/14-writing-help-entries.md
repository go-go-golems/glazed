---
Title: Writing Help Entries for Glazed
Slug: writing-help-entries
Short: Learn how to create and structure Markdown documents for the Glazed help system
Topics:
- documentation
- help system
- markdown
Commands:
- AddDocToHelpSystem
Flags:
- none
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

The Glazed help system allows you to create rich and interactive help pages for your command-line applications. These help pages are defined using Markdown files, which are then loaded into the help system at runtime.

## Markdown File Structure

Each Markdown file represents a single "section" in the help system. A section can be one of the following types:

1. **General Topic**: A general article or topic related to your application.
2. **Example**: A specific example of how to use a command or feature.
3. **Application**: A more complex use case or application of your tool, potentially involving multiple commands.
4. **Tutorial**: A step-by-step guide on how to use a specific functionality.

The structure of a Markdown file is as follows:

```yaml
---
Title: The title of the section
Slug: a-unique-slug-for-this-section
Short: A short description of the section (one or two sentences)
Topics:
- topic1
- topic2
Commands:
- command1
- command2
Flags:
- flag1
- flag2
IsTopLevel: true # Whether this section should be shown in the top-level help
IsTemplate: false # Whether this section is a template for other sections
ShowPerDefault: true # Whether this section should be shown by default
SectionType: GeneralTopic # The type of the section
---

# The main content of the section

This is where you can write the full Markdown content for the section.
```

Let's go through each of the fields in the YAML frontmatter:

1. **Title**: The title of the section, which will be displayed in the help output.
2. **Slug**: A unique identifier for the section, used to reference it internally.
3. **Short**: A short description of the section, typically one or two sentences.
4. **Topics**: A list of topics that this section is related to. This is used for filtering and grouping sections.
5. **Commands**: A list of commands that this section is related to. This is used for filtering and grouping sections.
6. **Flags**: A list of flags that this section is related to. This is used for filtering and grouping sections.
7. **IsTopLevel**: Whether this section should be shown in the top-level help output.
8. **IsTemplate**: Whether this section is a template for other sections (e.g., a reusable example).
9. **ShowPerDefault**: Whether this section should be shown by default in the help output.
10. **SectionType**: The type of the section (GeneralTopic, Example, Application, or Tutorial).

After the YAML frontmatter, you can write the full Markdown content for the section. This content will be displayed in the help output.

## Organizing Sections

You can organize your sections by placing them in different directories within your codebase. The Glazed help system will automatically load all Markdown files from the specified directory and its subdirectories.

For example, you could have the following directory structure:

```
docs/
  topics/
    01-introduction.md
    02-usage.md
  examples/
    01-simple-command.md
    02-advanced-usage.md
  applications/
    01-integrating-with-external-tool.md
  tutorials/
    01-getting-started.md
    02-advanced-features.md
```

In this example, the `docs` directory contains all the help sections, organized into different subdirectories based on the section type.

## Loading Sections into the Help System

To load the Markdown sections into the Glazed help system, you can use the `AddDocToHelpSystem` function provided in the `glazed/pkg/doc` package:

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/doc"
    "github.com/go-go-golems/glazed/pkg/help"
)

func main() {
    helpSystem := help.NewHelpSystem()
    err := doc.AddDocToHelpSystem(helpSystem)
    if err != nil {
        // Handle error
    }

    // Use the helpSystem in your application
}
```

This function will recursively load all Markdown files from the `docs` directory (or any other directory you specify) and add them to the `HelpSystem`.

## Accessing Help Sections

Once the help sections are loaded, you can access them using the `HelpSystem` API. For example, you can retrieve a specific section by its slug:

```go
section, err := helpSystem.GetSectionWithSlug("a-simple-help-system")
if err != nil {
    // Handle error
}

// Use the section information
fmt.Println(section.Title)
fmt.Println(section.Short)
fmt.Println(section.Content)
```

You can also use the `SectionQuery` type to filter and retrieve sections based on various criteria, such as section type, topics, commands, or flags.

By following this guide, you can create rich and interactive help pages for your Glazed-based command-line applications, making it easier for users to understand and use your tool.