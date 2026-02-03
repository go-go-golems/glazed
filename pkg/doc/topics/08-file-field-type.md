---
Title: Working with Files in Commands
Slug: file-fields
Short: Describes how to work with file inputs in command fields.
Topics:
- Commands
- Fields
IsTemplate: false
IsTopLevel: false
ShowPerDefault: false
SectionType: GeneralTopic
---

Glazed provides two new field types `file` and `fileList` that allow passing file paths which will be automatically loaded and parsed.

## The `FileData` structure

File fields are parsed into a single or a list of `FileData` structures which can then be accessed from within a template.

```go
package fields

const (
	Unknown FileType = "Unknown"
	JSON    FileType = "JSON"
	YAML    FileType = "YAML"
	CSV     FileType = "CSV"
	TEXT    FileType = "TEXT"
)

type FileData struct {
	Content          string
	ParsedContent    interface{}
	ParseError       error
	RawContent       []byte
	StringContent    string
	IsList           bool
	IsObject         bool
	BaseName         string
	Extension        string
	FileType         FileType
	Path             string
	RelativePath     string
	AbsolutePath     string
	Size             int64
	LastModifiedTime time.Time
	Permissions      os.FileMode
	IsDirectory      bool
}
```


### Using FileData in Templates

The `FileData` structure returned by the `file` and `fileList` fields can be easily used inside templates.

For example, to loop through a list of files:

```
{{ range .input_files }}

Filename: {{ .Name }}
Content: {{ .Content }}

{{ end }}
```

Since Glazed parses the file content based on the extension, you can access nested fields of parsed structures:

```
{{ range .json_files }}

Name: {{ .Name }}  
User: {{ .ParsedContent.user.name }}

{{ end }}
```

This will print the `name` field from the `user` object parsed from the JSON content.

Similar direct access is possible for YAML, CSV and other parsed file types.

This makes it very easy to build templates that can ingest a variety of file types.


## The `file` Type

The `file` field type allows passing a single file path, which will be read and parsed into a `FileData` structure.

For example:

```
  - name: file
    type: file
    help: Input file for generating JSON schema
    required: true
```

Then running 

```
command --file filename
```

This will read the content of `<filename>`, attempt to parse it based on the file extension, and return a `FileData` object that contains:

- The raw content as bytes
- The parsed content (e.g. JSON/YAML/CSV parsed into native Go structures)
- File metadata like name, path, size etc.

The command implementation can then easily access both the raw bytes/string content as well as the parsed representation.

## The `fileList` Field

Similarly, `fileList` allows passing multiple file paths, which will be turned into a list of `FileData` objects:

For example:

```
  - name: input_files
    type: fileList
    help: Input files for generating JSON schema
    required: true
```

Then running: 

```
command --input-files <file1>,<file2> ...
```

This makes it easy to load and process multiple files in a single command.

The command receives a slice of `FileData` instances to work with.
