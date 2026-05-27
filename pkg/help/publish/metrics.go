package publish

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// RegistryMetrics stores low-cardinality docs-registry counters.
//
// The registry currently runs as a single replica, so in-process counters are
// sufficient for immediate operational visibility. If the registry scales out,
// these metrics should be scraped from every pod and aggregated by Prometheus.
type RegistryMetrics struct {
	mu sync.Mutex

	requestsTotal map[requestMetricKey]uint64
	publishTotal  map[publishMetricKey]uint64
}

type requestMetricKey struct {
	RouteClass string
	Method     string
	Status     int
}

type publishMetricKey struct {
	PackageName string
	Outcome     string
	ErrorCode   string
}

func NewRegistryMetrics() *RegistryMetrics {
	return &RegistryMetrics{
		requestsTotal: map[requestMetricKey]uint64{},
		publishTotal:  map[publishMetricKey]uint64{},
	}
}

func (m *RegistryMetrics) RecordRequest(routeClass, method string, status int) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requestsTotal[requestMetricKey{RouteClass: routeClass, Method: method, Status: status}]++
}

func (m *RegistryMetrics) RecordPublish(event publishAuditEvent) {
	if m == nil {
		return
	}
	errorCode := event.ErrorCode
	if errorCode == "" {
		errorCode = "none"
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishTotal[publishMetricKey{PackageName: event.PackageName, Outcome: event.Outcome, ErrorCode: errorCode}]++
}

func (m *RegistryMetrics) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, m.PrometheusText())
}

func (m *RegistryMetrics) PrometheusText() string {
	if m == nil {
		return ""
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	var b strings.Builder
	b.WriteString("# HELP docs_registry_http_requests_total Total docs-registry HTTP requests by route class, method, and status.\n")
	b.WriteString("# TYPE docs_registry_http_requests_total counter\n")
	requestKeys := make([]requestMetricKey, 0, len(m.requestsTotal))
	for key := range m.requestsTotal {
		requestKeys = append(requestKeys, key)
	}
	sort.Slice(requestKeys, func(i, j int) bool {
		if requestKeys[i].RouteClass != requestKeys[j].RouteClass {
			return requestKeys[i].RouteClass < requestKeys[j].RouteClass
		}
		if requestKeys[i].Method != requestKeys[j].Method {
			return requestKeys[i].Method < requestKeys[j].Method
		}
		return requestKeys[i].Status < requestKeys[j].Status
	})
	for _, key := range requestKeys {
		fmt.Fprintf(&b, "docs_registry_http_requests_total{route_class=%q,method=%q,status=%q} %d\n",
			key.RouteClass, key.Method, fmt.Sprintf("%d", key.Status), m.requestsTotal[key])
	}

	b.WriteString("# HELP docs_registry_publish_attempts_total Total docs-registry publish attempts by package, outcome, and stable error code.\n")
	b.WriteString("# TYPE docs_registry_publish_attempts_total counter\n")
	publishKeys := make([]publishMetricKey, 0, len(m.publishTotal))
	for key := range m.publishTotal {
		publishKeys = append(publishKeys, key)
	}
	sort.Slice(publishKeys, func(i, j int) bool {
		if publishKeys[i].PackageName != publishKeys[j].PackageName {
			return publishKeys[i].PackageName < publishKeys[j].PackageName
		}
		if publishKeys[i].Outcome != publishKeys[j].Outcome {
			return publishKeys[i].Outcome < publishKeys[j].Outcome
		}
		return publishKeys[i].ErrorCode < publishKeys[j].ErrorCode
	})
	for _, key := range publishKeys {
		fmt.Fprintf(&b, "docs_registry_publish_attempts_total{package=%q,outcome=%q,error_code=%q} %d\n",
			key.PackageName, key.Outcome, key.ErrorCode, m.publishTotal[key])
	}
	return b.String()
}
