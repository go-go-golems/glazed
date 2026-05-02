# Tasks

## TODO

- [x] Add SectionHeading metadata to Go and TypeScript section summary contracts.
- [x] Implement and test backend Markdown heading extraction that ignores fenced code blocks and skips duplicate title headings.
- [x] Expose subsection heading metadata from GET /api/sections for each section summary.
- [x] Add stable heading IDs to Markdown rendering so subsection links can scroll to headings.
- [x] Create NavigationModeToggle component and wire Tree/Search mode state into App.tsx.
- [x] Create DocumentationTree component with document-type groups for GeneralTopic, Example, Application, and Tutorial.
- [x] Add tree search filtering for document titles and subsection headings.
- [x] Wire document node clicks to existing section navigation and subsection clicks to section hash navigation.
- [x] Style the tree/sidebar to match the provided Classic Mac documentation screenshot.
- [x] Add frontend tests for tree building, mode switching, conditional tree/search rendering, and subsection navigation.
- [x] Run Go help tests, frontend Vitest, TypeScript checks, web build, and embedded web asset generation.
- [ ] Smoke test with the existing multi-package tmux setup and document the validation result in the diary.
