Output doc/gifs/demo.gif

Set FontSize 14
Set Width 600
Set Height 400

Set TypingSpeed 0.01

Type "glaze json misc/test-data/*.json" Enter
Sleep 500ms

Type "glaze json misc/test-data/*.json \" Enter Sleep 0.1
Type "    --output csv " Enter
Sleep 500ms

Type "glaze json misc/test-data/*.json \" Enter Sleep 0.1
Type "    --table-format markdown | glow - " Enter
Sleep 500ms

Type "glaze json misc/test-data/2.json \" Enter Sleep 0.1
Type "    --output json " Enter
Sleep 500ms

Type "glaze json misc/test-data/2.json \" Enter Sleep 0.1
Type "    --output json --flatten" Enter
Sleep 500ms

Type "glaze json misc/test-data/*.json \" Enter Sleep 0.1
Type "    --fields c,b,a \" Enter Sleep 0.1
Type "    --table-format markdown | glow - " Enter
Sleep 500ms

Type "glaze json misc/test-data/*.json \" Enter Sleep 0.1
Type "    --filter d.e" Enter Sleep 0.1
Sleep 500ms

