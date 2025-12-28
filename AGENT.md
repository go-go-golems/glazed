# Agent Guidelines for go-go-golems go projects

## Build Commands

- Run a binary in XXX/YYY/FOOO: `go run ./XXX/YYY/FOOO` . Use this instead of build + ./XXX.
- Build: `go build ./...`
- Test: `go test ./...`
- Run single test: `go test ./pkg/path/to/package -run TestName`
- Generate: `go generate ./...`
- Lint: `golangci-lint run -v` or `make lint`
- Format: `go fmt ./...`
- Kill a process using port $PORT: `lsof-who -p $PORT -k`. When building a web server, ALWAYS use this command to kill the process.

## Code Style Guidelines

- **Imports**: Standard library first, third-party next, local packages last
- **Naming**: CamelCase for exported, camelCase for unexported
- **Error Handling**: Return errors with context (pkg/errors), use zerolog for logging
- **Structure**: Follow Go project layout conventions
- **Types**: Define clear interfaces, use struct tags for serialization
- **Commands**: Use Cobra for CLI commands

## Project Structure

- `cmd/`: CLI commands and entry points
- `pkg/`: Library code organized by domain
- `examples/`: Example configurations and usage
- `doc/`: Documentation

<goGuidelines>
- When implementing go interfaces, use the var _ Interface = &Foo{} to make sure the interface is always implemented correctly.
- When building web applications, use htmx, bootstrap and the templ templating language.
- Always use a context argument when appropriate.
- Use cobra for command-line applications.
- Use the "defaults" package name, instead of "default" package name, as it's reserved in go.
- Use github.com/pkg/errors for wrapping errors.
- When starting goroutines, use errgroup.

- Only use the toplevel go.mod, don't create new ones.
- When writing a new experiment / app, add zerolog logging to help debug and figure out how it works, add --log-level flag to set the log level.
</goGuidelines>

<webGuidelines>
- Use bun, react and rtk-query. Use typescript.
- Use bootstrap for styling.
- Store css, html and js in different files in a static directory.
- Use go:embed to serve static files.
- Use templ for go templates, assume I'm running templ generate -watch in the background.
- Always serve static files under /static/ URL paths, never directly under functional paths like /admin/
</webGuidelines>

<debuggingGuidelines>
If me or you the LLM agent seem to go down too deep in a debugging/fixing rabbit hole in our conversations, remind me to take a breath and think about the bigger picture instead of hacking away. Say: "I think I'm stuck, let's TOUCH GRASS".  IMPORTANT: Don't try to fix errors by yourself more than twice in a row. Then STOP. Don't do anything else.

</debuggingGuidelines>

<generalGuidelines>
Don't add backwards compatibility layers unless explicitly asked.

If it looks like your edits aren't applied, stop immediately and say "STOPPING BECAUSE EDITING ISN'T WORKING".

Run the format_file tool at the end of each response.
</generalGuidelines>
