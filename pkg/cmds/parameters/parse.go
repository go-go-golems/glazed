package parameters

import (
	"bufio"
	"bytes"
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
		return v[0], nil
	case ParameterTypeInteger:
		i, err := strconv.Atoi(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
		}
		return i, nil
	case ParameterTypeFloat:
		f, err := strconv.ParseFloat(v[0], 32)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as float", p.Name)
		}
		return float32(f), nil
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
		b, err := strconv.ParseBool(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as bool", p.Name)
		}
		return b, nil

	case ParameterTypeChoice:
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

	case ParameterTypeDate:
		parsedDate, err := ParseDate(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as date", p.Name)
		}
		return parsedDate, nil

	case ParameterTypeObjectListFromFile:
		fallthrough
	case ParameterTypeObjectFromFile:
		fallthrough
	case ParameterTypeStringFromFile:
		fallthrough
	case ParameterTypeStringListFromFile:
		fileName := v[0]
		if fileName == "" {
			return p.Default, nil
		}
		var f io.Reader
		if fileName == "-" {
			f = os.Stdin
		} else {
			f2, err := os.Open(fileName)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not read file %s", v[0])
			}
			defer func(f2 *os.File) {
				_ = f2.Close()
			}(f2)

			f = f2
		}

		ret, err := p.ParseFromReader(f, fileName)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read file %s", v[0])
		}

		return ret, nil

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

func (p *ParameterDefinition) ParseFromReader(f io.Reader, name string) (interface{}, error) {
	var err error
	//exhaustive:ignore
	switch p.Type {
	case ParameterTypeStringListFromFile:
		ret := make([]string, 0)
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			ret = append(ret, scanner.Text())
		}
		if err = scanner.Err(); err != nil {
			return nil, err
		}

		return ret, nil

	case ParameterTypeObjectFromFile:
		object := interface{}(nil)
		if name == "-" || strings.HasSuffix(name, ".json") {
			err = json.NewDecoder(f).Decode(&object)
		} else if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			err = yaml.NewDecoder(f).Decode(&object)
		} else {
			return nil, errors.Errorf("Could not parse file %s: unknown file type", name)
		}

		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse file %s", name)
		}

		return object, nil

	case ParameterTypeObjectListFromFile:
		objectList := []interface{}{}
		if name == "-" || strings.HasSuffix(name, ".json") {
			err = json.NewDecoder(f).Decode(&objectList)
		} else if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			err = yaml.NewDecoder(f).Decode(&objectList)
		} else {
			return nil, errors.Errorf("Could not parse file %s: unknown file type", name)
		}

		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse file %s", name)
		}

		return objectList, nil

	case ParameterTypeKeyValue:
		ret := interface{}(nil)
		if name == "-" || strings.HasSuffix(name, ".json") {
			err = json.NewDecoder(f).Decode(&ret)
		} else if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			err = yaml.NewDecoder(f).Decode(&ret)
		} else {
			return nil, errors.Errorf("Could not parse file %s: unknown file type", name)
		}

		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse file %s", name)
		}
		return ret, nil

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

// refTime is used to set a reference time for natural date parsing for unit test purposes
var refTime *time.Time

func ParseDate(value string) (time.Time, error) {
	parsedDate, err := dateparse.ParseAny(value)
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
