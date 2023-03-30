package yaml

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSimpleYAML(t *testing.T) {
	inputYAML := `---
a: 1
b: 2
c: 3
`
	expectedYAML := `---
a: 1
b: 2
c: 3
`

	actualYAML := Clean(inputYAML)
	assert.Equal(t, expectedYAML, actualYAML)
}

func TestRemoveMarkdown(t *testing.T) {
	inputYAML := "```yaml\n" + `---
a: 1
b: 2
c: 3`

	expectedYAML := `---
a: 1
b: 2
c: 3`

	actualYAML := Clean(inputYAML)
	assert.Equal(t, expectedYAML, actualYAML)

	inputYAML = "```\n" + `---
a: 1
b: 2
c: 3` + "\n```"

	actualYAML = Clean(inputYAML)
	assert.Equal(t, expectedYAML, actualYAML)
}

func TestInvalidStrings(t *testing.T) {
	inputYAML := `---
a: 1
b: foobar: blabla
invalidRightHandSide: &foobar
`

	expectedYAML := `---
a: 1
b: "foobar: blabla"
invalidRightHandSide: "&foobar"
`

	actualYAML := Clean(inputYAML)
	assert.Equal(t, expectedYAML, actualYAML)
}

func TestQuotedString(t *testing.T) {
	inputYAML := `---
a: 1
b: "foobar: blabla"
`

	expectedYAML := `---
a: 1
b: "foobar: blabla"
`

	actualYAML := Clean(inputYAML)
	assert.Equal(t, expectedYAML, actualYAML)

	inputYAML = `---
a: 1
b: 'foobar: blabla'
`

	expectedYAML = `---
a: 1
b: 'foobar: blabla'
`

	actualYAML = Clean(inputYAML)
	assert.Equal(t, expectedYAML, actualYAML)

	inputYAML = `---
- a: 1
  b: |
	foobar: blabla
`

	expectedYAML = `---
- a: 1
  b: |
	foobar: blabla
`

	actualYAML = Clean(inputYAML)
	assert.Equal(t, expectedYAML, actualYAML)
}

func TestNestedYAML(t *testing.T) {
	inputYAML := `---
a: 1
b: 
- 2
- 3
- d: 2
  e: fourbe: courbevoie
  f: |
    - 4
    - 5: 234234
    foobar
  g: 2
`

	expectedYAML := `---
a: 1
b: 
- 2
- 3
- d: 2
  e: "fourbe: courbevoie"
  f: |
    - 4
    - 5: 234234
    foobar
  g: 2
`

	actualYAML := Clean(inputYAML)
	assert.Equal(t, expectedYAML, actualYAML)

}

func TestNestedYAML2(t *testing.T) {

	inputYAML := `---
a: 1
b: 'foobar: blabla'
c: |
   foobar: foobar
d: &lkjsld
`

	expectedYAML := `---
a: 1
b: 'foobar: blabla'
c: |
   foobar: foobar
d: "&lkjsld"
`

	actualYAML := Clean(inputYAML)
	assert.Equal(t, expectedYAML, actualYAML)
}
