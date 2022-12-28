Output doc/gifs/demo.gif

Set FontSize 14
Set Width 800
Set Height 600

Set TypingSpeed 0.01

Type "glaze json misc/test-data/[123].json" Enter
Sleep 1200ms

Type "glaze help json" Enter
Sleep 2000ms

Type "glaze help json --examples" Enter
Sleep 2000ms

Type "glaze json misc/test-data/[123].json \" Enter Sleep 0.1
Type "    --output csv " Enter
Sleep 1200ms

Type "glaze json misc/test-data/[123].json \" Enter Sleep 0.1
Type "    --table-format markdown | glow - " Enter
Sleep 1200ms

Type "glaze json misc/test-data/2.json \" Enter Sleep 0.1
Type "    --output json " Enter
Sleep 1200ms

Type "glaze json misc/test-data/2.json \" Enter Sleep 0.1
Type "    --output json --flatten" Enter
Sleep 1200ms

Type "glaze json misc/test-data/[123].json \" Enter Sleep 0.1
Type "    --fields c,b,a \" Enter Sleep 0.1
Type "    --table-format markdown | glow - " Enter
Sleep 1200ms

Type "glaze json misc/test-data/[123].json \" Enter Sleep 0.1
Type "    --filter d.e" Enter Sleep 0.1
Sleep 1200ms

Type "glaze json --input-is-array misc/test-data/rows.json \" Enter Sleep 0.1
Type "    --output yaml" Enter Sleep 0.1
Sleep 1200ms

Type "glaze yaml misc/test-data/[123].yaml" Enter Sleep 0.1
Sleep 1200ms

Type "glaze help templates" Enter
Sleep 2000ms

Type "glaze help templates --examples" Enter
Sleep 2000ms

Type "glaze json misc/test-data/[123].json \" Enter Sleep 0.1
Type "    --template '{{.a}}-{{.b}}: {{.d.f}}' " Enter
Sleep 1200ms

Type "glaze json misc/test-data/[123].json \" Enter Sleep 0.1
Type "    --use-row-templates \" Enter Sleep 0.1
Type "    --fields a,_0 \" Enter Sleep 0.1
Type "    --template '{{.a}}-{{.b}}: {{.d_f}}' " Enter
Sleep 1200ms

Type "glaze json misc/test-data/[123].json \" Enter Sleep 0.1
Type "    --template-field 'foo:{{.a}}-{{.b}},bar:{{.d_f}}' \" Enter Sleep 0.1
Type "    --use-row-templates --fields a,foo,bar" Enter
Sleep 1200ms

Type "glaze json misc/test-data/[123].json \" Enter Sleep 0.1
Type "    --template-field '@misc/template-field-row.yaml' \" Enter Sleep 0.1
Type "    --use-row-templates  --output markdown" Enter Sleep 0.1
Sleep 1200ms

Type "glaze json misc/test-data/[123].json \" Enter Sleep 0.1
Type "    --template-field '@misc/template-field-object.yaml' \" Enter Sleep 0.1
Type "    --output json" Enter Sleep 0.1
Sleep 1200ms

Type "glaze json misc/test-data/[123].json --select a" Enter
Sleep 1200ms

Type "glaze json misc/test-data/[123].json \" Enter Sleep 0.1
Type "    --select-template '{{.a}}-{{.b}}'"  Enter
Sleep 1200ms

Type "glaze yaml misc/test-data/test.yaml --input-is-array --rename baz:blop" Enter
Sleep 1200ms

Type "glaze yaml misc/test-data/test.yaml --input-is-array \" Enter Sleep 0.1
Type "    --rename-regexp '^(.*)bar:${1}blop'" Enter
Sleep 1200ms

Type "glaze yaml misc/test-data/test.yaml --input-is-array \" Enter Sleep 0.1
Type "    --rename-yaml misc/rename.yaml" Enter
Sleep 1200ms

Sleep 2000ms
