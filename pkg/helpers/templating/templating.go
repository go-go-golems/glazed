package templating

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	html "html/template"
	"io"
	"io/fs"
	"math"
	"math/big"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/Masterminds/sprig"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/glazed/pkg/helpers/list"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
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
// 2023-02-02 - Manuel Odendahl - Added a decent amount of new functions, at which point I have a hard time
//                 	              remembering which ones are new and which ones are from cobra.
//                                With the addition of sprig templates, most of the cobra if not all of the cobra
//                                template functions can be removed.

// TemplateFuncs provides helpers for the standard cobra usage and help templates
var TemplateFuncs = template.FuncMap{
	// TODO(manuel, 2023-02-02) A lot of these are now deprecated since we added sprig
	// See #108
	"trim":                    strings.TrimSpace,
	"trimRightSpace":          trimRightSpace,
	"trimTrailingWhitespaces": trimRightSpace,
	"rpad":                    rpad,
	"quote":                   quote,
	"stripNewlines":           stripNewlines,
	"quoteNewlines":           quoteNewlines,

	"toUpper": strings.ToUpper,
	"toLower": strings.ToLower,

	"replaceRegexp": replaceRegexp,

	"add": add,
	"sub": sub,
	"div": div,
	"mul": mul,

	"parseFloat": parseFloat,
	"parseInt":   parseInt,

	"currency": currency,

	"padLeft":   padLeft,
	"padRight":  padRight,
	"padCenter": padCenter,

	"bold":          bold,
	"underline":     underline,
	"italic":        italic,
	"strikethrough": strikethrough,
	"code":          code,
	"codeBlock":     codeBlock,

	"toDate": toDate,

	"toYaml":      toYaml,
	"indentBlock": indentBlock,

	"toUrlParameter": toUrlParameter,

	"styleBold": styleBold,

	// Random functions
	"randomChoice":     randomChoice,
	"randomSubset":     randomSubset,
	"randomPermute":    randomPermute,
	"randomInt":        randomInt,
	"randomFloat":      randomFloat,
	"randomBool":       randomBool,
	"randomString":     randomString,
	"randomStringList": randomStringList,
}

func toDate(s interface{}) (string, error) {
	//exhaustive:ignore
	switch v := s.(type) {
	case string:
		// keep only yyyy-mm-dd
		return strings.Split(v, "T")[0], nil
	case time.Time:
		return v.Format("2006-01-02"), nil
	default:
		return "", errors.Errorf("cannot convert %v to date", v)
	}
}

// toUrlParameter encodes the value as a string that can be passed for url parameter decoding
func toUrlParameter(s interface{}) (string, error) {
	switch v := s.(type) {
	case string:
		return v, nil
	case time.Time:
		return v.Format("2006-01-02"), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%f", v), nil
	case []string:
		return list.SliceToCSV(v), nil
	case []int:
		return list.SliceToCSV(v), nil
	case []int8:
		return list.SliceToCSV(v), nil
	case []int16:
		return list.SliceToCSV(v), nil
	case []int32:
		return list.SliceToCSV(v), nil
	case []int64:
		return list.SliceToCSV(v), nil
	case []uint:
		return list.SliceToCSV(v), nil
	case []uint8:
		return list.SliceToCSV(v), nil
	case []uint16:
		return list.SliceToCSV(v), nil
	case []uint32:
		return list.SliceToCSV(v), nil
	case []uint64:
		return list.SliceToCSV(v), nil
	case []float32:
		return list.SliceToCSV(v), nil
	case []float64:
		return list.SliceToCSV(v), nil
	default:
		return fmt.Sprintf("%v", s), nil
	}
}

func styleBold(s string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}

func toYaml(value interface{}) string {
	var buffer bytes.Buffer

	err := yaml.NewEncoder(&buffer).Encode(value)
	if err != nil {
		return ""
	}

	return buffer.String()
}

func indentBlock(indent int, value string) string {
	var buffer bytes.Buffer

	for _, line := range strings.Split(value, "\n") {
		buffer.WriteString(fmt.Sprintf("%s%s\n", strings.Repeat(" ", indent), line))
	}

	return buffer.String()
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

func padLeft(value string, length_ interface{}) string {
	length, err := strconv.Atoi(fmt.Sprintf("%v", length_))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%*s", length, value)
}

func padRight(value string, length_ interface{}) string {
	length, err := strconv.Atoi(fmt.Sprintf("%v", length_))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%-*s", length, value)
}

func padCenter(value string, length_ interface{}) string {
	length, err := strconv.Atoi(fmt.Sprintf("%v", length_))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%*s%*s", (length+len(value))/2, value, (length-len(value))/2, "")
}

func add(a, b interface{}) interface{} {
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	//exhaustive:ignore
	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		//exhaustive:ignore
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() + bv.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintVal := bv.Uint()
			if uintVal > uint64(9223372036854775807) { // MaxInt64
				return fmt.Sprintf("%d + %d", av.Int(), uintVal)
			}
			return av.Int() + int64(uintVal)
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) + bv.Float()

		default:
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		//exhaustive:ignore
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal := bv.Int()
			if intVal < 0 {
				return fmt.Sprintf("%d + %d", av.Uint(), intVal)
			}
			return av.Uint() + uint64(intVal)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() + bv.Uint()
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) + bv.Float()

		default:
			return nil
		}

	case reflect.Float32, reflect.Float64:
		//exhaustive:ignore
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

	//exhaustive:ignore
	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		//exhaustive:ignore
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() - bv.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintVal := bv.Uint()
			if uintVal > uint64(9223372036854775807) { // MaxInt64
				return fmt.Sprintf("%d - %d", av.Int(), uintVal)
			}
			return av.Int() - int64(uintVal)
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) - bv.Float()

		default:
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		//exhaustive:ignore
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal := bv.Int()
			if intVal < 0 {
				return fmt.Sprintf("%d - %d", av.Uint(), intVal)
			}
			return av.Uint() - uint64(intVal)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() - bv.Uint()
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) - bv.Float()

		default:
			return nil
		}

	case reflect.Float32, reflect.Float64:
		//exhaustive:ignore
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

	//exhaustive:ignore
	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		//exhaustive:ignore
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() * bv.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintVal := bv.Uint()
			if uintVal > uint64(9223372036854775807) { // MaxInt64
				return fmt.Sprintf("%d * %d", av.Int(), uintVal)
			}
			return av.Int() * int64(uintVal)
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) * bv.Float()

		default:
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		//exhaustive:ignore
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal := bv.Int()
			if intVal < 0 {
				return fmt.Sprintf("%d * %d", av.Uint(), intVal)
			}
			return av.Uint() * uint64(intVal)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() * bv.Uint()
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) * bv.Float()

		default:
			return nil
		}

	case reflect.Float32, reflect.Float64:
		//exhaustive:ignore
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

	//exhaustive:ignore
	switch av.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		//exhaustive:ignore
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return av.Int() / bv.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			uintVal := bv.Uint()
			if uintVal > uint64(9223372036854775807) { // MaxInt64
				return fmt.Sprintf("%d / %d", av.Int(), uintVal)
			}
			return av.Int() / int64(uintVal)
		case reflect.Float32, reflect.Float64:
			return float64(av.Int()) / bv.Float()

		default:
			return nil
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		//exhaustive:ignore
		switch bv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			intVal := bv.Int()
			if intVal < 0 {
				return fmt.Sprintf("%d / %d", av.Uint(), intVal)
			}
			return av.Uint() / uint64(intVal)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return av.Uint() / bv.Uint()
		case reflect.Float32, reflect.Float64:
			return float64(av.Uint()) / bv.Float()

		default:
			return nil
		}

	case reflect.Float32, reflect.Float64:
		//exhaustive:ignore
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

func quoteNewlines(s string) string {
	return strings.ReplaceAll(s, "\n", `\n`)
}

func stripNewlines(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}

// rpad adds padding to the right of a string.
func rpad(s string, padding_ interface{}) string {
	padding, ok := cast.CastNumberInterfaceToInt[int](padding_)
	if !ok {
		panic("padding must be an int")
	}

	t := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(t, s)
}

func replaceRegexp(s, old, new_ string) string {
	re, err := regexp.Compile(old)
	if err != nil {
		return s
	}
	return re.ReplaceAllString(s, new_)
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

	//exhaustive:ignore
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

type TemplateExecute interface {
	Execute(wr io.Writer, data any) error
}

func RenderTemplate(tmpl TemplateExecute, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func RenderTemplateString(tmpl string, data interface{}) (string, error) {
	t, err := CreateTemplate("template").Parse(tmpl)
	if err != nil {
		return "", err
	}

	return RenderTemplate(t, data)
}

func RenderHtmlTemplateString(tmpl string, data interface{}) (string, error) {
	t, err := CreateHTMLTemplate("template").Parse(tmpl)
	if err != nil {
		return "", err
	}

	return RenderTemplate(t, data)
}

func RenderTemplateFile(filename string, data interface{}) (string, error) {
	t, err := CreateTemplate("template").ParseFiles(filename)
	if err != nil {
		return "", err
	}
	return RenderTemplate(t, data)
}

func CreateTemplate(name string) *template.Template {
	return template.New(name).
		Funcs(sprig.TxtFuncMap()).
		Funcs(TemplateFuncs)
}

func CreateHTMLTemplate(name string) *html.Template {
	return html.New(name).
		Funcs(sprig.HtmlFuncMap()).
		Funcs(TemplateFuncs)
}

// ParseFS will recursively glob for all the files matching the given patterns,
// and load them into one big template (with sub-templates).
//
// The globs use bmatcuk/doublestar and support ** notation for recursive globbing.
func ParseFS(t *template.Template, f fs.FS, patterns ...string) error {
	listMap := make(map[string]struct{})
	for _, p := range patterns {
		list, err := doublestar.FilepathGlob(p, doublestar.WithFilesOnly())
		if err != nil {
			return err
		}

		for _, l := range list {
			listMap[l] = struct{}{}
		}
	}

	for filename := range listMap {
		b, err := fs.ReadFile(f, filename)
		if err != nil {
			return errors.Wrapf(err, "failed to read template %s", filename)
		}

		_, err = t.New(filename).
			Funcs(sprig.TxtFuncMap()).
			Funcs(TemplateFuncs).
			Parse(string(b))
		if err != nil {
			return errors.Wrapf(err, "failed to parse template %s", filename)
		}
	}

	return nil
}

// ParseHTMLFS will recursively glob for all the files matching the given patterns,
// and load them into one big template (with sub-templates).
// It is the html.Template equivalent of ParseFS.
//
// The globs use bmatcuk/doublestar and support ** notation for recursive globbing.
//
// NOTE(manuel, 2023-04-18) Interestingly, we have a baseDir parameter here but only one pattern
// However, the text.template version supports multiple patterns, but has no basedir. Maybe unify?
func ParseHTMLFS(t *html.Template, f fs.FS, patterns []string, baseDir string) error {
	list := []string{}

	for _, pattern := range patterns {
		pattern = filepath.Join(baseDir, pattern)
		err := doublestar.GlobWalk(f, pattern, func(path string, d fs.DirEntry) error {
			if !strings.HasPrefix(path, baseDir) {
				return nil
			}
			list = append(list, path)
			return nil
		}, doublestar.WithFilesOnly())
		if err != nil {
			return err
		}
	}

	for _, filename := range list {
		b, err := fs.ReadFile(f, filename)
		if err != nil {
			return errors.Wrapf(err, "failed to read template %s", filename)
		}

		// strip baseDir from filename
		filename_ := strings.TrimPrefix(filename, baseDir)
		filename_ = strings.TrimPrefix(filename_, "/")
		_, err = t.New(filename_).
			Funcs(sprig.HtmlFuncMap()).
			Funcs(TemplateFuncs).
			Parse(string(b))
		if err != nil {
			return errors.Wrapf(err, "failed to parse template %s", filename)
		}
	}

	return nil
}

// securePerm generates a cryptographically secure random permutation of integers [0, n).
func securePerm(n int) ([]int, error) {
	if n < 0 {
		return nil, errors.New("n must be non-negative")
	}
	p := make([]int, n)
	for i := 0; i < n; i++ {
		p[i] = i
	}
	for i := n - 1; i > 0; i-- {
		// Generate a secure random index j from [0, i]
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate random index for permutation")
		}
		j := int(nBig.Int64())
		// Swap p[i] and p[j]
		p[i], p[j] = p[j], p[i]
	}
	return p, nil
}

// randomChoice returns a random element from a list
func randomChoice(list interface{}) (interface{}, error) {
	v := reflect.ValueOf(list)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, errors.New("list is not a slice or array")
	}
	if v.Len() == 0 {
		return nil, errors.New("list is empty")
	}

	// Generate cryptographically secure random index
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(v.Len())))
	if err != nil {
		return nil, err
	}
	return v.Index(int(nBig.Int64())).Interface(), nil
}

// randomSubset returns a random subset of size n from a list
func randomSubset(list interface{}, n int) (interface{}, error) {
	v := reflect.ValueOf(list)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, errors.New("input must be a slice or array")
	}
	if v.Len() == 0 {
		return nil, errors.New("input slice is empty")
	}
	if n > v.Len() {
		n = v.Len()
	}
	if n < 0 {
		return nil, errors.New("n must be non-negative")
	}

	// Create a new slice of the same type as the input
	result := reflect.MakeSlice(v.Type(), n, n)

	indices, err := securePerm(v.Len())
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate random subset indices")
	}
	for i := 0; i < n; i++ {
		result.Index(i).Set(v.Index(indices[i]))
	}
	return result.Interface(), nil
}

// randomPermute returns a random permutation of a list
func randomPermute(list interface{}) (interface{}, error) {
	v := reflect.ValueOf(list)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, errors.New("input must be a slice or array")
	}
	if v.Len() == 0 {
		return nil, errors.New("input slice is empty")
	}

	// Create a new slice of the same type as the input
	result := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
	indices, err := securePerm(v.Len())
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate random permutation indices")
	}
	for i := 0; i < v.Len(); i++ {
		result.Index(i).Set(v.Index(indices[i]))
	}
	return result.Interface(), nil
}

// randomInt returns a random integer between min and max (inclusive)
func randomInt(min_, max_ interface{}) int {
	minVal, ok := cast.CastNumberInterfaceToInt[int](min_)
	if !ok {
		return 0
	}
	maxVal, ok := cast.CastNumberInterfaceToInt[int](max_)
	if !ok {
		return 0
	}
	if minVal >= maxVal {
		return minVal
	}

	// Generate cryptographically secure random int
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(maxVal-minVal+1)))
	if err != nil {
		// Fallback in case of error (unlikely)
		return minVal
	}
	return minVal + int(nBig.Int64())
}

// randomFloat returns a random float between min and max
func randomFloat(min_, max_ interface{}) float64 {
	minVal, ok := cast.CastNumberInterfaceToFloat[float64](min_)
	if !ok {
		return 0
	}
	maxVal, ok := cast.CastNumberInterfaceToFloat[float64](max_)
	if !ok {
		return 0
	}
	if minVal >= maxVal {
		return minVal
	}

	// Generate cryptographically secure random float
	var buf [8]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		return minVal
	}
	f := float64(binary.BigEndian.Uint64(buf[:])) / float64(math.MaxUint64)
	return minVal + f*(maxVal-minVal)
}

// randomBool returns a random boolean value
func randomBool() bool {
	n, err := rand.Int(rand.Reader, big.NewInt(2))
	if err != nil {
		return false
	}
	return n.Int64() == 1
}

// randomString returns a random string of the given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	// Generate cryptographically secure random index for charset
	for i := range b {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// In case of error, just use a predictable value - unlikely to happen
			b[i] = charset[0]
			continue
		}
		b[i] = charset[nBig.Int64()]
	}
	return string(b)
}

// randomStringList returns a list of random strings
func randomStringList(count, minLength, maxLength int) []string {
	result := make([]string, count)
	for i := range result {
		length := randomInt(minLength, maxLength)
		result[i] = randomString(length)
	}
	return result
}
