// Package server provides an HTTP API surface for browsing Glazed help sections.
// It is self-contained and has no dependencies on the CLI or Cobra — it only needs
// a HelpSystem instance to operate.
//
// The public entry point is NewHandler, which returns an http.Handler that can be
// mounted anywhere. It routes GET /api/health, GET /api/sections, GET /api/sections/search,
// and GET /api/sections/:slug internally using a lightweight ServeMux.
//
// Example usage as a standalone server:
//
//	srv := &http.Server{Addr: ":8088", Handler: server.NewHandler(deps)}
//	log.Fatal(srv.ListenAndServe())
//
// Example usage as a sub-router on an existing server:
//
//	mux := http.NewServeMux()
//	mux.Handle("/help/", server.NewHandler(deps))
//	http.ListenAndServe(":8080", mux)  // serves /help/api/health, etc.
package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/store"
)

// HandlerDeps holds the dependencies for all HTTP handlers.
type HandlerDeps struct {
	Store  *store.Store
	Logger *slog.Logger
}

// Handler is an http.Handler that routes all /api/* requests internally.
// Construct it with NewHandler.
type Handler struct {
	deps HandlerDeps
	mux  *http.ServeMux
}

// NewHandler returns an http.Handler that serves the Glazed help browser API.
// It panics if deps.Store is nil.
//
// The returned handler already includes CORS headers for all responses.
// For callers that want to add additional middleware, use the returned
// http.Handler directly.
func NewHandler(deps HandlerDeps) http.Handler {
	if deps.Store == nil {
		panic("server.NewHandler: deps.Store must not be nil")
	}
	if deps.Logger == nil {
		deps.Logger = slog.Default()
	}

	h := &Handler{deps: deps}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", h.handleHealth)
	mux.HandleFunc("GET /api/sections/search", h.handleListSections)
	mux.HandleFunc("GET /api/sections", h.handleListSections)
	mux.HandleFunc("GET /api/sections/{slug}", h.handleGetSection)

	h.mux = mux

	// CORS is always applied so any caller gets correct headers without needing
	// to remember to wrap with NewCORSHandler.
	return NewCORSHandler(h)
}

// ServeHTTP implements http.Handler by delegating to the internal mux.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// ---------------------------------------------------------------------------
// Response helpers
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{Error: code, Message: message})
}

// ---------------------------------------------------------------------------
// GET /api/health
// ---------------------------------------------------------------------------

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	count, err := h.deps.Store.Count(ctx)
	if err != nil {
		h.deps.Logger.Error("health check failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to count sections")
		return
	}
	writeJSON(w, http.StatusOK, HealthResponse{OK: true, Sections: int(count)})
}

// ---------------------------------------------------------------------------
// GET /api/sections  and  GET /api/sections/search
// ---------------------------------------------------------------------------

func (h *Handler) handleListSections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := parseListParams(r)

	var sections []*model.Section
	var err error

	if pred := buildPredicate(params); pred != nil {
		sections, err = h.deps.Store.Find(ctx, pred)
	} else {
		sections, err = h.deps.Store.List(ctx, "")
	}
	if err != nil {
		h.deps.Logger.Error("failed to list sections", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list sections")
		return
	}

	total := len(sections)

	// Pagination (offset/limit on the in-memory slice; correct for moderate sizes).
	if params.Offset < 0 {
		params.Offset = 0
	}
	if params.Offset >= len(sections) {
		sections = nil
	} else {
		end := params.Offset + params.Limit
		if params.Limit > 0 && end < len(sections) {
			sections = sections[params.Offset:end]
		} else {
			sections = sections[params.Offset:]
		}
	}

	summaries := make([]SectionSummary, len(sections))
	for i, s := range sections {
		summaries[i] = SummaryFromModel(s)
	}

	writeJSON(w, http.StatusOK, ListSectionsResponse{
		Sections: summaries,
		Total:    total,
		Limit:    params.Limit,
		Offset:   params.Offset,
	})
}

// ---------------------------------------------------------------------------
// GET /api/sections/:slug
// ---------------------------------------------------------------------------

func (h *Handler) handleGetSection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := r.PathValue("slug")

	section, err := h.deps.Store.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, store.ErrSectionNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "section not found: "+slug)
			return
		}
		h.deps.Logger.Error("failed to get section", "slug", slug, "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get section")
		return
	}

	writeJSON(w, http.StatusOK, DetailFromModel(section))
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// parseListParams extracts ListSectionsParams from the query string of r.
func parseListParams(r *http.Request) ListSectionsParams {
	q := r.URL.Query()
	return ListSectionsParams{
		SectionType: q.Get("type"),
		Topic:       q.Get("topic"),
		Command:     q.Get("command"),
		Flag:        q.Get("flag"),
		Search:      q.Get("q"),
		Limit:       parseInt(q.Get("limit"), -1),
		Offset:      parseInt(q.Get("offset"), 0),
	}
}

// parseInt safely parses s as int, returning def if s is empty or invalid.
func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// buildPredicate chains predicates based on params.
// Returns nil if no filters are active (meaning "return everything").
func buildPredicate(params ListSectionsParams) store.Predicate {
	var preds []store.Predicate

	if params.SectionType != "" {
		st, err := model.SectionTypeFromString(params.SectionType)
		if err == nil {
			preds = append(preds, store.IsType(st))
		}
	}
	if params.Topic != "" {
		preds = append(preds, store.HasTopic(params.Topic))
	}
	if params.Command != "" {
		preds = append(preds, store.HasCommand(params.Command))
	}
	if params.Flag != "" {
		preds = append(preds, store.HasFlag(params.Flag))
	}
	if params.Search != "" {
		term := strings.ToLower(params.Search)
		preds = append(preds, func(qc *store.QueryCompiler) {
			// Match against title, short description, and content body.
			qc.AddWhere(
				"LOWER(s.title) LIKE ? OR LOWER(s.short) LIKE ? OR LOWER(s.content) LIKE ?",
				"%"+term+"%", "%"+term+"%", "%"+term+"%",
			)
		})
	}

	if len(preds) == 0 {
		return nil
	}
	if len(preds) == 1 {
		return preds[0]
	}
	return func(qc *store.QueryCompiler) {
		for _, p := range preds {
			p(qc)
		}
	}
}
