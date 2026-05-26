package server

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	stdpath "path"

	"github.com/go-go-golems/glazed/pkg/help/store"
)

// wellKnownPaths maps well-known agent file paths to their Content-Type.
// The Go server intercepts these paths before the SPA fallback, so agents
// get real content instead of an HTML shell.
var wellKnownPaths = map[string]string{
	"/llms.txt":    "text/plain; charset=utf-8",
	"/robots.txt":  "text/plain; charset=utf-8",
	"/AGENTS.md":   "text/markdown; charset=utf-8",
	"/sitemap.xml": "application/xml; charset=utf-8",
	"/sitemap.md":  "text/markdown; charset=utf-8",
	"/index.md":    "text/markdown; charset=utf-8",
}

// WellKnownHandler generates well-known agent files dynamically from the
// help database. It intercepts paths like /llms.txt, /robots.txt, /AGENTS.md,
// /sitemap.xml, /sitemap.md, and /index.md before they hit the SPA fallback.
type WellKnownHandler struct {
	deps HandlerDeps
}

func NewWellKnownHandler(deps HandlerDeps) *WellKnownHandler {
	return &WellKnownHandler{deps: deps}
}

// CanHandle returns true if the path is a well-known agent file.
func (h *WellKnownHandler) CanHandle(path string) bool {
	_, ok := wellKnownPaths[path]
	return ok
}

func (h *WellKnownHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cleanPath := stdpath.Clean("/" + r.URL.Path)
	contentType, ok := wellKnownPaths[cleanPath]
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=300")

	// Add Link header with canonical URL for markdown files.
	baseURL := deriveBaseURL(r)
	if cleanPath == "/index.md" {
		w.Header().Set("Link", fmt.Sprintf("<%s/>; rel=\"canonical\"", baseURL))
	} else if strings.HasSuffix(cleanPath, ".md") {
		htmlPath := strings.TrimSuffix(cleanPath, ".md")
		w.Header().Set("Link", fmt.Sprintf("<%s%s>; rel=\"canonical\"", baseURL, htmlPath))
	}

	var body string
	switch cleanPath {
	case "/llms.txt":
		body = h.generateLlmsTxt(r)
	case "/robots.txt":
		body = h.generateRobotsTxt()
	case "/AGENTS.md":
		body = h.generateAgentsMd(r)
	case "/sitemap.xml":
		body = h.generateSitemapXML(r)
	case "/sitemap.md":
		body = h.generateSitemapMd(r)
	case "/index.md":
		body = h.generateIndexMd(r)
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

// deriveBaseURL infers the site's base URL from the request.
func deriveBaseURL(r *http.Request) string {
	scheme := "http"
	if r.Header.Get("X-Forwarded-Proto") == "https" || r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		host = "localhost:8088"
	}
	return scheme + "://" + host
}

// ---------------------------------------------------------------------------
// /llms.txt
// ---------------------------------------------------------------------------

func (h *WellKnownHandler) generateLlmsTxt(r *http.Request) string {
	baseURL := deriveBaseURL(r)
	ctx := context.Background()

	pkgs, err := h.deps.Store.ListPackages(ctx)
	if err != nil {
		return "# Glazed Help Browser\n\n> Error loading packages.\n"
	}

	var b strings.Builder
	b.WriteString("# Glazed Help Browser\n\n")
	b.WriteString("> Documentation browser for the Glazed CLI framework and Go-Go-Golems tools.\n\n")
	b.WriteString("This site provides structured help documentation for Go-Go-Golems CLI tools. ")
	b.WriteString("Browse by package and version, or search across all sections.\n\n")

	pkgMap := groupPackages(pkgs)
	names := sortedKeys(pkgMap)

	b.WriteString("## Packages\n\n")
	for _, name := range names {
		versions := pkgMap[name]
		pkgURL := fmt.Sprintf("%s/%s/_", baseURL, name)
		var totalCount int
		for _, v := range versions {
			totalCount += v.SectionCount
		}
		fmt.Fprintf(&b, "- [%s](%s): %d sections\n", name, pkgURL, totalCount)
		for _, v := range versions {
			if v.Version == "" {
				continue
			}
			vURL := fmt.Sprintf("%s/%s/%s", baseURL, name, v.Version)
			fmt.Fprintf(&b, "  - [%s %s](%s): %d sections\n", name, v.Version, vURL, v.SectionCount)
		}
	}

	// List top-level sections — links use .md suffix for agent readability.
	b.WriteString("\n## Sections\n\n")
	for _, name := range names {
		sections, err := h.deps.Store.List(ctx, "")
		if err != nil {
			continue
		}
		for _, s := range sections {
			if !s.IsTopLevel || s.PackageName != name {
				continue
			}
			ver := s.PackageVersion
			if ver == "" {
				ver = "_"
			}
			// Use .md suffix so agents get markdown content directly.
			url := fmt.Sprintf("%s/%s/%s/sections/%s.md", baseURL, name, ver, s.Slug)
			desc := ""
			if s.Short != "" {
				desc = ": " + s.Short
			}
			fmt.Fprintf(&b, "- [%s](%s)%s\n", s.Title, url, desc)
		}
	}

	return b.String()
}

// ---------------------------------------------------------------------------
// /robots.txt
// ---------------------------------------------------------------------------

func (h *WellKnownHandler) generateRobotsTxt() string {
	var b strings.Builder
	b.WriteString("User-agent: *\n")
	b.WriteString("Allow: /\n\n")
	b.WriteString("User-agent: GPTBot\n")
	b.WriteString("Allow: /\n\n")
	b.WriteString("User-agent: ChatGPT-User\n")
	b.WriteString("Allow: /\n\n")
	b.WriteString("User-agent: CCBot\n")
	b.WriteString("Allow: /\n\n")
	b.WriteString("User-agent: ClaudeBot\n")
	b.WriteString("Allow: /\n\n")
	b.WriteString("User-agent: Google-Extended\n")
	b.WriteString("Allow: /\n\n")
	b.WriteString("Sitemap: https://docs.yolo.scapegoat.dev/sitemap.xml\n")
	return b.String()
}

// ---------------------------------------------------------------------------
// /AGENTS.md
// ---------------------------------------------------------------------------

func (h *WellKnownHandler) generateAgentsMd(r *http.Request) string {
	baseURL := deriveBaseURL(r)
	ctx := context.Background()

	pkgs, err := h.deps.Store.ListPackages(ctx)
	if err != nil {
		return "# AGENTS.md\n\nError loading packages.\n"
	}

	var b strings.Builder
	b.WriteString("# AGENTS.md\n\n")
	b.WriteString("This site hosts documentation for Go-Go-Golems CLI tools.\n\n")

	b.WriteString("## How to install\n\n")
	b.WriteString("Install the `glaze` CLI from the Go-Go-Golems repository:\n\n")
	b.WriteString("```bash\ngo install github.com/go-go-golems/glazed/cmd/glaze@latest\n```\n\n")

	b.WriteString("## How to use\n\n")
	fmt.Fprintf(&b, "Browse documentation at %s. ", baseURL)
	b.WriteString("URL scheme: `/{package}/{version}/sections/{slug}`. ")
	b.WriteString("Append `.md` to any section URL for raw Markdown. ")
	b.WriteString("Use `/_/` for unversioned packages.\n\n")

	b.WriteString("## API\n\n")
	fmt.Fprintf(&b, "- `GET %s/api/packages` — list all packages and versions\n", baseURL)
	fmt.Fprintf(&b, "- `GET %s/api/sections?package=X&version=Y` — list sections\n", baseURL)
	fmt.Fprintf(&b, "- `GET %s/api/sections/{slug}?package=X&version=Y` — get section content\n", baseURL)
	fmt.Fprintf(&b, "- `GET %s/api/health` — health check\n\n", baseURL)

	b.WriteString("## Available packages\n\n")
	pkgMap := groupPackages(pkgs)
	for _, name := range sortedKeys(pkgMap) {
		versions := pkgMap[name]
		var totalCount int
		for _, v := range versions {
			totalCount += v.SectionCount
		}
		fmt.Fprintf(&b, "- **%s**: %d sections\n", name, totalCount)
	}

	return b.String()
}

// ---------------------------------------------------------------------------
// /sitemap.xml
// ---------------------------------------------------------------------------

func (h *WellKnownHandler) generateSitemapXML(r *http.Request) string {
	baseURL := deriveBaseURL(r)
	ctx := context.Background()

	sections, err := h.deps.Store.List(ctx, "")
	if err != nil {
		return `<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"></urlset>`
	}

	now := time.Now().Format("2006-01-02")

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString("\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n")

	// Root URL
	b.WriteString("  <url>\n")
	fmt.Fprintf(&b, "    <loc>%s/</loc>\n", baseURL)
	fmt.Fprintf(&b, "    <lastmod>%s</lastmod>\n", now)
	b.WriteString("    <changefreq>daily</changefreq>\n")
	b.WriteString("    <priority>1.0</priority>\n")
	b.WriteString("  </url>\n")

	// Per-section URLs
	for _, s := range sections {
		ver := s.PackageVersion
		if ver == "" {
			ver = "_"
		}
		url := fmt.Sprintf("%s/%s/%s/sections/%s", baseURL, s.PackageName, ver, s.Slug)
		b.WriteString("  <url>\n")
		fmt.Fprintf(&b, "    <loc>%s</loc>\n", url)
		fmt.Fprintf(&b, "    <lastmod>%s</lastmod>\n", now)
		b.WriteString("    <changefreq>weekly</changefreq>\n")
		fmt.Fprintf(&b, "    <priority>%.1f</priority>\n", 0.8)
		b.WriteString("  </url>\n")
	}

	b.WriteString("</urlset>\n")
	return b.String()
}

// ---------------------------------------------------------------------------
// /sitemap.md
// ---------------------------------------------------------------------------

func (h *WellKnownHandler) generateSitemapMd(r *http.Request) string {
	baseURL := deriveBaseURL(r)
	ctx := context.Background()

	pkgs, err := h.deps.Store.ListPackages(ctx)
	if err != nil {
		return "# Sitemap\n\nError loading packages.\n"
	}

	var b strings.Builder
	b.WriteString("# Sitemap\n\n")
	fmt.Fprintf(&b, "Documentation browser for Go-Go-Golems tools: %s\n\n", baseURL)

	pkgMap := groupPackages(pkgs)
	for _, name := range sortedKeys(pkgMap) {
		fmt.Fprintf(&b, "## %s\n\n", name)

		sections, err := h.deps.Store.List(ctx, "")
		if err != nil {
			continue
		}
		for _, s := range sections {
			if s.PackageName != name || !s.IsTopLevel {
				continue
			}
			ver := s.PackageVersion
			if ver == "" {
				ver = "_"
			}
			// Use .md suffix for agent-friendly links.
			url := fmt.Sprintf("%s/%s/%s/sections/%s.md", baseURL, name, ver, s.Slug)
			fmt.Fprintf(&b, "- [%s](%s)\n", s.Title, url)
		}
		b.WriteString("\n")

		for _, v := range pkgMap[name] {
			if v.Version == "" {
				continue
			}
			vURL := fmt.Sprintf("%s/%s/%s", baseURL, name, v.Version)
			fmt.Fprintf(&b, "- [%s %s](%s) — %d sections\n", name, v.Version, vURL, v.SectionCount)
		}
		b.WriteString("\n")
	}

	return b.String()
}

// ---------------------------------------------------------------------------
// /index.md (Markdown mirror of the root page)
// ---------------------------------------------------------------------------

func (h *WellKnownHandler) generateIndexMd(r *http.Request) string {
	baseURL := deriveBaseURL(r)
	ctx := context.Background()

	pkgs, err := h.deps.Store.ListPackages(ctx)
	if err != nil {
		return "---\ntitle: Glazed Help Browser\n---\n\nError loading packages.\n"
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("title: Glazed Help Browser\n")
	b.WriteString("description: Documentation browser for the Glazed CLI framework and Go-Go-Golems tools.\n")
	b.WriteString("doc_version: 1\n")
	fmt.Fprintf(&b, "last_updated: %s\n", time.Now().Format("2006-01-02"))
	b.WriteString("---\n\n")

	b.WriteString("# Glazed Help Browser\n\n")
	b.WriteString("Documentation browser for the Glazed CLI framework and Go-Go-Golems tools. ")
	fmt.Fprintf(&b, "Browse at <%s>.\n\n", baseURL)

	b.WriteString("## Packages\n\n")
	pkgMap := groupPackages(pkgs)
	for _, name := range sortedKeys(pkgMap) {
		versions := pkgMap[name]
		var totalCount int
		for _, v := range versions {
			totalCount += v.SectionCount
		}
		pkgURL := fmt.Sprintf("%s/%s/_", baseURL, name)
		fmt.Fprintf(&b, "- [%s](%s) — %d sections\n", name, pkgURL, totalCount)
	}

	b.WriteString("\n## Sitemap\n\n")
	fmt.Fprintf(&b, "- [Sitemap (Markdown)](%s/sitemap.md)\n", baseURL)
	fmt.Fprintf(&b, "- [Sitemap (XML)](%s/sitemap.xml)\n", baseURL)
	fmt.Fprintf(&b, "- [llms.txt](%s/llms.txt)\n", baseURL)
	fmt.Fprintf(&b, "- [AGENTS.md](%s/AGENTS.md)\n", baseURL)

	return b.String()
}

// ---------------------------------------------------------------------------
// .md suffix URL handler (e.g. /glazed/_/sections/foo.md → markdown)
// ---------------------------------------------------------------------------

// isMarkdownSuffixURL returns true if the path ends with .md and matches
// the section URL pattern: /{package}/{version}/sections/{slug}.md
func isMarkdownSuffixURL(path string) (string, string, string, bool) {
	if !strings.HasSuffix(path, ".md") {
		return "", "", "", false
	}
	// Strip .md suffix and parse as section URL
	withoutMd := strings.TrimSuffix(path, ".md")
	return parseSectionURL(withoutMd)
}

// handleMarkdownSuffix serves the raw Markdown content for a .md URL
// (e.g. /glazed/_/sections/jq-filter.md). This is the primary way agents
// access section content — they append .md to any section URL.
func handleMarkdownSuffix(deps HandlerDeps, w http.ResponseWriter, r *http.Request, pkgName, ver, slug string) {
	ctx := context.Background()

	version := ver
	if version == "_" {
		version = ""
	}

	section, err := deps.Store.GetByPackageSlug(ctx, pkgName, version, slug)
	if err != nil || section == nil {
		http.Error(w, "section not found", http.StatusNotFound)
		return
	}

	baseURL := deriveBaseURL(r)
	canonicalURL := fmt.Sprintf("%s/%s/%s/sections/%s", baseURL, pkgName, ver, slug)

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Link", fmt.Sprintf("<%s>; rel=\"canonical\"", canonicalURL))

	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %s\n", section.Title)
	if section.Short != "" {
		fmt.Fprintf(&b, "description: %s\n", section.Short)
	}
	b.WriteString("doc_version: 1\n")
	fmt.Fprintf(&b, "last_updated: %s\n", time.Now().Format("2006-01-02"))
	b.WriteString("---\n\n")
	b.WriteString(section.Content)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(b.String()))
}

// handleMarkdownMirror serves the raw Markdown content for a normal section
// URL when content negotiation asks for text/markdown. It shares the same
// response shape as the explicit .md suffix handler, but is reached from the
// original HTML URL via Accept: text/markdown.
func handleMarkdownMirror(deps HandlerDeps, w http.ResponseWriter, r *http.Request, pkgName, ver, slug string) {
	handleMarkdownSuffix(deps, w, r, pkgName, ver, slug)
}

// handleRootMarkdownMirror returns the index.md content for the root URL
// when Accept: text/markdown is sent.
func handleRootMarkdownMirror(deps HandlerDeps, w http.ResponseWriter, r *http.Request) {
	h := NewWellKnownHandler(deps)
	body := h.generateIndexMd(r)

	baseURL := deriveBaseURL(r)
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.Header().Set("Link", fmt.Sprintf("<%s/>; rel=\"canonical\"", baseURL))

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

type pkgVersionInfo struct {
	Version      string
	SectionCount int
}

func groupPackages(pkgs []store.PackageInfo) map[string][]pkgVersionInfo {
	m := make(map[string][]pkgVersionInfo)
	for _, p := range pkgs {
		m[p.Name] = append(m[p.Name], pkgVersionInfo{
			Version:      p.Version,
			SectionCount: p.SectionCount,
		})
	}
	return m
}

func sortedKeys(m map[string][]pkgVersionInfo) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// parseSectionURL extracts package, version, and slug from a URL path that
// matches the /{package}/{version}/sections/{slug} scheme.
func parseSectionURL(path string) (string, string, string, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 4 && parts[2] == "sections" {
		return parts[0], parts[1], parts[3], true
	}
	return "", "", "", false
}

// wantsMarkdown returns true if the request's Accept header prefers
// text/markdown over text/html.
func wantsMarkdown(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return false
	}
	mdIdx := strings.Index(accept, "text/markdown")
	htmlIdx := strings.Index(accept, "text/html")
	if mdIdx < 0 {
		return false
	}
	if htmlIdx < 0 {
		return true
	}
	return mdIdx < htmlIdx
}

// alternateLinkWriter wraps an http.ResponseWriter to inject Link and
// <link rel=alternate> headers into HTML responses for section pages.
type alternateLinkWriter struct {
	http.ResponseWriter
	path    string
	baseURL string
	header  bool
}

func (w *alternateLinkWriter) WriteHeader(code int) {
	if !w.header {
		mdURL := fmt.Sprintf("%s%s.md", w.baseURL, w.path)
		w.ResponseWriter.Header().Add("Link", fmt.Sprintf("<%s>; rel=\"alternate\"; type=\"text/markdown\"", mdURL))
		w.header = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *alternateLinkWriter) Write(b []byte) (int, error) {
	if !w.header {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}
