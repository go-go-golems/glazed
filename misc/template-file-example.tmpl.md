# Counts Rows {{ len .rows }}
{{ range $row := .rows }}
## Row {{.b}}

- c: {{.c}}
- a: {{.a}}
{{ end }}