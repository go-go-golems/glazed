Output doc/gifs/demo.gif

Set FontSize 14
Set Width 800
Set Height 600

Set TypingSpeed 0.01

Type "glaze json misc/test-data/[123].json" Enter
Sleep 1200ms

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

Type "glaze json --input-is-array misc/test-data/rows.json" Enter Sleep 0.1
Sleep 1200ms

Type "glaze yaml misc/test-data/[123].yaml" Enter Sleep 0.1
Sleep 1200ms
