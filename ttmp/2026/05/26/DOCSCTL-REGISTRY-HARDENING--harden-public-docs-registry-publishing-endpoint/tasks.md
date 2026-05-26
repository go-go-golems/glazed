# Tasks

## TODO

- [ ] Phase 2: Add request ID, structured access logs, explicit production body limit, publish concurrency limit, and basic rate limiting.
- [ ] Phase 3: Add immutable release-version policy with idempotent same-SHA retries and quota checks.
- [ ] Phase 4: Enrich `PublisherIdentity` with non-sensitive JWT claims and emit publish audit events.
- [ ] Phase 5: Add metrics and alerting, or document log-based alerts if metrics are not available yet.
- [ ] Phase 6: Add local and production-safe negative auth proof cases.
- [ ] Phase 7: Roll out to production and capture validation evidence.

## DONE

- [x] Create `DOCSCTL-REGISTRY-HARDENING` ticket.
- [x] Write intern-oriented analysis/design/implementation guide.
- [x] Relate core source, workflow, Terraform, and k3s files to the design doc.
- [x] Upload guide to reMarkable.
