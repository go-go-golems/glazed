# New API: Build First Command

This example is a minimal, runnable counterpart to the “Build Your First Glazed Command” tutorial, using the newer façade packages:

- `pkg/cmds/schema`
- `pkg/cmds/fields`
- `pkg/cmds/values`

Run:

```bash
go run ./cmd/examples/new-api-build-first-command list-users --limit 3
go run ./cmd/examples/new-api-build-first-command list-users --limit 3 --output json
go run ./cmd/examples/new-api-build-first-command list-users --name-filter engineering --fields id,name,department
```

