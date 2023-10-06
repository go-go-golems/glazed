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

	actualYAML := Clean(inputYAML, false)
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

	actualYAML := Clean(inputYAML, false)
	assert.Equal(t, expectedYAML, actualYAML)

	inputYAML = "```\n" + `---
a: 1
b: 2
c: 3` + "\n```"

	actualYAML = Clean(inputYAML, false)
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

	actualYAML := Clean(inputYAML, false)
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

	actualYAML := Clean(inputYAML, false)
	assert.Equal(t, expectedYAML, actualYAML)

	inputYAML = `---
a: 1
b: 'foobar: blabla'
`

	expectedYAML = `---
a: 1
b: 'foobar: blabla'
`

	actualYAML = Clean(inputYAML, false)
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

	actualYAML = Clean(inputYAML, false)
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

	actualYAML := Clean(inputYAML, false)
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

	actualYAML := Clean(inputYAML, false)
	assert.Equal(t, expectedYAML, actualYAML)
}

func TestQuotedStrings(t *testing.T) {
	inputYAML := `
title: "Evergreen Trees - The Tree Center"
meta_title: "Buy Evergreen Trees Online - Create Year-Round Privacy and Beauty | The Tree Center"
meta_description: "Shop a variety of evergreen trees that stay green all year at The Tree Center. Create year-round privacy and beauty with our selection of evergreen trees. Find the perfect evergreen for your garden today!"
`

	expectedYAML := `
title: "Evergreen Trees - The Tree Center"
meta_title: "Buy Evergreen Trees Online - Create Year-Round Privacy and Beauty | The Tree Center"
meta_description: "Shop a variety of evergreen trees that stay green all year at The Tree Center. Create year-round privacy and beauty with our selection of evergreen trees. Find the perfect evergreen for your garden today!"
`

	actualYAML := Clean(inputYAML, false)
	assert.Equal(t, expectedYAML, actualYAML)
}
