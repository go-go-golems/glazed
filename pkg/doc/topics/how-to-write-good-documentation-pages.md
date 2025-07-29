---
Title: How to Write Good Documentation Pages
Slug: how-to-write-good-documentation-pages
Short: Style guide and best practices for creating clear, consistent, and helpful documentation.
Topics:
- documentation
- style-guide
- writing
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Documentation Style Guide

## 1. Overview

This document provides a set of standards and best practices for writing clear, consistent, and helpful documentation for the Glazed project. The goal is to ensure that all documentation, whether written by humans or AI, is easy to understand, accurate, and useful for our target audience of developers.

Following these guidelines will help us create a cohesive and professional documentation suite that empowers users to learn and succeed with the Glazed framework.

## 2. Core Principles

All documentation should adhere to these fundamental principles:

-   **Clarity**: Write in simple, direct, and unambiguous language. Avoid jargon where possible, or explain it clearly if it's necessary.
-   **Accuracy**: Ensure all information, especially code examples and technical specifications, is correct, tested, and up-to-date.
-   **Conciseness**: Be direct and to the point. Eliminate wordiness and focus on delivering information efficiently.
-   **Completeness**: Cover the topic thoroughly, but avoid going off-topic. Anticipate the user's questions and answer them proactively.
-   **Audience-Centric**: Always write with the developer-user in mind. What is their goal? What problem are they trying to solve? Frame your explanation in that context.

## 3. Document Structure

A well-structured document is easy to navigate and digest. Every documentation page should follow this general structure.

### 3.1. YAML Front Matter

All documents must begin with a YAML front matter block that provides metadata for the help system.

```yaml
---
Title: Glazed Commands Reference
Slug: commands-reference
Short: Complete reference for creating, configuring, and running commands in Glazed
Topics:
- commands
- interfaces
- flags
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---
```

### 3.2. Main Title

The `H1` title should match the `Title` field in the front matter.

### 3.3. Section Introduction Paragraphs

Every major `H2` section must begin with a single, topic-focused paragraph. This paragraph should **explain the core concept** of the section, not just describe what the section contains. Its purpose is to give the reader immediate context and understanding of the "what" and "why" before they dive into the details.

**Example: How to write a good introductory paragraph**

-   **Avoid (Simple Section Overview):**
    > This section will cover the command interfaces. It will explain what `BareCommand`, `WriterCommand`, and `GlazeCommand` are and provide examples for each one.

-   **Prefer (Topic-Focused Explanation):**
    > A command's output contract is defined by the interface it implements. Glazed offers three primary interfaces to support different use cases: `BareCommand` for direct `stdout` control, `WriterCommand` for sending text to any `io.Writer`, and `GlazeCommand` for producing structured data that can be automatically formatted. This design allows a command's business logic to be decoupled from its final output format.

The second example is much better because it immediately teaches the reader about the core design principle (decoupling logic from output) and the purpose of the interfaces, rather than just stating what the section is about.

### 3.4. Headings and Content

-   Use `##` (H2) for major sections and `###` (H3) for sub-sections.
-   Keep paragraphs short and focused on a single idea.
-   Use bulleted lists to break up long blocks of text and present information in a scannable format.

### 3.5. Code Blocks

Code examples are the most critical part of developer documentation.

-   **Keep them minimal and focused:** An example should demonstrate exactly one concept. Remove all boilerplate or irrelevant logic.
-   **Use comments to explain the *why*, not the *what*:**
    -   **Bad:** `// Create a new row`
    -   **Good:** `// Use types.MRP to ensure type-safe key-value pairs`
-   **Ensure they are runnable:** Whenever possible, code examples should be copy-paste-runnable.
-   **Show the output:** For commands or code that produces output, show the expected result in a separate block or as a comment.

```bash
# Example command
./my-app list-users --output json

# Expected output
[
  {
    "id": 1,
    "name": "Alice"
  }
]
```

### 3.6. Linking Between Documents

Always use the `glaze help` format for internal links. This ensures that the documentation is accessible both on the web and from the command line.

-   **Do:**
    > For more details, see the layers guide:
    > ```
    > glaze help layers-guide
    > ```

-   **Don't:**
    > For more details, see the [layers guide](./layers-guide.md).

## 4. Writing Style and Tone

-   **Audience:** Assume the reader is a developer familiar with Go but new to the Glazed framework.
-   **Tone:** Professional, helpful, and direct.
-   **Voice:** Use the active voice. It is more direct and easier to read.
    -   **Good:** "The framework provides three interfaces."
    -   **Bad:** "Three interfaces are provided by the framework."
-   **Terminology:** Use consistent names for concepts. For example, always refer to them as "Parameter Layers," not "Layers of Parameters" or "Parameter Groups."

## 5. A Complete Example

Here is a full "before and after" of a section being improved to meet these guidelines.

### Before (Lacks context, examples are too long)

```markdown
## Command Interfaces

This section describes the command interfaces in Glazed.

### GlazeCommand

The `GlazeCommand` is for commands that output structured data.

```go
type MonitorServersCommand struct {
    *cmds.CommandDescription
}

func (c *MonitorServersCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    s := &MonitorSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    // Get server data from various sources
    servers := getServersFromInventory(s.Environment)
    
    for _, server := range servers {
        // Check server health
        health := checkServerHealth(server.Hostname)
        
        // Produce a rich data row with nested information
        row := types.NewRow(
            types.MRP("hostname", server.Hostname),
            types.MRP("environment", server.Environment),
            types.MRP("cpu_percent", health.CPUPercent),
            // ... and 10 more fields ...
        )
        
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}
```
```

### After (Adds a topic-focused intro, trims the example)

```markdown
## Command Interfaces

A command's output contract is defined by the interface it implements. Glazed offers three primary interfaces to support different use cases: `BareCommand` for direct `stdout` control, `WriterCommand` for sending text to any `io.Writer`, and `GlazeCommand` for producing structured data that can be automatically formatted. This design allows a command's business logic to be decoupled from its final output format.

### GlazeCommand

The `GlazeCommand` interface is for commands that produce structured data. By yielding `types.Row` objects to a `Processor`, your command can automatically support multiple output formats (JSON, YAML, CSV, etc.) without any format-specific code.

**Use cases:**
- API clients that fetch and display data
- Commands that query a database
- Any tool whose output is meant to be piped or consumed by other scripts

**Example implementation:**
```go
type ListUsersCommand struct {
    *cmds.CommandDescription
}

func (c *ListUsersCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Business logic to fetch users would go here
    users := []User{
        {ID: 1, Name: "Alice", Email: "alice@example.com"},
        {ID: 2, Name: "Bob", Email: "bob@example.com"},
    }

    // Instead of printing, create structured rows
    for _, user := range users {
        row := types.NewRow(
            types.MRP("id", user.ID),
            types.MRP("name", user.Name),
            types.MRP("email", user.Email),
        )
        
        // The processor handles filtering, formatting, and output
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}
```
``` 