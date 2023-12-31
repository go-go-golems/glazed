package parameters

import "strings"

type ParameterType string

const (
	ParameterTypeString ParameterType = "string"

	// TODO(2023-02-13, manuel) Should the "default" of a stringFromFile be the filename, or the string?
	//
	// See https://github.com/go-go-golems/glazed/issues/137

	ParameterTypeStringFromFile  ParameterType = "stringFromFile"
	ParameterTypeStringFromFiles ParameterType = "stringFromFiles"

	// ParameterTypeFile and ParameterTypeFileList are a more elaborate version that loads and parses
	// the file content and returns a list of FileData objects (or a single object in the case
	// of ParameterTypeFile).
	ParameterTypeFile     ParameterType = "file"
	ParameterTypeFileList ParameterType = "fileList"

	// TODO(manuel, 2023-09-19) Add some more types and maybe revisit the entire concept of loading things from files
	// - string (potentially from file if starting with @)
	// - string/int/float list from file is another useful type

	ParameterTypeObjectListFromFile  ParameterType = "objectListFromFile"
	ParameterTypeObjectListFromFiles ParameterType = "objectListFromFiles"
	ParameterTypeObjectFromFile      ParameterType = "objectFromFile"
	ParameterTypeStringListFromFile  ParameterType = "stringListFromFile"
	ParameterTypeStringListFromFiles ParameterType = "stringListFromFiles"

	// ParameterTypeKeyValue signals either a string with comma separate key-value options,
	// or when beginning with @, a file with key-value options
	ParameterTypeKeyValue ParameterType = "keyValue"

	ParameterTypeInteger     ParameterType = "int"
	ParameterTypeFloat       ParameterType = "float"
	ParameterTypeBool        ParameterType = "bool"
	ParameterTypeDate        ParameterType = "date"
	ParameterTypeStringList  ParameterType = "stringList"
	ParameterTypeIntegerList ParameterType = "intList"
	ParameterTypeFloatList   ParameterType = "floatList"
	ParameterTypeChoice      ParameterType = "choice"
	ParameterTypeChoiceList  ParameterType = "choiceList"
)

// IsFileLoading returns true if the parameter type is one that loads a file, when provided with the given
// value. This slightly odd API is because some types like ParameterTypeKeyValue can be either a string or a file. A
// beginning character of @ indicates a file.
func (p ParameterType) IsFileLoading(value string) bool {
	//exhaustive:ignore
	switch p {
	case ParameterTypeStringFromFile,
		ParameterTypeObjectListFromFile,
		ParameterTypeObjectFromFile,
		ParameterTypeStringListFromFile,
		ParameterTypeObjectListFromFiles,
		ParameterTypeStringListFromFiles,
		ParameterTypeStringFromFiles,
		ParameterTypeFile,
		ParameterTypeFileList:
		return true

	case ParameterTypeKeyValue:
		return strings.HasPrefix(value, "@")
	default:
		return false
	}
}

// IsList returns true if the parameter has to be parsed from a list of strings,
// not if its value is actually a list.
func (p ParameterType) IsList() bool {
	//exhaustive:ignore
	switch p {
	case ParameterTypeObjectListFromFile,
		ParameterTypeObjectListFromFiles,
		ParameterTypeStringListFromFile,
		ParameterTypeStringListFromFiles,
		ParameterTypeStringList,
		ParameterTypeIntegerList,
		ParameterTypeFloatList,
		ParameterTypeChoiceList,
		ParameterTypeKeyValue,
		ParameterTypeFileList:
		return true
	default:
		return false
	}
}

func (p ParameterType) IsObject() bool {
	//exhaustive:ignore
	switch p {
	case ParameterTypeObjectFromFile,
		ParameterTypeObjectListFromFile,
		ParameterTypeObjectListFromFiles:
		return true
	default:
		return false
	}
}

func (p ParameterType) IsKeyValue() bool {
	//exhaustive:ignore
	switch p {
	case ParameterTypeKeyValue:
		return true
	default:
		return false
	}
}

func (c ParameterType) IsObjectList() bool {
	//exhaustive:ignore
	switch c {
	case ParameterTypeObjectListFromFile,
		ParameterTypeObjectListFromFiles:
		return true
	default:
		return false
	}
}
