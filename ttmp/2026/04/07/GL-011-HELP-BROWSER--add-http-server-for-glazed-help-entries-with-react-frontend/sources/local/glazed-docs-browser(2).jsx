import { useState, useMemo } from "react";

const SECTIONS = [
  {
    slug: "help-system",
    title: "Help System",
    short: "Glazed provides a powerful, queryable help system for creating rich CLI documentation with sections, metadata, and programmatic access.",
    type: "GeneralTopic",
    topics: ["help", "documentation", "cli", "sections", "query"],
    commands: ["help"],
    flags: ["flag", "topic", "command", "list", "topics", "examples", "applications", "tutorials", "help"],
    isTopLevel: true,
    content: `## Overview

The Glazed help system provides a structured, queryable approach to CLI documentation that goes beyond basic command help. It organizes documentation into typed sections (topics, examples, applications, tutorials) with rich metadata for filtering and discovery.

The system supports both human-readable help pages and programmatic querying through a simple DSL, making it easy to build comprehensive CLI documentation that users can explore efficiently.

The help system stores sections in an SQLite-backed store, enabling fast queries, text search, and metadata filtering.

## Section Types

- **GeneralTopic** — Conceptual documentation explaining how features work
- **Example** — Focused demonstrations of specific command usage
- **Application** — Real-world use cases combining multiple features
- **Tutorial** — Step-by-step guides for complex workflows

## Programmatic Usage

\`\`\`go
// Create new help system with in-memory storage
hs := help.NewHelpSystem()

// Load documentation from embedded filesystem
//go:embed docs
var docsFS embed.FS
err := hs.LoadSectionsFromFS(docsFS, "docs")
\`\`\`

## Query System and DSL

The help system treats documentation as structured data that can be queried using boolean logic and metadata filters.

\`\`\`go
// Query by section type
examples, _ := hs.QuerySections("type:example")

// Boolean logic
results, _ := hs.QuerySections("type:example AND topic:database")

// Full-text search
results, _ := hs.QuerySections(\`"SQLite integration"\`)
\`\`\`

## Cobra Integration

The help system extends Cobra's built-in help functionality by automatically displaying relevant documentation sections when users request help for specific commands.

\`\`\`go
help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)
\`\`\`

## See Also

- \`writing-help-entries\` — How to create new documentation sections
- \`how-to-write-good-documentation-pages\` — Style guide for docs
- \`using-the-query-api\` — Full query DSL reference`,
  },
  {
    slug: "writing-help-entries",
    title: "Writing Help Entries",
    short: "Learn how to create and structure Markdown documents for the Glazed help system.",
    type: "GeneralTopic",
    topics: ["documentation", "help system", "markdown"],
    commands: ["AddDocToHelpSystem"],
    flags: [],
    isTopLevel: true,
    content: `## Before You Start

Follow the conventions in \`how-to-write-good-documentation-pages\` when drafting new sections. In particular:

- Keep frontmatter concise and single-purpose
- Use the same "friendly but factual" tone across documents
- Prefer present tense and active voice
- Verify each Slug is unique before committing

## Markdown File Structure

Each Markdown file represents a single "section" in the help system. A section can be one of these types:

1. **General Topic** — A general article or topic
2. **Example** — A specific usage demonstration
3. **Application** — A complex use case with multiple commands
4. **Tutorial** — A step-by-step guide

## Frontmatter Fields

\`\`\`yaml
---
Title: The title of the section
Slug: a-unique-slug-for-this-section
Short: A short description (one or two sentences)
Topics:
- topic1
Commands:
- command1
Flags:
- flag1
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---
\`\`\`

Note: There is no top-level \`#\` title — the help system adds that automatically.

## Loading Sections

Create a \`doc\` package with an embedded filesystem:

\`\`\`go
package doc

import (
    "embed"
    "github.com/go-go-golems/glazed/pkg/help"
)

//go:embed *
var docFS embed.FS

func AddDocToHelpSystem(hs *help.HelpSystem) error {
    return hs.LoadSectionsFromFS(docFS, ".")
}
\`\`\`

## Registering with Cobra

\`\`\`go
helpSystem := help.NewHelpSystem()
doc.AddDocToHelpSystem(helpSystem)
help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)
\`\`\`

## See Also

- \`help-system\` — Overview of the help system architecture
- \`how-to-write-good-documentation-pages\` — Style guide`,
  },
  {
    slug: "how-to-write-good-documentation-pages",
    title: "Writing Good Docs",
    short: "Style guide and best practices for creating clear, consistent, and helpful documentation.",
    type: "GeneralTopic",
    topics: ["documentation", "style-guide", "writing"],
    commands: [],
    flags: [],
    isTopLevel: false,
    content: `## Why This Matters

Bad documentation wastes developer time. Good documentation respects the reader's time by being **immediately useful**. Every sentence should either teach something or help the reader navigate to what they need.

## Avoiding Terse Documentation

For each concept, answer these questions:

1. **What is it?** — One sentence definition
2. **Why does it exist?** — The problem it solves
3. **When do you use it?** — The triggering situation
4. **How do you use it?** — Code example
5. **What happens if you don't?** — The failure mode
6. **What's related?** — Links to connected concepts

## Section Introductions

Every \`##\` section must start with a paragraph that explains the concept, not just describes the section.

**Bad:** "This section covers the event system."

**Good:** "Events flow from engines through a Watermill-backed router to your handlers. Each token, tool call, and error becomes a structured event."

## Code Examples

- Keep examples minimal — only what's needed for the concept
- Comments explain *why*, not *what*
- Show expected output when code produces it

## Document Types

| Reader's Question | Document Type |
|---|---|
| "What is X and how does it work?" | Topic |
| "How do I learn to use X?" | Tutorial |
| "How do I do X right now?" | Playbook |

## Anti-Patterns

- **Wall of Text** — Break into sections, use bullets
- **Code-First** — Explain what and why before every code block
- **Undefined Jargon** — Explain terms on first use
- **Missing Failure Modes** — Document what happens when things go wrong
- **Orphan Docs** — Always cross-reference

## Writing Style

- Active voice, direct tone
- Write like explaining to a colleague
- Assume the reader knows the language but is new to the library

## See Also

- \`writing-help-entries\` — How to create help sections
- \`help-system\` — The help system architecture`,
  },
  {
    slug: "using-the-query-api",
    title: "Using the Query API",
    short: "Reference for the Glazed help system query DSL — filters, boolean logic, and full-text search.",
    type: "GeneralTopic",
    topics: ["query", "dsl", "api", "search"],
    commands: ["help"],
    flags: ["flag", "topic", "command"],
    isTopLevel: true,
    content: `## Query Basics

The query API lets you search and filter help sections using a simple DSL. Queries combine field filters with boolean operators.

## Field Filters

\`\`\`
type:example        — Match by section type
topic:database      — Match by topic tag
command:json        — Match by associated command
flag:--output       — Match by associated flag
toplevel:true       — Match top-level sections
default:true        — Match default-shown sections
\`\`\`

## Boolean Operators

\`\`\`
type:example AND topic:database
type:example OR type:tutorial
type:example AND NOT topic:advanced
(type:example OR type:tutorial) AND topic:database
\`\`\`

## Full-Text Search

Wrap phrases in double quotes for full-text search:

\`\`\`
"SQLite integration"
"query DSL"
\`\`\`

## Go API

\`\`\`go
results, err := hs.QuerySections("type:example AND topic:database")
\`\`\`

## Debugging Queries

\`\`\`go
hs.PrintQueryDebug("type:example AND topic:database", true, true)
// Shows AST and generated SQL
\`\`\`

## See Also

- \`help-system\` — Overview and architecture
- \`writing-help-entries\` — Creating queryable sections`,
  },
];

const TYPE_BADGE = {
  GeneralTopic: { label: "Topic", color: "#4a7c59" },
  Example: { label: "Example", color: "#b8860b" },
  Application: { label: "App", color: "#4a6a8c" },
  Tutorial: { label: "Tutorial", color: "#8c4a6a" },
};

const STRIPE = `url("data:image/svg+xml,%3Csvg width='1' height='2' xmlns='http://www.w3.org/2000/svg'%3E%3Crect width='1' height='1' fill='%23000'/%3E%3C/svg%3E")`;
const DITHER = `url("data:image/svg+xml,%3Csvg width='4' height='4' xmlns='http://www.w3.org/2000/svg'%3E%3Crect x='0' y='0' width='2' height='2' fill='%23000'/%3E%3Crect x='2' y='2' width='2' height='2' fill='%23000'/%3E%3C/svg%3E")`;

function TitleBar({ title }) {
  return (
    <div style={{ background: "#fff", borderBottom: "2px solid #000", display: "flex", alignItems: "stretch", height: 22, userSelect: "none", flexShrink: 0 }}>
      <div style={{ width: 20, borderRight: "2px solid #000", display: "flex", alignItems: "center", justifyContent: "center" }}>
        <div style={{ width: 11, height: 11, border: "2px solid #000", background: "#fff" }} />
      </div>
      <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
        <div style={{ flex: 1, height: 6, margin: "0 6px", backgroundImage: STRIPE, backgroundSize: "1px 2px", backgroundRepeat: "repeat" }} />
        <span style={{ fontFamily: "var(--font)", fontSize: 12, fontWeight: 700, padding: "0 8px", background: "#fff", whiteSpace: "nowrap", letterSpacing: 0.3 }}>{title}</span>
        <div style={{ flex: 1, height: 6, margin: "0 6px", backgroundImage: STRIPE, backgroundSize: "1px 2px", backgroundRepeat: "repeat" }} />
      </div>
    </div>
  );
}

function Badge({ text, variant = "topic" }) {
  const isType = variant === "type";
  const color = isType ? (TYPE_BADGE[text]?.color || "#000") : variant === "command" ? "#4a7c59" : variant === "flag" ? "#8c4a6a" : "#000";
  return (
    <span style={{
      display: "inline-block", padding: "1px 7px", border: `1.5px solid ${color}`,
      borderRadius: 2, fontSize: 10, fontFamily: "var(--font)", fontWeight: isType ? 700 : 400,
      color, background: "#fff", whiteSpace: "nowrap", letterSpacing: 0.3,
    }}>
      {isType ? TYPE_BADGE[text]?.label || text : text}
    </span>
  );
}

function inlineFormat(text) {
  const parts = [];
  const re = /(\*\*(.+?)\*\*)|(`([^`]+?)`)/g;
  let last = 0, m, k = 0;
  while ((m = re.exec(text)) !== null) {
    if (m.index > last) parts.push(text.slice(last, m.index));
    if (m[2]) parts.push(<strong key={k++}>{m[2]}</strong>);
    if (m[4]) parts.push(<code key={k++} style={{ background: "#e8e8e8", padding: "1px 4px", border: "1px solid #bbb", fontFamily: "var(--font-mono)", fontSize: "0.9em" }}>{m[4]}</code>);
    last = m.index + m[0].length;
  }
  if (last < text.length) parts.push(text.slice(last));
  return parts.length ? parts : text;
}

function renderMarkdown(md) {
  const lines = md.split("\n");
  const els = [];
  let i = 0, key = 0;
  while (i < lines.length) {
    const line = lines[i];
    if (line.startsWith("```")) {
      const lang = line.slice(3).trim();
      const code = []; i++;
      while (i < lines.length && !lines[i].startsWith("```")) { code.push(lines[i]); i++; }
      i++;
      els.push(
        <div key={key++} style={{ margin: "14px 0", border: "2px solid #000", background: "#1a1a1a" }}>
          {lang && <div style={{ background: "#000", color: "#888", padding: "3px 8px", fontSize: 10, fontFamily: "var(--font)", borderBottom: "1px solid #333", letterSpacing: 0.5, textTransform: "uppercase" }}>{lang}</div>}
          <pre style={{ margin: 0, padding: "12px 14px", color: "#c0c0c0", fontSize: 12, lineHeight: 1.65, fontFamily: "var(--font-mono)", overflow: "auto", whiteSpace: "pre-wrap" }}><code>{code.join("\n")}</code></pre>
        </div>
      );
      continue;
    }
    if (line.startsWith("## ")) {
      els.push(<h2 key={key++} style={{ fontSize: 16, fontWeight: 700, fontFamily: "var(--font)", margin: "28px 0 10px", borderBottom: "1px solid #000", paddingBottom: 4 }}>{line.slice(3)}</h2>);
      i++; continue;
    }
    if (line.startsWith("| ") && lines[i + 1]?.startsWith("|--")) {
      const headers = line.split("|").filter(Boolean).map(h => h.trim());
      i += 2; const rows = [];
      while (i < lines.length && lines[i].startsWith("|")) { rows.push(lines[i].split("|").filter(Boolean).map(c => c.trim())); i++; }
      els.push(
        <div key={key++} style={{ margin: "12px 0", border: "2px solid #000", overflow: "auto" }}>
          <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 12, fontFamily: "var(--font)" }}>
            <thead><tr style={{ background: "#e8e8e8" }}>{headers.map((h, j) => <th key={j} style={{ textAlign: "left", padding: "6px 10px", borderBottom: "2px solid #000", borderRight: j < headers.length - 1 ? "1px solid #999" : "none", fontWeight: 700 }}>{h}</th>)}</tr></thead>
            <tbody>{rows.map((row, ri) => <tr key={ri}>{row.map((cell, ci) => <td key={ci} style={{ padding: "5px 10px", borderBottom: ri < rows.length - 1 ? "1px solid #ccc" : "none", borderRight: ci < row.length - 1 ? "1px solid #ddd" : "none" }}>{inlineFormat(cell)}</td>)}</tr>)}</tbody>
          </table>
        </div>
      );
      continue;
    }
    if (line.startsWith("- ")) {
      const items = [];
      while (i < lines.length && lines[i].startsWith("- ")) { items.push(lines[i].slice(2)); i++; }
      els.push(<ul key={key++} style={{ margin: "10px 0", paddingLeft: 22 }}>{items.map((item, j) => <li key={j} style={{ marginBottom: 4, lineHeight: 1.6 }}>{inlineFormat(item)}</li>)}</ul>);
      continue;
    }
    if (/^\d+\. /.test(line)) {
      const items = [];
      while (i < lines.length && /^\d+\. /.test(lines[i])) { items.push(lines[i].replace(/^\d+\. /, "")); i++; }
      els.push(<ol key={key++} style={{ margin: "10px 0", paddingLeft: 22 }}>{items.map((item, j) => <li key={j} style={{ marginBottom: 4, lineHeight: 1.6 }}>{inlineFormat(item)}</li>)}</ol>);
      continue;
    }
    if (line.trim() === "") { i++; continue; }
    els.push(<p key={key++} style={{ margin: "10px 0", lineHeight: 1.65, fontSize: 13 }}>{inlineFormat(line)}</p>);
    i++;
  }
  return els;
}

export default function App() {
  const [active, setActive] = useState("help-system");
  const [search, setSearch] = useState("");
  const [filter, setFilter] = useState("All");

  const types = ["All", "GeneralTopic", "Example", "Application", "Tutorial"];

  const filtered = useMemo(() => {
    return SECTIONS.filter(s => {
      if (filter !== "All" && s.type !== filter) return false;
      if (!search) return true;
      const q = search.toLowerCase();
      return s.title.toLowerCase().includes(q) || s.short.toLowerCase().includes(q) || s.topics.some(t => t.includes(q)) || s.slug.includes(q);
    });
  }, [search, filter]);

  const section = SECTIONS.find(s => s.slug === active);

  return (
    <div style={{
      "--font": "'Chicago_', 'Geneva', 'Charcoal', 'Lucida Grande', 'Helvetica Neue', sans-serif",
      "--font-mono": "'Monaco', 'Courier New', monospace",
      display: "flex", flexDirection: "column", height: "100vh", fontFamily: "var(--font)", fontSize: 13, color: "#000",
      background: "#a8a8a8", backgroundImage: DITHER, backgroundSize: "4px 4px",
    }}>
      <style>{`
        @font-face { font-family: 'Chicago_'; src: url('https://cdn.jsdelivr.net/gh/polgfred/mac-fonts@main/Chicago.woff2') format('woff2'); }
        * { box-sizing: border-box; }
        ::-webkit-scrollbar { width: 16px; }
        ::-webkit-scrollbar-track { background: #fff; border-left: 1px solid #000; }
        ::-webkit-scrollbar-thumb { background: #888; border: 1px solid #000; }
        ::selection { background: #000; color: #fff; }
      `}</style>

      {/* Menu bar */}
      <div style={{ height: 22, background: "#fff", borderBottom: "2px solid #000", display: "flex", alignItems: "center", padding: "0 12px", gap: 18, fontSize: 12, fontWeight: 700, flexShrink: 0 }}>
        <span style={{ fontSize: 14 }}>&#63743;</span>
        <span>File</span>
        <span>Edit</span>
        <span>View</span>
        <span>Help</span>
        <span style={{ marginLeft: "auto", fontWeight: 400, fontSize: 11, letterSpacing: 0.3 }}>Glazed Documentation Browser</span>
      </div>

      <div style={{ display: "flex", flex: 1, padding: 8, gap: 1, overflow: "hidden" }}>
        {/* Sidebar window */}
        <div style={{ width: 280, minWidth: 280, border: "2px solid #000", background: "#fff", boxShadow: "2px 2px 0 #000", display: "flex", flexDirection: "column", flexShrink: 0 }}>
          <TitleBar title="📁 Sections" />
          <div style={{ padding: "10px 10px 8px", borderBottom: "2px solid #000" }}>
            <div style={{ border: "2px inset #999", background: "#fff", display: "flex", alignItems: "center", marginBottom: 8 }}>
              <span style={{ padding: "0 6px", fontSize: 12 }}>🔍</span>
              <input value={search} onChange={e => setSearch(e.target.value)} placeholder="Search…"
                style={{ border: "none", outline: "none", padding: "5px 6px 5px 0", fontSize: 12, fontFamily: "var(--font)", width: "100%", background: "transparent" }} />
            </div>
            <div style={{ display: "flex", gap: 4, flexWrap: "wrap" }}>
              {types.map(t => (
                <button key={t} onClick={() => setFilter(t)} style={{
                  padding: "2px 8px", fontSize: 10, fontFamily: "var(--font)", fontWeight: filter === t ? 700 : 400,
                  border: filter === t ? "2px solid #000" : "1px solid #999",
                  background: filter === t ? "#000" : "#fff", color: filter === t ? "#fff" : "#000",
                  cursor: "pointer", borderRadius: 0,
                }}>{t === "All" ? "All" : TYPE_BADGE[t]?.label || t}</button>
              ))}
            </div>
          </div>
          <div style={{ flex: 1, overflow: "auto" }}>
            {filtered.map((s, idx) => {
              const on = active === s.slug;
              return (
                <button key={s.slug} onClick={() => setActive(s.slug)} style={{
                  width: "100%", textAlign: "left", padding: "10px 12px", border: "none",
                  borderBottom: "1px solid #ddd", cursor: "pointer", fontFamily: "var(--font)",
                  display: "block", background: on ? "#000" : idx % 2 === 0 ? "#fff" : "#f5f5f5",
                  color: on ? "#fff" : "#000",
                }}>
                  <div style={{ display: "flex", gap: 6, alignItems: "center", marginBottom: 3 }}>
                    {!on && <Badge text={s.type} variant="type" />}
                    {on && <span style={{ fontSize: 10, fontWeight: 700, color: "#aaa" }}>{TYPE_BADGE[s.type]?.label}</span>}
                    {s.isTopLevel && <span style={{ fontSize: 9, color: on ? "#888" : "#999" }}>◆ TOP</span>}
                  </div>
                  <div style={{ fontWeight: 700, fontSize: 12, marginBottom: 2 }}>{s.title}</div>
                  <div style={{ fontSize: 10, color: on ? "#aaa" : "#777", lineHeight: 1.4, display: "-webkit-box", WebkitLineClamp: 2, WebkitBoxOrient: "vertical", overflow: "hidden" }}>{s.short}</div>
                </button>
              );
            })}
          </div>
          <div style={{ borderTop: "2px solid #000", padding: "5px 10px", fontSize: 10, color: "#777", display: "flex", justifyContent: "space-between" }}>
            <span>{filtered.length} sections</span>
            <span>Glazed v0.1</span>
          </div>
        </div>

        {/* Main content window */}
        <div style={{ flex: 1, border: "2px solid #000", background: "#fff", boxShadow: "2px 2px 0 #000", display: "flex", flexDirection: "column", overflow: "hidden" }}>
          <TitleBar title={section ? `📄 ${section.title} — glaze help ${section.slug}` : "📄 Documentation"} />
          {!section ? (
            <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: "#999", fontSize: 13 }}>
              <div style={{ textAlign: "center" }}>
                <div style={{ fontSize: 40, marginBottom: 10, filter: "grayscale(1)" }}>📖</div>
                Select a section from the list.
              </div>
            </div>
          ) : (
            <div style={{ flex: 1, overflow: "auto" }}>
              <div style={{ maxWidth: 680, margin: "0 auto", padding: "28px 32px 60px", fontFamily: "var(--font)" }}>
                <div style={{ marginBottom: 24 }}>
                  <div style={{ display: "flex", gap: 8, alignItems: "center", marginBottom: 10, flexWrap: "wrap" }}>
                    <Badge text={section.type} variant="type" />
                    <span style={{ fontSize: 10, color: "#999", fontFamily: "var(--font-mono)", background: "#f0f0f0", padding: "1px 6px", border: "1px solid #ccc" }}>{section.slug}</span>
                  </div>
                  <h1 style={{ fontSize: 24, fontWeight: 700, margin: "0 0 8px", letterSpacing: -0.3 }}>{section.title}</h1>
                  <p style={{ fontSize: 12, color: "#555", lineHeight: 1.55, margin: "0 0 14px" }}>{section.short}</p>
                  <div style={{ display: "flex", gap: 5, flexWrap: "wrap" }}>
                    {section.topics.map(t => <Badge key={t} text={t} />)}
                    {section.commands.map(c => <Badge key={c} text={c} variant="command" />)}
                    {section.flags.slice(0, 4).map(f => <Badge key={f} text={f} variant="flag" />)}
                    {section.flags.length > 4 && <span style={{ fontSize: 10, color: "#999" }}>+{section.flags.length - 4}</span>}
                  </div>
                </div>
                <div style={{ borderTop: "2px solid #000", paddingTop: 20 }}>
                  {renderMarkdown(section.content)}
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
