package parameters

import (
	"encoding/csv"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type FileType string

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

func GetFileData(filename string) (*FileData, error) {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	contentBytes, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	content := string(contentBytes)

	extension := strings.ToLower(filepath.Ext(filename))
	baseName := filepath.Base(filename)
	relativePath := ""
	// get the absolute path to our working directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get working directory")
	} else {
		relativePath, err = filepath.Rel(currentDir, absPath)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get relative path")
		}
	}

	var parsedContent interface{}
	isList := false
	isObject := false
	var fileType FileType
	var parseError error

	switch extension {
	case ".json":
		fileType = JSON
		err := json.Unmarshal(contentBytes, &parsedContent)
		if err == nil {
			switch parsedContent.(type) {
			case []interface{}:
				isList = true
			case map[string]interface{}:
				isObject = true
			}
		} else {
			parseError = err
		}

	case ".yaml", ".yml":
		fileType = YAML
		err := yaml.Unmarshal(contentBytes, &parsedContent)
		if err == nil {
			switch parsedContent.(type) {
			case []interface{}:
				isList = true
			case map[interface{}]interface{}:
				isObject = true
			}
		} else {
			parseError = err
		}

	case ".csv":
		fileType = CSV
		reader := csv.NewReader(strings.NewReader(content))
		records, err := reader.ReadAll()
		if err == nil {
			isList = true
			parsedContent = records
		} else {
			parseError = err
		}

	default:
		fileType = TEXT
		parsedContent = nil
	}

	return &FileData{
		Content:          content,
		ParsedContent:    parsedContent,
		RawContent:       contentBytes,
		StringContent:    content,
		ParseError:       parseError,
		IsList:           isList,
		IsObject:         isObject,
		BaseName:         baseName,
		Extension:        extension,
		FileType:         fileType,
		Path:             filename,
		RelativePath:     relativePath,
		AbsolutePath:     absPath,
		Size:             info.Size(),
		LastModifiedTime: info.ModTime(),
		Permissions:      info.Mode(),
		IsDirectory:      info.IsDir(),
	}, nil
}
