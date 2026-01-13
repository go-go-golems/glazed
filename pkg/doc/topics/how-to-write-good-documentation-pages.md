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

## Why This Matters

Bad documentation wastes developer time. When a developer reads your doc and still doesn't understand how to accomplish their goal, they:
- Dig through source code (your job, not theirs)
- Ask questions in Slack/Discord (your time, not theirs)
- Give up and use a different library

Good documentation respects the reader's time by being **immediately useful**. Every sentence should either teach something or help the reader navigate to what they need.

## Quick Reference

| Rule | Why |
|------|-----|
| Lead with "Why" | Readers need motivation before details |
| One concept per section | Cognitive load kills comprehension |
| Code examples must be runnable | Broken examples destroy trust |
| Comments explain *why*, not *what* | "Create user" is obvious; "Validate before insert" isn't |
| End with troubleshooting table | Readers often arrive via error messages |
| Add "See Also" cross-references | Help readers find related content |
| Tables for comparisons, bullets for lists | Tables are scannable; bullets are readable |

## Avoiding Terse Documentation

Terse docs are the most common failure mode. They assume the reader already knows what you know. They don't.

### The Problem

Terse documentation looks like this:

```markdown
## Tool Registry

Use `toolcontext.WithRegistry(ctx, registry)` to attach tools.
```

This tells the reader *what* to do but not:
- **Why** they need to do it
- **When** to do it (before or after something else?)
- **What happens** if they don't
- **What it connects to** in the bigger picture

### The Fix: Expand Every Section

For each concept, answer these questions:

1. **What is it?** — One sentence definition
2. **Why does it exist?** — The problem it solves
3. **When do you use it?** — The triggering situation
4. **How do you use it?** — Code example
5. **What happens if you don't?** — The failure mode
6. **What's related?** — Links to connected concepts

### Before and After

**Before (Terse):**

```markdown
## Tool Registry

Use `toolcontext.WithRegistry(ctx, registry)` to attach tools.
```

**After (Complete):**

```markdown
## Tool Registry

The tool registry is an in-memory store of callable tools. Engines read from the registry to know which tools to advertise to the model.

**Why it exists:** Tools contain function pointers, which aren't serializable. Keeping the registry in `context.Context` separates runtime state from persistable Turn data.

**When to use it:** Before calling `RunInference` or `RunToolCallingLoop`, attach the registry to your context:

```go
ctx = toolcontext.WithRegistry(ctx, registry)
```

**What happens if you skip this:** The engine won't advertise any tools to the model, so tool calls will never be requested.

**See also:** [Tools](07-tools.md), [Turns](08-turns.md)
```

The expanded version is 10x more useful because it answers the questions readers actually have.

### Expansion Checklist

Before publishing any section, verify:

- [ ] First paragraph explains *what* and *why*, not just *what*
- [ ] At least one code example with context
- [ ] Failure mode documented ("What happens if...")
- [ ] Cross-references to related concepts
- [ ] No undefined jargon (or jargon is explained on first use)

## Document Types

Choose the right format for your content:

### Topics (Reference)

**Purpose:** Explain a concept thoroughly. Reader might read end-to-end or jump to a section.

**Structure:**
```markdown
# Concept Name

## Why [Concept]?
Motivation and problem solved.

## Core Concepts
Key abstractions and how they relate.

## Usage
How to use it with examples.

## Configuration
Options, flags, settings.

## Troubleshooting
Common problems and solutions.

## See Also
Related docs.
```

**Example:** [Events](04-events.md), [Turns](08-turns.md)

### Tutorials (Learning)

**Purpose:** Teach by building. Reader follows start-to-finish.

**Structure:**
```markdown
# Build [Thing]

## What You'll Build
Concrete outcome.

## Prerequisites
What reader needs before starting.

## Step 1 — [First Action]
Explanation + code.

## Step 2 — [Second Action]
Explanation + code.

...

## Complete Example
Full working code.

## Troubleshooting
Common problems during the tutorial.

## See Also
Where to go next.
```

**Example:** [Streaming Tutorial](../tutorials/01-streaming-inference-with-tools.md)

### Playbooks (Operations)

**Purpose:** Step-by-step checklist for a specific task. Reader wants to accomplish something now.

**Structure:**
```markdown
# [Task Name]

## Prerequisites
What's needed before starting.

## Steps

### Step 1: [Action]
Minimal explanation + code.

### Step 2: [Action]
Minimal explanation + code.

...

## Complete Example
Working code combining all steps.

## Troubleshooting
| Problem | Cause | Solution |

## See Also
Related docs.
```

**Example:** [Add a New Tool](../playbooks/01-add-a-new-tool.md)

### When to Use Each

| Reader's Question | Document Type |
|-------------------|---------------|
| "What is X and how does it work?" | Topic |
| "How do I learn to use X?" | Tutorial |
| "How do I do X right now?" | Playbook |

## Section Introductions

Every `##` section must start with a paragraph that **explains the concept**, not just describes the section.

### Bad (Meta-description)

> This section covers the event system. It explains event types, routing, and handlers.

This tells the reader what the section *contains* but teaches nothing.

### Good (Concept-focused)

> Events flow from engines through a Watermill-backed router to your handlers. Each token, tool call, and error becomes a structured event you can log, display, or aggregate. This enables responsive UIs that show results as they stream in, rather than after a 10-second wait.

This teaches the reader *what events are* and *why they matter* before diving into details.

### Template

Use this pattern:

> [What it is] + [How it works at a high level] + [Why you'd care / what it enables]

## Code Examples

Code examples are the most important part of developer documentation. Get them right.

### Keep Examples Minimal

Remove everything that isn't essential to the concept:

**Bad:**
```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/go-go-golems/geppetto/pkg/embeddings"
    "github.com/go-go-golems/glazed/pkg/cli"
    // ... 10 more imports
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()
    
    // ... 50 lines of setup
    
    embedding, err := provider.GenerateEmbedding(ctx, "Hello")
    // ... error handling, cleanup, etc.
}
```

**Good:**
```go
provider := embeddings.NewOpenAIProvider(apiKey, model, 1536)

embedding, err := provider.GenerateEmbedding(ctx, "Hello, world!")
if err != nil { panic(err) }

fmt.Printf("Generated %d-dimensional vector\n", len(embedding))
```

The good example shows only what's needed to understand the concept.

### Comments Explain *Why*, Not *What*

**Bad:**
```go
// Create a new registry
registry := tools.NewInMemoryToolRegistry()

// Register the tool
registry.RegisterTool("get_weather", *toolDef)

// Attach to context
ctx = toolcontext.WithRegistry(ctx, registry)
```

**Good:**
```go
registry := tools.NewInMemoryToolRegistry()
registry.RegisterTool("get_weather", *toolDef)

// Engines read from context, not from Turn.Data (functions aren't serializable)
ctx = toolcontext.WithRegistry(ctx, registry)
```

### Show Expected Output

When code produces output, show it:

```go
similarity := cosineSimilarity(vec1, vec2)
fmt.Printf("Similarity: %.4f\n", similarity)
// Output: Similarity: 0.9234
```

Or in a separate block:

```bash
$ go run main.go
Similarity: 0.9234
```

## Tables vs Bullets

Use the right format for your content:

### Use Tables For

- **Comparisons:** When items have multiple attributes to compare
- **Options/Config:** When documenting flags, parameters, or settings
- **Troubleshooting:** Problem/Cause/Solution format
- **Quick reference:** Scannable lookup information

```markdown
| Provider | Model | Dimensions |
|----------|-------|------------|
| OpenAI | text-embedding-3-small | 1536 |
| Ollama | all-minilm | 384 |
```

### Use Bullets For

- **Lists of items:** When items are peers without attributes to compare
- **Steps that need explanation:** When each item needs a paragraph
- **Features/benefits:** Marketing-style lists

```markdown
**Use cases:**
- Semantic search over documents
- Clustering similar texts
- Classification based on content
```

### Use Numbered Lists For

- **Sequential steps:** When order matters
- **Prioritized items:** When ranking is important

## Troubleshooting Tables

Every doc should end with a troubleshooting table. Readers often arrive via error messages.

```markdown
## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| "tool not found" error | Registry not attached | Add `ctx = toolcontext.WithRegistry(ctx, registry)` |
| No events received | Router not running | Wait for `<-router.Running()` before inference |
| Tokens not streaming | Sink not configured | Pass `engine.WithSink(sink)` when creating engine |
```

**Why this format works:**
- Problem column matches what reader is searching for
- Cause column helps reader understand
- Solution column gives actionable fix

## Cross-References

End every doc with a "See Also" section:

```markdown
## See Also

- [Related Concept](path/to/doc.md) — One-line description
- [Tutorial](path/to/tutorial.md) — What you'll build
- Example: `path/to/example/main.go`
```

**Rules:**
- Use relative markdown links for doc-to-doc references
- Include a brief description (not just the link)
- Link to examples when they exist

## Anti-Patterns

### Wall of Text

**Problem:** Long paragraphs without structure.

**Fix:** Break into sections, use bullets, add headings.

### Code-First, Explanation-Later

**Problem:** Jumping straight to code without context.

**Fix:** One paragraph explaining *what* and *why* before every code block.

### Undefined Jargon

**Problem:** Using terms without explanation.

**Bad:** "Attach the sink to the context."

**Good:** "Attach the sink (the event publisher) to the context so that helpers and tools can emit events."

### Missing Failure Modes

**Problem:** Only explaining the happy path.

**Fix:** Add "What happens if you don't..." or a troubleshooting section.

### Orphan Docs

**Problem:** Docs with no links to or from other docs.

**Fix:** Add cross-references in both directions.

## Writing Style

### Voice and Tone

- **Active voice:** "The engine processes the Turn" not "The Turn is processed by the engine"
- **Direct:** "Use `WithSink`" not "You might want to consider using `WithSink`"
- **Helpful, not formal:** Write like you're explaining to a colleague

### Terminology

Use consistent names. Pick one and stick with it:

| Use This | Not These |
|----------|-----------|
| Turn | conversation, message, request |
| Block | message, content, part |
| Engine | step, provider, client |
| Registry | store, map, collection |

### Assume the Reader

- Knows Go
- Is new to this library
- Wants to accomplish something specific
- Will skim before reading closely

## Templates

### Topic Template

```markdown
---
Title: [Concept Name]
Slug: [concept-name]
Short: [One sentence description]
Topics:
- [topic1]
- [topic2]
IsTopLevel: true
SectionType: GeneralTopic
---

# [Concept Name]

## Why [Concept]?

[2-3 sentences explaining the problem this solves and why you'd use it.]

## Quick Start

```[language]
[Minimal working example - 5-10 lines]
```

## Core Concepts

[Explanation of key abstractions with diagrams if helpful]

## [Main Section 1]

[Content]

## [Main Section 2]

[Content]

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| ... | ... | ... |

## See Also

- [Related Doc](path.md) — Description
```

### Tutorial Template

```markdown
---
Title: [Build/Create] [Thing]
Slug: [thing]-tutorial
Short: [What you'll build in one sentence]
SectionType: Tutorial
---

# [Build/Create] [Thing]

## What You'll Build

[Concrete description of the end result]

## Prerequisites

- [Requirement 1]
- [Requirement 2]

## Step 1 — [First Action]

[Explanation of what and why]

```[language]
[Code for this step]
```

## Step 2 — [Second Action]

[Explanation]

```[language]
[Code]
```

## Complete Example

```[language]
[Full working code combining all steps]
```

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| ... | ... | ... |

## See Also

- [Topic Reference](path.md)
- Example: `path/to/example/`
```

## Checklist Before Publishing

- [ ] Every section starts with a concept-focused paragraph
- [ ] All code examples are runnable
- [ ] Jargon is explained on first use
- [ ] Troubleshooting section exists
- [ ] See Also section with cross-references
- [ ] No orphan docs (linked from index or related docs)
- [ ] Failure modes documented
