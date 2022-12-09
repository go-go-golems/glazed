package helpers

import (
	"fmt"
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
}

func quote(s string) string {
	return fmt.Sprintf("`%s`", s)
}
func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	t := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(t, s)
}
