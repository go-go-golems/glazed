# New API: Dual Mode Command

This example shows a single command implementing both:

- `cmds.BareCommand` (`Run`) for classic, human output
- `cmds.GlazeCommand` (`RunIntoGlazeProcessor`) for structured output

Run:

```bash
go run ./cmd/examples/new-api-dual-mode status
go run ./cmd/examples/new-api-dual-mode status --verbose

# Switch to glaze mode
go run ./cmd/examples/new-api-dual-mode status --with-glaze-output
go run ./cmd/examples/new-api-dual-mode status --with-glaze-output --output json
```

