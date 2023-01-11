# {{ .data.title }}

Author: {{ .data.author }}

{{ range $row := .rows }}
## Row {{.b}}

- c: {{.c}}
- a: {{.a}}
  {{ end }}