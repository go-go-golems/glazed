package fields

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/pkg/errors"
	"github.com/tj/go-naturaldate"
	"gopkg.in/yaml.v3"
)

type ParseStep struct {
	Source   string
	Value    interface{}
	Metadata map[string]interface{}
}

const SourceDefaults = "defaults"

// ParseParameter parses command line arguments according to the given Definition.
// It returns the parsed parameter value and a non-nil error if parsing failed.
//
// The function takes a list of strings that can be gathered from the command line arguments.
// This is because cobra for example allows slice flags to be passed by reusing the same flag multiple times
// (or by parsing comma-separated values).
//
// If the parameter is required and not provided, an error is returned.
// If the parameter is optional and not provided, the default value is returned.
//
// ## Expected type parsing
//
// The Definition specifies the expected type and how to parse the arguments:
//
//   - TypeString: parsed from a single string value
//   - TypeInteger, TypeFloat, TypeBool: parsed from a single value
//   - TypeStringList, TypeIntegerList, TypeFloatList: parsed from multiple values
//   - TypeFile: load file contents into a FileData struct
//   - TypeFileList: load multiple files into []*FileData
//   - TypeChoice, TypeChoiceList: validated against allowed choices
//   - TypeKeyValue: parsed from colon separated strings or files
//   - TypeObjectListFromFile, TypeObjectListFromFiles: deserialized object lists from JSON/YAML files
//   - TypeObjectFromFile: deserialized a single object from a JSON/YAML file
//   - TypeStringFromFile, TypeStringFromFiles: load file contents as strings
//   - TypeStringListFromFile, TypeStringListFromFiles: load file lines as a string list
//   - TypeDate: parsed into time.Time
//
// The parsing logic depends on the Type in the Definition.
//
// ## Type -> Bype mappings
//
// TypeString -> string
// TypeInteger -> int
// TypeFloat -> float64
// TypeBool -> bool
// TypeStringList -> []string
// TypeIntegerList -> []int
// TypeFloatList -> []float64
// TypeChoice -> string
// TypeChoiceList -> []string
// TypeDate -> time.Time
// TypeFile -> *FileData
// TypeFileList -> []*FileData
// TypeObjectListFromFile -> []interface{}
// TypeObjectFromFile -> map[string]interface{}
// TypeStringFromFile -> string
// TypeStringFromFiles -> string
// TypeStringListFromFile -> []string
// TypeStringListFromFiles -> []string
// TypeKeyValue -> map[string]interface{}
//
// TODO(manuel, 2023-12-22) We should provide the parsing context from higher up here, instead of just calling it strings
func (p *Definition) ParseParameter(v []string, options ...ParseOption) (*ParsedParameter, error) {
	ret := &ParsedParameter{
		Definition: p,
	}

	if len(v) == 0 {
		if p.Required {
			return nil, errors.Errorf("Argument %s not found", p.Name)
		} else {
			if p.Default != nil {
				options_ := append(options, WithSource("default"))
				err := ret.Update(*p.Default, options_...)
				if err != nil {
					return nil, err
				}
			}
			return ret, nil
		}
	}

	var v_ interface{}

	switch p.Type {
	case TypeString:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single string", p.Name)
		}
		v_ = v[0]
	case TypeSecret:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single secret", p.Name)
		}
		v_ = v[0]
	case TypeInteger:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single integer", p.Name)
		}
		i, err := strconv.Atoi(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
		}
		v_ = i
	case TypeFloat:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single float", p.Name)
		}
		f, err := strconv.ParseFloat(v[0], 64)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as float", p.Name)
		}
		v_ = f
	case TypeStringList:
		v_ = v
	case TypeIntegerList:
		ints := make([]int, 0)
		for _, arg := range v {
			i, err := strconv.Atoi(arg)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
			}
			ints = append(ints, i)
		}
		v_ = ints

	case TypeBool:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single boolean", p.Name)
		}
		switch v[0] {
		case "on":
			v_ = true
		case "off":
			v_ = false
		default:
			b, err := strconv.ParseBool(v[0])
			if err != nil {
				return nil, errors.Wrapf(err, "Could not parse argument %s as bool", p.Name)
			}
			v_ = b
		}

	case TypeChoice:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single choice", p.Name)
		}
		choice := v[0]
		found := false
		for _, c := range p.Choices {
			if c == choice {
				found = true
			}
		}
		if !found {
			return nil, errors.Errorf("Argument %s has invalid choice %s", p.Name, choice)
		}
		v_ = choice

	case TypeChoiceList:
		choices := make([]string, 0)
		for _, arg := range v {
			found := false
			for _, c := range p.Choices {
				if c == arg {
					found = true
				}
			}
			if !found {
				return nil, errors.Errorf("Argument %s has invalid choice %s", p.Name, arg)
			}
			choices = append(choices, arg)
		}
		v_ = choices

	case TypeDate:
		parsedDate, err := ParseDate(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as date", p.Name)
		}
		v_ = parsedDate

	case TypeFile:
		v__, err := GetFileData(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read file %s", v[0])
		}
		v_ = v__

	case TypeFileList:
		ret := []interface{}{}
		for _, fileName := range v {
			v, err := GetFileData(fileName)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not read file %s", fileName)
			}
			ret = append(ret, v)
		}
		v_ = ret

	case TypeObjectListFromFiles:
		fallthrough
	case TypeObjectListFromFile:
		ret_ := []map[string]interface{}{}
		for _, fileName := range v {
			l, err := parseFromFileName(fileName, p, options...)
			if err != nil {
				return nil, err
			}
			lObj, ok := cast.CastList2[map[string]interface{}, interface{}](l.Value)
			if !ok {
				return nil, errors.Errorf("Could not parse file %s as list of objects", fileName)
			}

			ret_ = append(ret_, lObj...)
		}
		v_ = ret_

	case TypeObjectFromFile:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single file name", p.Name)
		}
		v__, err := parseFromFileName(v[0], p, options...)
		if err != nil {
			return nil, err
		}
		v_ = v__.Value

	case TypeStringFromFile:
		fallthrough
	case TypeStringFromFiles:
		res := strings.Builder{}
		for _, fileName := range v {
			s, err := parseFromFileName(fileName, p, options...)
			if err != nil {
				return nil, err
			}
			sObj, ok := s.Value.(string)
			if !ok {
				return nil, errors.Errorf("Could not parse file %s as string", fileName)
			}
			res.WriteString(sObj)
		}
		v_ = res.String()

	case TypeStringListFromFiles:
		fallthrough
	case TypeStringListFromFile:
		res := []string{}
		for _, fileName := range v {
			s, err := parseFromFileName(fileName, p, options...)
			if err != nil {
				return nil, err
			}
			sObj, ok := s.Value.([]string)
			if !ok {
				return nil, errors.Errorf("Could not parse file %s as string list", fileName)
			}
			res = append(res, sObj...)
		}
		v_ = res

	case TypeKeyValue:
		switch {
		case len(v) == 0:
			if p.Default == nil {
				return ret, nil
			}
			v_ = *p.Default

		case len(v) == 1 && strings.HasPrefix(v[0], "@"):
			// load from file
			templateDataFile := v[0][1:]

			var f io.Reader
			if templateDataFile == "-" {
				f = os.Stdin
			} else {
				f2, err := os.Open(templateDataFile)
				if err != nil {
					return nil, errors.Wrapf(err, "Could not read file %s", templateDataFile)
				}
				defer func(f2 *os.File) {
					_ = f2.Close()
				}(f2)
				f = f2
			}

			v__, err := p.ParseFromReader(f, templateDataFile)
			if err != nil {
				return nil, err
			}
			v_ = v__.Value

		default:
			ret_ := map[string]interface{}{}
			for _, arg := range v {
				// TODO(2023-02-11): The separator could be stored in the parameter itself?
				// It was configurable before.
				//
				// See https://github.com/go-go-golems/glazed/issues/129
				parts := strings.Split(arg, ":")
				if len(parts) != 2 {
					return nil, errors.Errorf("Could not parse argument %s as key=value pair", arg)
				}
				ret_[parts[0]] = parts[1]
			}
			v_ = ret_
		}

	case TypeFloatList:
		floats := make([]float64, 0)
		for _, arg := range v {
			// parse to float
			f, err := strconv.ParseFloat(arg, 64)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
			}
			floats = append(floats, f)
		}
		v_ = floats

	default:
		return nil, errors.Errorf("Unknown parameter type %s", p.Type)
	}

	options_ := append(options, WithMetadata(map[string]interface{}{
		"parsed-strings": v,
	}))
	err := ret.Update(v_, options_...)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func parseFromFileName(fileName string, p *Definition, options ...ParseOption) (*ParsedParameter, error) {
	ret := &ParsedParameter{
		Definition: p,
	}
	if fileName == "" {
		if p.Default != nil {
			err := ret.Update(*p.Default, append(options, WithSource("default"))...)
			if err != nil {
				return nil, err
			}
		}
		return ret, nil
	}
	var f io.Reader
	if fileName == "-" {
		f = os.Stdin
	} else {
		f2, err := os.Open(fileName)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read file %s", fileName)
		}
		defer func(f2 *os.File) {
			_ = f2.Close()
		}(f2)

		f = f2
	}

	ret, err := p.ParseFromReader(f, fileName, options...)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read file %s", fileName)
	}

	return ret, nil
}

func parseObjectListFromCSV(f io.Reader, filename string) ([]interface{}, error) {
	csvReader := csv.NewReader(f)
	csvReader.FieldsPerRecord = -1
	csvReader.TrimLeadingSpace = true

	// check TSV
	if strings.HasSuffix(filename, ".tsv") {
		csvReader.Comma = '\t'
	}

	csvData, err := csvReader.ReadAll()
	if err != nil {
		return nil, errors.Wrapf(err, "Could not parse file %s", filename)
	}

	// if the file is entirely empty, return an empty list
	if len(csvData) == 0 {
		return []interface{}{}, nil
	}

	// check we have both headers and more than one line
	if len(csvData) < 2 {
		return nil, errors.Errorf("File %s does not contain a header line", filename)
	}

	// parse headers
	headers := csvData[0]
	// check we have at least one header
	if len(headers) == 0 {
		return nil, errors.Errorf("File %s does not contain a header line", filename)
	}

	// parse data
	data := make([]interface{}, 0)
	for _, line := range csvData[1:] {
		if len(line) != len(headers) {
			return nil, errors.Errorf("File %s contains a line with a different number of columns than the header", filename)
		}
		lineMap := make(map[string]interface{})
		for i, header := range headers {
			lineMap[header] = line[i]
		}
		data = append(data, lineMap)
	}

	return data, nil
}

// ParseFromReader parses a single element for the type from the reader.
// In the case of parameters taking multiple files, this needs to be called for each file
// and merged at the caller level.
func (p *Definition) ParseFromReader(
	f io.Reader, filename string,
	options ...ParseOption,
) (*ParsedParameter, error) {
	ret := &ParsedParameter{
		Definition: p,
	}

	options = append(options, WithMetadata(map[string]interface{}{
		"filename":    filename,
		"parsed-type": p.Type,
	}))

	var err error
	//exhaustive:ignore
	switch p.Type {
	case TypeStringListFromFiles, TypeStringListFromFile:
		ret_ := make([]string, 0)

		// check for json
		if strings.HasSuffix(filename, ".json") {
			err = json.NewDecoder(f).Decode(&ret_)
			if err != nil {
				return nil, err
			}
			err = ret.Update(ret_, options...)
			if err != nil {
				return nil, err
			}
			return ret, nil
		} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			err = yaml.NewDecoder(f).Decode(&ret_)
			if err != nil {
				return nil, err
			}
			err = ret.Update(ret_, options...)
			if err != nil {
				return nil, err
			}
			return ret, nil
		} else {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				ret_ = append(ret_, scanner.Text())
			}
			if err = scanner.Err(); err != nil {
				return nil, err
			}
			if strings.HasSuffix(filename, ".csv") || strings.HasSuffix(filename, ".tsv") {
				if len(ret_) == 0 {
					return nil, errors.Errorf("File %s does not contain any lines", filename)
				}
				// remove headers
				ret_ = ret_[1:]
				err = ret.Update(ret_, options...)
				if err != nil {
					return nil, err
				}
			} else {
				err = ret.Update(ret_, options...)
				if err != nil {
					return nil, err
				}
			}
		}

		return ret, nil

	case TypeObjectFromFile:
		object := interface{}(nil)
		if filename == "-" || strings.HasSuffix(filename, ".json") {
			err = json.NewDecoder(f).Decode(&object)
			if err != nil {
				return nil, err
			}
			err = ret.Update(object, options...)
			if err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			err = yaml.NewDecoder(f).Decode(&object)
			if err != nil {
				return nil, err
			}
			err = ret.Update(object, options...)
			if err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(filename, ".csv") || strings.HasSuffix(filename, ".tsv") {
			objects, err := parseObjectListFromCSV(f, filename)
			if err != nil {
				return nil, err
			}
			if len(objects) != 1 {
				return nil, errors.Errorf("File %s does not contain exactly one object", filename)
			}
			object = objects[0]
			err = ret.Update(object, options...)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.Errorf("Could not parse file %s: unknown file type", filename)
		}

		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse file %s", filename)
		}

		return ret, nil

	case TypeObjectListFromFiles, TypeObjectListFromFile:
		return p.parseObjectListFromReader(f, filename, options...)

	case TypeKeyValue:
		ret_ := interface{}(nil)
		if filename == "-" || strings.HasSuffix(filename, ".json") {
			err = json.NewDecoder(f).Decode(&ret_)
			if err != nil {
				return nil, err
			}
			err = ret.Update(ret_, options...)
			if err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			err = yaml.NewDecoder(f).Decode(&ret_)
			if err != nil {
				return nil, err
			}
			err = ret.Update(ret_, options...)
			if err != nil {
				return nil, err
			}
		} else if strings.HasSuffix(filename, ".csv") || strings.HasSuffix(filename, ".tsv") {
			objects, err := parseObjectListFromCSV(f, filename)
			if err != nil {
				return nil, err
			}
			if len(objects) != 1 {
				return nil, errors.Errorf("File %s does not contain exactly one object", filename)
			}
			err = ret.Update(objects[0], options...)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, errors.Errorf("Could not parse file %s: unknown file type", filename)
		}

		return ret, nil

	case TypeStringFromFiles, TypeStringFromFile:
		var b bytes.Buffer
		_, err := io.Copy(&b, f)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read from stdin")
		}
		err = ret.Update(b.String(), options...)
		if err != nil {
			return nil, err
		}
		return ret, nil

	default:
		return nil, errors.New("Cannot parse from file for this parameter type")
	}
}

func (p *Definition) parseObjectListFromReader(
	f io.Reader,
	filename string,
	options ...ParseOption,
) (*ParsedParameter, error) {
	ret := &ParsedParameter{
		Definition: p,
	}

	objectList := []interface{}{}
	var object interface{}
	if filename == "-" || strings.HasSuffix(filename, ".json") {
		b, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		err = json.NewDecoder(bytes.NewReader(b)).Decode(&objectList)
		// if err, try again with single object
		if err != nil {
			err = json.NewDecoder(bytes.NewReader(b)).Decode(&object)
			if err != nil {
				return nil, err
			}
			objectList = []interface{}{object}
		}
		err = ret.Update(objectList, options...)
		if err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(filename, ".ndjson") || strings.HasSuffix(filename, ".jsonl") {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			err := json.Unmarshal([]byte(line), &object)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not parse line %s as JSON", line)
			}
			objectList = append(objectList, object)
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		err := ret.Update(objectList, options...)
		if err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		b, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		err = yaml.NewDecoder(bytes.NewReader(b)).Decode(&objectList)
		// if err, try again with single object
		if err != nil {
			err = yaml.NewDecoder(bytes.NewReader(b)).Decode(&object)
			if err != nil {
				return nil, err
			}
			objectList = []interface{}{object}
		}

		err = ret.Update(objectList, options...)
		if err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(filename, ".csv") || strings.HasSuffix(filename, ".tsv") {
		var err error
		objectList, err = parseObjectListFromCSV(f, filename)
		if err != nil {
			return nil, err
		}

		err = ret.Update(objectList, options...)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.Errorf("Could not parse file %s: unknown file type", filename)
	}

	return ret, nil
}

// refTime is used to set a reference time for natural date parsing for unit test purposes
var refTime *time.Time

// ParseDate parses a string into a time.Time based on predefined date formats.
//
// It first tries parsing with dateparse.ParseAny using standard formats.
// If that fails, it tries naturaldate.Parse which handles relative natural language dates.
//
// If both parsing attempts fail, an error is returned.
// The reference time passed to naturaldate.Parse defaults to time.Now().
func ParseDate(value string) (time.Time, error) {
	parsedDate, err := dateparse.ParseLocal(value)
	if err != nil {
		refTime_ := time.Now()
		if refTime != nil {
			refTime_ = *refTime
		}
		parsedDate, err = naturaldate.Parse(value, refTime_)
		if err != nil {
			return time.Time{}, errors.Wrapf(err, "Could not parse date: %s", value)
		}
	}

	return parsedDate, nil
}
