package config

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type ConfigLayer string

const (
	LayerSystem   ConfigLayer = "system"
	LayerUser     ConfigLayer = "user"
	LayerRepo     ConfigLayer = "repo"
	LayerCWD      ConfigLayer = "cwd"
	LayerExplicit ConfigLayer = "explicit"
)

type DiscoverFunc func(ctx context.Context) ([]string, error)

type ConditionFunc func(ctx context.Context) bool

type SourceSpec struct {
	Name        string
	Layer       ConfigLayer
	SourceKind  string
	Discover    DiscoverFunc
	Optional    bool
	StopIfFound bool
	EnabledIf   ConditionFunc
}

func (s SourceSpec) Named(name string) SourceSpec {
	s.Name = strings.TrimSpace(name)
	return s
}

func (s SourceSpec) InLayer(layer ConfigLayer) SourceSpec {
	s.Layer = layer
	return s
}

func (s SourceSpec) Kind(kind string) SourceSpec {
	s.SourceKind = strings.TrimSpace(kind)
	return s
}

func (s SourceSpec) When(enabledIf ConditionFunc) SourceSpec {
	s.EnabledIf = enabledIf
	return s
}

type ResolvedConfigFile struct {
	Path       string
	Layer      ConfigLayer
	SourceName string
	SourceKind string
	Index      int
}

type ResolvedSource struct {
	Name            string
	Layer           ConfigLayer
	SourceKind      string
	DiscoveredPaths []string
	AddedPaths      []string
	DedupedPaths    []string
	Found           bool
	SkippedReason   string
}

type PlanReport struct {
	Sources []ResolvedSource
	Files   []ResolvedConfigFile
}

func (r *PlanReport) Paths() []string {
	if r == nil || len(r.Files) == 0 {
		return nil
	}
	ret := make([]string, 0, len(r.Files))
	for _, f := range r.Files {
		ret = append(ret, f.Path)
	}
	return ret
}

func (r *PlanReport) String() string {
	if r == nil {
		return "config resolution plan: <nil>"
	}

	lines := []string{"Config resolution plan:"}
	for i, src := range r.Sources {
		status := "skipped"
		detail := src.SkippedReason
		if len(src.AddedPaths) > 0 {
			status = "found"
			detail = strings.Join(src.AddedPaths, ", ")
			if len(src.DedupedPaths) > 0 {
				detail += fmt.Sprintf(" (deduped: %s)", strings.Join(src.DedupedPaths, ", "))
			}
		} else if len(src.DedupedPaths) > 0 {
			status = "deduped"
			detail = strings.Join(src.DedupedPaths, ", ")
		}
		if strings.TrimSpace(detail) == "" {
			detail = "not found"
		}
		lines = append(lines,
			fmt.Sprintf("%d. %s layer=%s %s %s", i+1, src.Name, src.Layer, status, detail),
		)
	}
	return strings.Join(lines, "\n")
}

type Plan struct {
	layerOrder []ConfigLayer
	sources    []SourceSpec
	dedupe     bool
}

type PlanOption func(*Plan)

func WithLayerOrder(layers ...ConfigLayer) PlanOption {
	return func(p *Plan) {
		p.layerOrder = append([]ConfigLayer(nil), layers...)
	}
}

func WithDedupePaths() PlanOption {
	return func(p *Plan) {
		p.dedupe = true
	}
}

func NewPlan(opts ...PlanOption) *Plan {
	ret := &Plan{}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (p *Plan) Add(sources ...SourceSpec) *Plan {
	p.sources = append(p.sources, sources...)
	return p
}

func (p *Plan) Explain(ctx context.Context) (*PlanReport, error) {
	_, report, err := p.Resolve(ctx)
	return report, err
}

func (p *Plan) Resolve(ctx context.Context) ([]ResolvedConfigFile, *PlanReport, error) {
	report := &PlanReport{}
	if p == nil {
		return nil, report, nil
	}

	sources := p.orderedSources()
	seen := map[string]struct{}{}
	files := make([]ResolvedConfigFile, 0)

	for _, src := range sources {
		resolved := ResolvedSource{
			Name:       src.Name,
			Layer:      src.Layer,
			SourceKind: src.SourceKind,
		}

		if src.EnabledIf != nil && !src.EnabledIf(ctx) {
			resolved.SkippedReason = "disabled by condition"
			report.Sources = append(report.Sources, resolved)
			continue
		}

		if src.Discover == nil {
			resolved.SkippedReason = "no discover function"
			report.Sources = append(report.Sources, resolved)
			continue
		}

		paths, err := src.Discover(ctx)
		if err != nil {
			if src.Optional {
				resolved.SkippedReason = err.Error()
				report.Sources = append(report.Sources, resolved)
				continue
			}
			return nil, report, err
		}
		resolved.DiscoveredPaths = append([]string(nil), paths...)

		if len(paths) == 0 {
			resolved.SkippedReason = "not found"
			report.Sources = append(report.Sources, resolved)
			continue
		}

		for _, path := range paths {
			normalized := normalizePath(path)
			if p.dedupe {
				if _, ok := seen[normalized]; ok {
					resolved.DedupedPaths = append(resolved.DedupedPaths, normalized)
					continue
				}
				seen[normalized] = struct{}{}
			}

			resolved.AddedPaths = append(resolved.AddedPaths, normalized)
			files = append(files, ResolvedConfigFile{
				Path:       normalized,
				Layer:      src.Layer,
				SourceName: src.Name,
				SourceKind: src.SourceKind,
				Index:      len(files),
			})
		}
		resolved.Found = len(resolved.AddedPaths) > 0
		if !resolved.Found && len(resolved.DedupedPaths) > 0 {
			resolved.SkippedReason = "all discovered paths were duplicates"
		}

		report.Sources = append(report.Sources, resolved)
		if src.StopIfFound && resolved.Found {
			break
		}
	}

	report.Files = append(report.Files, files...)
	return files, report, nil
}

func (p *Plan) orderedSources() []SourceSpec {
	ret := append([]SourceSpec(nil), p.sources...)
	if len(ret) <= 1 || len(p.layerOrder) == 0 {
		return ret
	}

	rank := map[ConfigLayer]int{}
	for i, layer := range p.layerOrder {
		rank[layer] = i
	}

	sort.SliceStable(ret, func(i, j int) bool {
		ri, okI := rank[ret[i].Layer]
		rj, okJ := rank[ret[j].Layer]
		switch {
		case okI && okJ:
			return ri < rj
		case okI:
			return true
		case okJ:
			return false
		default:
			return false
		}
	})
	return ret
}

func normalizePath(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	if abs, err := filepath.Abs(path); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}
