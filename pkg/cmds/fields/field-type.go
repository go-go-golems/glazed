package fields

import "strings"

type Type string

const (
	TypeString Type = "string"
	TypeSecret Type = "secret"

	// TODO(2023-02-13, manuel) Should the "default" of a stringFromFile be the filename, or the string?
	//
	// See https://github.com/go-go-golems/glazed/issues/137

	TypeStringFromFile  Type = "stringFromFile"
	TypeStringFromFiles Type = "stringFromFiles"

	// TypeFile and TypeFileList are a more elaborate version that loads and parses
	// the file content and returns a list of FileData objects (or a single object in the case
	// of TypeFile).
	TypeFile     Type = "file"
	TypeFileList Type = "fileList"

	// TODO(manuel, 2023-09-19) Add some more types and maybe revisit the entire concept of loading things from files
	// - string (potentially from file if starting with @)
	// - string/int/float list from file is another useful type

	TypeObjectListFromFile  Type = "objectListFromFile"
	TypeObjectListFromFiles Type = "objectListFromFiles"
	TypeObjectFromFile      Type = "objectFromFile"
	TypeStringListFromFile  Type = "stringListFromFile"
	TypeStringListFromFiles Type = "stringListFromFiles"

	// TypeKeyValue signals either a string with comma separate key-value options,
	// or when beginning with @, a file with key-value options
	TypeKeyValue Type = "keyValue"

	TypeInteger     Type = "int"
	TypeFloat       Type = "float"
	TypeBool        Type = "bool"
	TypeDate        Type = "date"
	TypeStringList  Type = "stringList"
	TypeIntegerList Type = "intList"
	TypeFloatList   Type = "floatList"
	TypeChoice      Type = "choice"
	TypeChoiceList  Type = "choiceList"
)

// NeedsFileContent returns true if the field type is one that loads one or more files, when provided with the given
// value. This slightly odd API is because some types like TypeKeyValue can be either a string or a file. A
// beginning character of @ indicates a file.
func (p Type) NeedsFileContent(value string) bool {
	//exhaustive:ignore
	switch p {
	case TypeStringFromFile,
		TypeObjectListFromFile,
		TypeObjectFromFile,
		TypeStringListFromFile,
		TypeObjectListFromFiles,
		TypeStringListFromFiles,
		TypeStringFromFiles:
		return true

	case TypeKeyValue:
		return strings.HasPrefix(value, "@")
	default:
		return false
	}
}

// NeedsMultipleFileContent returns true if the field type is one that loads multiple files.
func (p Type) NeedsMultipleFileContent() bool {
	//exhaustive:ignore
	switch p {
	case TypeObjectListFromFiles,
		TypeStringListFromFiles,
		TypeStringFromFiles,
		TypeFileList:
		return true

	default:
		return false
	}
}

func (p Type) IsFile() bool {
	//exhaustive:ignore
	switch p {
	case TypeFile,
		TypeFileList:
		return true

	default:
		return false
	}
}

// IsList returns true if the field has to be parsed from a list of strings,
// not if its value is actually a list.
func (p Type) IsList() bool {
	//exhaustive:ignore
	switch p {
	case TypeObjectListFromFile,
		TypeObjectListFromFiles,
		TypeStringListFromFile,
		TypeStringListFromFiles,
		TypeStringList,
		TypeIntegerList,
		TypeFloatList,
		TypeChoiceList,
		TypeKeyValue,
		TypeFileList:
		return true
	default:
		return false
	}
}

func (p Type) IsObject() bool {
	//exhaustive:ignore
	switch p {
	case TypeObjectFromFile,
		TypeObjectListFromFile,
		TypeObjectListFromFiles:
		return true
	default:
		return false
	}
}

func (p Type) IsKeyValue() bool {
	//exhaustive:ignore
	switch p {
	case TypeKeyValue:
		return true
	default:
		return false
	}
}

func (c Type) IsObjectList() bool {
	//exhaustive:ignore
	switch c {
	case TypeObjectListFromFile,
		TypeObjectListFromFiles:
		return true
	default:
		return false
	}
}
