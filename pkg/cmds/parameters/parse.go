package parameters

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
	"github.com/pkg/errors"
	"github.com/tj/go-naturaldate"
	"gopkg.in/yaml.v3"
)

type ParseStep struct {
	Source   string
	Value    interface{}
	Metadata map[string]interface{}
}

// ParseParameter parses command line arguments according to the given ParameterDefinition.
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
// The ParameterDefinition specifies the expected type and how to parse the arguments:
//
//   - ParameterTypeString: parsed from a single string value
//   - ParameterTypeInteger, ParameterTypeFloat, ParameterTypeBool: parsed from a single value
//   - ParameterTypeStringList, ParameterTypeIntegerList, ParameterTypeFloatList: parsed from multiple values
//   - ParameterTypeFile: load file contents into a FileData struct
//   - ParameterTypeFileList: load multiple files into []*FileData
//   - ParameterTypeChoice, ParameterTypeChoiceList: validated against allowed choices
//   - ParameterTypeKeyValue: parsed from colon separated strings or files
//   - ParameterTypeObjectListFromFile, ParameterTypeObjectListFromFiles: deserialized object lists from JSON/YAML files
//   - ParameterTypeObjectFromFile: deserialized a single object from a JSON/YAML file
//   - ParameterTypeStringFromFile, ParameterTypeStringFromFiles: load file contents as strings
//   - ParameterTypeStringListFromFile, ParameterTypeStringListFromFiles: load file lines as a string list
//   - ParameterTypeDate: parsed into time.Time
//
// The parsing logic depends on the Type in the ParameterDefinition.
//
// ## ParameterType -> Bype mappings
//
// ParameterTypeString -> string
// ParameterTypeInteger -> int
// ParameterTypeFloat -> float64
// ParameterTypeBool -> bool
// ParameterTypeStringList -> []string
// ParameterTypeIntegerList -> []int
// ParameterTypeFloatList -> []float64
// ParameterTypeChoice -> string
// ParameterTypeChoiceList -> []string
// ParameterTypeDate -> time.Time
// ParameterTypeFile -> *FileData
// ParameterTypeFileList -> []*FileData
// ParameterTypeObjectListFromFile -> []interface{}
// ParameterTypeObjectFromFile -> map[string]interface{}
// ParameterTypeStringFromFile -> string
// ParameterTypeStringFromFiles -> string
// ParameterTypeStringListFromFile -> []string
// ParameterTypeStringListFromFiles -> []string
// ParameterTypeKeyValue -> map[string]interface{}
//
// TODO(manuel, 2023-12-22) We should provide the parsing context from higher up here, instead of just calling it strings
func (p *ParameterDefinition) ParseParameter(v []string, options ...ParseStepOption) (*ParsedParameter, error) {
	ret := &ParsedParameter{
		ParameterDefinition: p,
	}

	if len(v) == 0 {
		if p.Required {
			return nil, errors.Errorf("Argument %s not found", p.Name)
		} else {
			options_ := append(options, WithParseStepSource("default"))
			ret.Set(p.Default, options_...)
			return ret, nil
		}
	}

	var v_ interface{}

	switch p.Type {
	case ParameterTypeString:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single string", p.Name)
		}
		v_ = v[0]
	case ParameterTypeInteger:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single integer", p.Name)
		}
		i, err := strconv.Atoi(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
		}
		v_ = i
	case ParameterTypeFloat:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single float", p.Name)
		}
		f, err := strconv.ParseFloat(v[0], 64)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as float", p.Name)
		}
		v_ = f
	case ParameterTypeStringList:
		v_ = v
	case ParameterTypeIntegerList:
		ints := make([]int, 0)
		for _, arg := range v {
			i, err := strconv.Atoi(arg)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
			}
			ints = append(ints, i)
		}
		v_ = ints

	case ParameterTypeBool:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single boolean", p.Name)
		}
		switch {
		case v[0] == "on":
			v_ = true
		case v[0] == "off":
			v_ = false
		default:
			b, err := strconv.ParseBool(v[0])
			if err != nil {
				return nil, errors.Wrapf(err, "Could not parse argument %s as bool", p.Name)
			}
			v_ = b
		}

	case ParameterTypeChoice:
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

	case ParameterTypeChoiceList:
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

	case ParameterTypeDate:
		parsedDate, err := ParseDate(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as date", p.Name)
		}
		v_ = parsedDate

	case ParameterTypeFile:
		v__, err := GetFileData(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read file %s", v[0])
		}
		v_ = v__

	case ParameterTypeFileList:
		ret := []interface{}{}
		for _, fileName := range v {
			v, err := GetFileData(fileName)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not read file %s", fileName)
			}
			ret = append(ret, v)
		}
		v_ = ret

	case ParameterTypeObjectListFromFiles:
		fallthrough
	case ParameterTypeObjectListFromFile:
		ret_ := []interface{}{}
		for _, fileName := range v {
			l, err := parseFromFileName(fileName, p, options...)
			if err != nil {
				return nil, err
			}
			lObj, ok := l.Value.([]interface{})
			if !ok {
				return nil, errors.Errorf("Could not parse file %s as list of objects", fileName)
			}

			ret_ = append(ret_, lObj...)
		}
		v_ = ret_

	case ParameterTypeObjectFromFile:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single file name", p.Name)
		}
		v__, err := parseFromFileName(v[0], p, options...)
		if err != nil {
			return nil, err
		}
		v_ = v__

	case ParameterTypeStringFromFile:
		fallthrough
	case ParameterTypeStringFromFiles:
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

	case ParameterTypeStringListFromFiles:
		fallthrough
	case ParameterTypeStringListFromFile:
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

	case ParameterTypeKeyValue:
		switch {
		case len(v) == 0:
			v_ = p.Default

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
			v_ = v__

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

	case ParameterTypeFloatList:
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

	options_ := append(options, WithParseStepMetadata(map[string]interface{}{
		"parsed-strings": v,
	}))
	ret.Set(v_, options_...)
	return ret, nil
}

func parseFromFileName(fileName string, p *ParameterDefinition, options ...ParseStepOption) (*ParsedParameter, error) {
	ret := &ParsedParameter{
		ParameterDefinition: p,
	}
	if fileName == "" {
		ret.Set(p.Default, append(options, WithParseStepSource("default"))...)
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
func (p *ParameterDefinition) ParseFromReader(
	f io.Reader, filename string,
	options ...ParseStepOption,
) (*ParsedParameter, error) {
	ret := &ParsedParameter{
		ParameterDefinition: p,
	}

	options = append(options, WithParseStepMetadata(map[string]interface{}{
		"filename":    filename,
		"parsed-type": p.Type,
	}))

	var err error
	//exhaustive:ignore
	switch p.Type {
	case ParameterTypeStringListFromFiles:
		fallthrough
	case ParameterTypeStringListFromFile:
		ret_ := make([]string, 0)

		// check for json
		if strings.HasSuffix(filename, ".json") {
			err = json.NewDecoder(f).Decode(&ret_)
			if err != nil {
				return nil, err
			}
			ret.Set(ret_, options...)
			return ret, nil
		} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			err = yaml.NewDecoder(f).Decode(&ret_)
			if err != nil {
				return nil, err
			}
			ret.Set(ret_, options...)
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
				ret.Set(ret_, options...)
			} else {
				ret.Set(ret_, options...)
			}
		}

		return ret, nil

	case ParameterTypeObjectFromFile:
		object := interface{}(nil)
		if filename == "-" || strings.HasSuffix(filename, ".json") {
			err = json.NewDecoder(f).Decode(&object)
			ret.Set(object, options...)
		} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			err = yaml.NewDecoder(f).Decode(&object)
			ret.Set(object, options...)
		} else if strings.HasSuffix(filename, ".csv") || strings.HasSuffix(filename, ".tsv") {
			objects, err := parseObjectListFromCSV(f, filename)
			if err != nil {
				return nil, err
			}
			if len(objects) != 1 {
				return nil, errors.Errorf("File %s does not contain exactly one object", filename)
			}
			object = objects[0]
			ret.Set(object, options...)
		} else {
			return nil, errors.Errorf("Could not parse file %s: unknown file type", filename)
		}

		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse file %s", filename)
		}

		return ret, nil

	case ParameterTypeObjectListFromFiles:
		fallthrough
	case ParameterTypeObjectListFromFile:
		return p.parseObjectListFromReader(f, filename, options...)

	case ParameterTypeKeyValue:
		ret_ := interface{}(nil)
		if filename == "-" || strings.HasSuffix(filename, ".json") {
			err = json.NewDecoder(f).Decode(&ret_)
			if err != nil {
				return nil, err
			}
			ret.Set(ret_, options...)
		} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			err = yaml.NewDecoder(f).Decode(&ret_)
			if err != nil {
				return nil, err
			}
			ret.Set(ret_, options...)
		} else if strings.HasSuffix(filename, ".csv") || strings.HasSuffix(filename, ".tsv") {
			objects, err := parseObjectListFromCSV(f, filename)
			if err != nil {
				return nil, err
			}
			if len(objects) != 1 {
				return nil, errors.Errorf("File %s does not contain exactly one object", filename)
			}
			ret.Set(objects[0], options...)
		} else {
			return nil, errors.Errorf("Could not parse file %s: unknown file type", filename)
		}

		return ret, nil

	case ParameterTypeStringFromFiles:
		fallthrough
	case ParameterTypeStringFromFile:
		var b bytes.Buffer
		_, err := io.Copy(&b, f)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read from stdin")
		}
		ret.Set(b.String(), options...)
		return ret, nil

	default:
		return nil, errors.New("Cannot parse from file for this parameter type")
	}
}

func (p *ParameterDefinition) parseObjectListFromReader(
	f io.Reader,
	filename string,
	options ...ParseStepOption,
) (*ParsedParameter, error) {
	ret := &ParsedParameter{
		ParameterDefinition: p,
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
		ret.Set(objectList, options...)
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

		ret.Set(objectList, options...)
	} else if strings.HasSuffix(filename, ".csv") || strings.HasSuffix(filename, ".tsv") {
		var err error
		objectList, err = parseObjectListFromCSV(f, filename)
		if err != nil {
			return nil, err
		}

		ret.Set(objectList, options...)
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

// GatherParametersFromMap gathers parameter values from a map based on the provided ParameterDefinitions.
//
// For each ParameterDefinition, it checks if a matching value is present in the map:
//
// - If the parameter is missing and required, an error is returned.
// - If the parameter is missing and optional, the default value is used.
// - If the value is provided, it is validated against the definition.
//
// Values are looked up by parameter name, as well as short flag if provided.
//
// The returned map contains the gathered parameter values, with defaults filled in
// for any missing optional parameters.
func GatherParametersFromMap(
	m map[string]interface{},
	ps *ParameterDefinitions,
	onlyProvided bool,
	options ...ParseStepOption,
) (*ParsedParameters, error) {
	ret := NewParsedParameters()

	for v := ps.Oldest(); v != nil; v = v.Next() {
		name, p := v.Key, v.Value

		parsed := &ParsedParameter{
			ParameterDefinition: p,
		}

		v_, ok := m[name]
		if !ok {
			if p.ShortFlag != "" {
				v_, ok = m[p.ShortFlag]
			}
			if onlyProvided {
				continue
			}
			if !ok {
				parsed.Set(p.Default)
				ret.Set(name, parsed)
				continue
			}
		}
		options_ := append(options, WithParseStepMetadata(map[string]interface{}{
			"map-value": v_,
		}))

		err := p.CheckValueValidity(v_)
		if err != nil {
			return nil, errors.Wrapf(err, "Invalid value for parameter %s", name)
		}
		// NOTE(manuel, 2023-12-22) We might want to pass in that name instead of just saying from-map
		parsed.Set(v_, options_...)
		ret.Set(name, parsed)
	}

	return ret, nil
}
