package helpers

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

// These templating helpers have been lifted from cobra, to maintain compatibility
// with the standard cobra formatting for usage and help commands.
//
// Original: https://github.com/spf13/cobra
//
// Copyright 2013-2022 The Cobra Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
//
// 2022-12-03 - Manuel Odendahl - Removed deprecated template functions

// TemplateFuncs provides helpers for the standard cobra usage and help templates
var TemplateFuncs = template.FuncMap{
	"trim":                    strings.TrimSpace,
	"trimRightSpace":          trimRightSpace,
	"trimTrailingWhitespaces": trimRightSpace,
	"rpad":                    rpad,
	"quote":                   quote,
	"stripNewlines":           stripNewlines,

	"toUpper": strings.ToUpper,
	"toLower": strings.ToLower,

	"replace":       replace,
	"replaceRegexp": replaceRegexp,

	"add": add,
	"sub": sub,
	"div": div,
	"mul": mul,

	"parseFloat": parseFloat,
	"parseInt":   parseInt,

	"currency": currency,

	"padLeft":  padLeft,
	"padRight": padRight,

	"bold":          bold,
	"underline":     underline,
	"italic":        italic,
	"strikethrough": strikethrough,
	"code":          code,
	"codeBlock":     codeBlock,
}

func bold(s string) string {
	return fmt.Sprintf("**%s**", s)
}

func underline(s string) string {
	return fmt.Sprintf("__%s__", s)
}

func italic(s string) string {
	return fmt.Sprintf("*%s*", s)
}

func strikethrough(s string) string {
	return fmt.Sprintf("~~%s~~", s)
}

func code(s string) string {
	return fmt.Sprintf("`%s`", s)
}

func codeBlock(s string, lang string) string {
	return fmt.Sprintf("```%s\n%s\n```", lang, s)
}

func padLeft(value string, length int) string {
	return fmt.Sprintf("%*s", length, value)
}

func padRight(value string, length int) string {
	return fmt.Sprintf("%-*s", length, value)
}

func add(a, b interface{}) interface{} {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() + bv.Int()

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() + int64(bv.Uint())

		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) + bv.Float()

		default:
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) + bv.Int()

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() + bv.Uint()

		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) + bv.Float()

		default:
			return nil
		}

	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() + float64(bv.Int())

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() + float64(bv.Uint())

		case reflect.Float32, reflect.Float64:
			return av.Float() + bv.Float()

		default:
			return nil
		}

	default:
		return nil
	}
}

func sub(a, b interface{}) interface{} {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() - bv.Int()

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() - int64(bv.Uint())

		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) - bv.Float()

		default:
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) - bv.Int()

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() - bv.Uint()

		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) - bv.Float()

		default:
			return nil
		}

	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() - float64(bv.Int())

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() - float64(bv.Uint())

		case reflect.Float32, reflect.Float64:
			return av.Float() - bv.Float()

		default:
			return nil
		}

	default:
		return nil
	}
}

func mul(a, b interface{}) interface{} {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() * bv.Int()

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() * int64(bv.Uint())

		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) * bv.Float()

		default:
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) * bv.Int()

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() * bv.Uint()

		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) * bv.Float()

		default:
			return nil
		}

	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() * float64(bv.Int())

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() * float64(bv.Uint())

		case reflect.Float32, reflect.Float64:
			return av.Float() * bv.Float()

		default:
			return nil
		}

	default:
		return nil
	}
}

func div(a, b interface{}) interface{} {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() / bv.Int()

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Int() / int64(bv.Uint())

		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) / bv.Float()

		default:
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int64(av.Uint()) / bv.Int()

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() / bv.Uint()

		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) / bv.Float()

		default:
			return nil
		}

	case reflect.Float32, reflect.Float64:
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Float() / float64(bv.Int())

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Float() / float64(bv.Uint())

		case reflect.Float32, reflect.Float64:
			return av.Float() / bv.Float()

		default:
			return nil
		}

	default:
		return nil
	}
}

func quote(s string) string {
	return fmt.Sprintf("`%s`", s)
}
func trimRightSpace(s string) string {

	return strings.TrimRightFunc(s, unicode.IsSpace)
}

func stripNewlines(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	t := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(t, s)
}

func replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

func replaceRegexp(s, old, new string) string {
	re, err := regexp.Compile(old)
	if err != nil {
		return s
	}
	return re.ReplaceAllString(s, new)
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseInt(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func currency(i interface{}) string {
	iv := reflect.ValueOf(i)

	switch iv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d.00", iv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d.00", iv.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", iv.Float())
	default:
		return ""
	}
}
