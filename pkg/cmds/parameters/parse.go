package parameters

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"github.com/araddon/dateparse"
	"github.com/pkg/errors"
	"github.com/tj/go-naturaldate"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

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
//
// The parsing logic depends on the Type in the ParameterDefinition.
func (p *ParameterDefinition) ParseParameter(v []string) (interface{}, error) {
	if len(v) == 0 {
		if p.Required {
			return nil, errors.Errorf("Argument %s not found", p.Name)
		} else {
			return p.Default, nil
		}
	}

	switch p.Type {
	case ParameterTypeString:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single string", p.Name)
		}
		return v[0], nil
	case ParameterTypeInteger:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single integer", p.Name)
		}
		i, err := strconv.Atoi(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
		}
		return i, nil
	case ParameterTypeFloat:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single float", p.Name)
		}
		f, err := strconv.ParseFloat(v[0], 64)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as float", p.Name)
		}
		return f, nil
	case ParameterTypeStringList:
		return v, nil
	case ParameterTypeIntegerList:
		ints := make([]int, 0)
		for _, arg := range v {
			i, err := strconv.Atoi(arg)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
			}
			ints = append(ints, i)
		}
		return ints, nil

	case ParameterTypeBool:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single boolean", p.Name)
		}
		if v[0] == "on" {
			return true, nil
		}
		if v[0] == "off" {
			return false, nil
		}
		b, err := strconv.ParseBool(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as bool", p.Name)
		}
		return b, nil

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
		return choice, nil

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
		return choices, nil

	case ParameterTypeDate:
		parsedDate, err := ParseDate(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as date", p.Name)
		}
		return parsedDate, nil

	case ParameterTypeFile:
		v_, err := GetFileData(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read file %s", v[0])
		}
		return v_, nil

	case ParameterTypeFileList:
		ret := []interface{}{}
		for _, fileName := range v {
			v, err := GetFileData(fileName)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not read file %s", fileName)
			}
			ret = append(ret, v)
		}
		return ret, nil

	case ParameterTypeObjectListFromFiles:
		fallthrough
	case ParameterTypeObjectListFromFile:
		ret := []interface{}{}
		for _, fileName := range v {
			l, err := parseFromFileName(fileName, p)
			if err != nil {
				return nil, err
			}
			lObj, ok := l.([]interface{})
			if !ok {
				return nil, errors.Errorf("Could not parse file %s as list of objects", fileName)
			}

			ret = append(ret, lObj...)
		}
		return ret, nil

	case ParameterTypeObjectFromFile:
		if len(v) > 1 {
			return nil, errors.Errorf("Argument %s must be a single file name", p.Name)
		}
		return parseFromFileName(v[0], p)

	case ParameterTypeStringFromFile:
		fallthrough
	case ParameterTypeStringFromFiles:
		res := strings.Builder{}
		for _, fileName := range v {
			s, err := parseFromFileName(fileName, p)
			if err != nil {
				return nil, err
			}
			sObj, ok := s.(string)
			if !ok {
				return nil, errors.Errorf("Could not parse file %s as string", fileName)
			}
			res.WriteString(sObj)
		}
		return res.String(), nil

	case ParameterTypeStringListFromFiles:
		fallthrough
	case ParameterTypeStringListFromFile:
		res := []string{}
		for _, fileName := range v {
			s, err := parseFromFileName(fileName, p)
			if err != nil {
				return nil, err
			}
			sObj, ok := s.([]string)
			if !ok {
				return nil, errors.Errorf("Could not parse file %s as string list", fileName)
			}
			res = append(res, sObj...)
		}
		return res, nil

	case ParameterTypeKeyValue:
		if len(v) == 0 {
			return p.Default, nil
		}
		ret := map[string]interface{}{}
		if len(v) == 1 && strings.HasPrefix(v[0], "@") {
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

			return p.ParseFromReader(f, templateDataFile)
		} else {
			for _, arg := range v {
				// TODO(2023-02-11): The separator could be stored in the parameter itself?
				// It was configurable before.
				//
				// See https://github.com/go-go-golems/glazed/issues/129
				parts := strings.Split(arg, ":")
				if len(parts) != 2 {
					return nil, errors.Errorf("Could not parse argument %s as key=value pair", arg)
				}
				ret[parts[0]] = parts[1]
			}
		}
		return ret, nil

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
		return floats, nil
	}

	return nil, errors.Errorf("Unknown parameter type %s", p.Type)
}

func parseFromFileName(fileName string, p *ParameterDefinition) (interface{}, error) {
	if fileName == "" {
		return p.Default, nil
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

	ret, err := p.ParseFromReader(f, fileName)
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
func (p *ParameterDefinition) ParseFromReader(f io.Reader, filename string) (interface{}, error) {
	var err error
	//exhaustive:ignore
	switch p.Type {
	case ParameterTypeStringListFromFiles:
		fallthrough
	case ParameterTypeStringListFromFile:
		ret := make([]string, 0)

		// check for json
		if strings.HasSuffix(filename, ".json") {
			err = json.NewDecoder(f).Decode(&ret)
			if err != nil {
				return nil, err
			}
			return ret, nil
		} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			err = yaml.NewDecoder(f).Decode(&ret)
			if err != nil {
				return nil, err
			}
			return ret, nil
		} else {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				ret = append(ret, scanner.Text())
			}
			if err = scanner.Err(); err != nil {
				return nil, err
			}
			if strings.HasSuffix(filename, ".csv") || strings.HasSuffix(filename, ".tsv") {
				if len(ret) == 0 {
					return nil, errors.Errorf("File %s does not contain any lines", filename)
				}
				// remove headers
				ret = ret[1:]
			}
		}

		return ret, nil

	case ParameterTypeObjectFromFile:
		object := interface{}(nil)
		if filename == "-" || strings.HasSuffix(filename, ".json") {
			err = json.NewDecoder(f).Decode(&object)
		} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			err = yaml.NewDecoder(f).Decode(&object)
		} else if strings.HasSuffix(filename, ".csv") || strings.HasSuffix(filename, ".tsv") {
			objects, err := parseObjectListFromCSV(f, filename)
			if err != nil {
				return nil, err
			}
			if len(objects) != 1 {
				return nil, errors.Errorf("File %s does not contain exactly one object", filename)
			}
			object = objects[0]
		} else {
			return nil, errors.Errorf("Could not parse file %s: unknown file type", filename)
		}

		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse file %s", filename)
		}

		return object, nil

	case ParameterTypeObjectListFromFiles:
		fallthrough
	case ParameterTypeObjectListFromFile:
		return parseObjectListFromReader(f, filename)

	case ParameterTypeKeyValue:
		ret := interface{}(nil)
		if filename == "-" || strings.HasSuffix(filename, ".json") {
			err = json.NewDecoder(f).Decode(&ret)
		} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
			err = yaml.NewDecoder(f).Decode(&ret)
		} else if strings.HasSuffix(filename, ".csv") || strings.HasSuffix(filename, ".tsv") {
			objects, err := parseObjectListFromCSV(f, filename)
			if err != nil {
				return nil, err
			}
			if len(objects) != 1 {
				return nil, errors.Errorf("File %s does not contain exactly one object", filename)
			}
			ret = objects[0]
		} else {
			return nil, errors.Errorf("Could not parse file %s: unknown file type", filename)
		}

		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse file %s", filename)
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
		return b.String(), nil

	default:
		return nil, errors.New("Cannot parse from file for this parameter type")
	}
}

func parseObjectListFromReader(f io.Reader, filename string) (interface{}, error) {
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
	} else if strings.HasSuffix(filename, ".csv") || strings.HasSuffix(filename, ".tsv") {
		var err error
		objectList, err = parseObjectListFromCSV(f, filename)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.Errorf("Could not parse file %s: unknown file type", filename)
	}

	return objectList, nil
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
	ps map[string]*ParameterDefinition,
	onlyProvided bool,
) (map[string]interface{}, error) {
	ret := map[string]interface{}{}

	for name, p := range ps {
		v, ok := m[name]
		if !ok {
			if p.ShortFlag != "" {
				v, ok = m[p.ShortFlag]
			}
			if onlyProvided {
				continue
			}
			if !ok {
				ret[name] = p.Default
				continue
			}
		}
		err := p.CheckValueValidity(v)
		if err != nil {
			return nil, errors.Wrapf(err, "Invalid value for parameter %s", name)
		}
		ret[name] = v
	}

	return ret, nil
}
