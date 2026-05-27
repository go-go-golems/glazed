package publish

import (
	"log/slog"
	"net/http"
	"time"
)

type publishAuditEvent struct {
	RequestID     string
	PackageName   string
	Version       string
	Status        int
	Outcome       string
	ErrorCode     string
	Duration      time.Duration
	ContentLength int64
	UploadBytes   int64
	SectionCount  int
	SlugCount     int
	SHA256        string
	Identity      *PublisherIdentity
}

func logPublishAudit(r *http.Request, event publishAuditEvent) {
	attrs := []any{
		"request_id", event.RequestID,
		"package", event.PackageName,
		"version", event.Version,
		"status", event.Status,
		"outcome", event.Outcome,
		"error_code", event.ErrorCode,
		"duration_ms", event.Duration.Milliseconds(),
		"content_length", event.ContentLength,
		"upload_bytes", event.UploadBytes,
		"section_count", event.SectionCount,
		"slug_count", event.SlugCount,
		"sha256", event.SHA256,
		"client_ip", clientIP(r),
		"remote_addr", r.RemoteAddr,
		"user_agent", r.UserAgent(),
	}
	if event.Identity != nil {
		attrs = append(attrs,
			"subject", event.Identity.Subject,
			"auth_method", event.Identity.Method,
			"identity_package", event.Identity.PackageName,
			"repository", event.Identity.Repository,
			"repository_id", event.Identity.RepositoryID,
			"workflow_ref", event.Identity.WorkflowRef,
			"job_workflow_ref", event.Identity.JobWorkflowRef,
			"run_id", event.Identity.RunID,
		)
	}
	slog.Info("docs registry publish", attrs...)
}
