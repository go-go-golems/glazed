package html

import (
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"strings"
	"testing"
)

type TestProcessor struct {
	Objects   []map[string]interface{}
	formatter formatters.OutputFormatter
}

func NewTestProcessor() *TestProcessor {
	return &TestProcessor{
		formatter: &TestFormatter{},
	}
}

type TestFormatter struct{}

func (t TestFormatter) AddRow(row types.Row) {
}

func (t TestFormatter) SetColumnOrder(columnOrder []types.FieldName) {
}

func (t TestFormatter) AddTableMiddleware(m middlewares.TableMiddleware) {
}

func (t TestFormatter) AddTableMiddlewareInFront(m middlewares.TableMiddleware) {
}

func (t TestFormatter) AddTableMiddlewareAtIndex(i int, m middlewares.TableMiddleware) {
}

func (t TestFormatter) GetTable() (*types.Table, error) {
	return nil, errors.New("not implemented")
}

func (t TestFormatter) Output() (string, error) {
	return "", nil
}

func (t *TestProcessor) ProcessInputObject(obj map[string]interface{}) error {
	t.Objects = append(t.Objects, obj)
	return nil
}

func (t *TestProcessor) OutputFormatter() formatters.OutputFormatter {
	return nil
}

func TestSimpleHeaderParse(t *testing.T) {
	gp := NewTestProcessor()
	html_ := "<h1>Header</h1>"

	doc, err := html.Parse(strings.NewReader(html_))
	require.NoError(t, err)

	hsp := NewHTMLHeadingSplitParser(gp, []string{})

	n, err := hsp.ProcessNode(doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 1, len(gp.Objects))
	assert.Equal(t, "Header", gp.Objects[0]["Title"])
	assert.Equal(t, "h1", gp.Objects[0]["Tag"])
}

func TestTwoHeadersParse(t *testing.T) {
	gp := NewTestProcessor()
	html_ := "<h1>Header</h1><h2>Subheader</h2>"

	doc, err := html.Parse(strings.NewReader(html_))
	require.NoError(t, err)

	hsp := NewHTMLHeadingSplitParser(gp, []string{})

	n, err := hsp.ProcessNode(doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))
	assert.Equal(t, "Header", gp.Objects[0]["Title"])
	assert.Equal(t, "h1", gp.Objects[0]["Tag"])
	assert.Equal(t, "Subheader", gp.Objects[1]["Title"])
	assert.Equal(t, "h2", gp.Objects[1]["Tag"])
}

func TestTwoHeadersBody(t *testing.T) {
	gp := NewTestProcessor()
	// go:filetype html
	html_ := `
<html>
	<head>
		<title>Foobar</title>
	</head>
	<body>
		<h1>Header</h1>
		<h2>Subheader</h2>
	</body>
</html>
`

	doc, err := html.Parse(strings.NewReader(html_))
	require.NoError(t, err)

	hsp := NewHTMLHeadingSplitParser(gp, []string{})

	n, err := hsp.ProcessNode(doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))
	assert.Equal(t, "Header", gp.Objects[0]["Title"])
	assert.Equal(t, "h1", gp.Objects[0]["Tag"])
	assert.Equal(t, "Subheader", gp.Objects[1]["Title"])
	assert.Equal(t, "h2", gp.Objects[1]["Tag"])
}

func TestTwoHeadersSomeTextBody(t *testing.T) {
	gp := NewTestProcessor()

	html_ := `
<html>
	<head>
		<title>Foobar</title>
	</head>
	<body>
		<h1>Header</h1>
		<p>Some text</p>
		<h2>Subheader</h2>
		<p>Some text</p>
	</body>
</html>
`

	doc, err := html.Parse(strings.NewReader(html_))
	require.NoError(t, err)

	hsp := NewHTMLHeadingSplitParser(gp, []string{})

	n, err := hsp.ProcessNode(doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))
	assert.Equal(t, "Header", gp.Objects[0]["Title"])
	assert.Equal(t, "h1", gp.Objects[0]["Tag"])
	assert.Equal(t, "<p>Some text</p>", gp.Objects[0]["Body"])

	assert.Equal(t, "Subheader", gp.Objects[1]["Title"])
	assert.Equal(t, "h2", gp.Objects[1]["Tag"])
	assert.Equal(t, "<p>Some text</p>", gp.Objects[1]["Body"])
}

func TestTwoHeadersSomeTextNodes(t *testing.T) {

	html_ := `
<html>
	<head>
		<title>Foobar</title>
	</head>
	<body>
		<h1>Header</h1>
		<p>Some text</p>
		<p>Some text2</p>
		<p>Some text3</p>
		<h2>Subheader</h2>
		<p>Some text</p>
	</body>
</html>
`

	doc, err := html.Parse(strings.NewReader(html_))
	require.NoError(t, err)

	gp := NewTestProcessor()
	hsp := NewHTMLHeadingSplitParser(gp, []string{})

	n, err := hsp.ProcessNode(doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))
	assert.Equal(t, "Header", gp.Objects[0]["Title"])
	assert.Equal(t, "h1", gp.Objects[0]["Tag"])
	assert.Equal(t, "<p>Some text</p>\n\t\t<p>Some text2</p>\n\t\t<p>Some text3</p>", gp.Objects[0]["Body"])

	assert.Equal(t, "Subheader", gp.Objects[1]["Title"])
	assert.Equal(t, "h2", gp.Objects[1]["Tag"])
	assert.Equal(t, "<p>Some text</p>", gp.Objects[1]["Body"])

	gp = NewTestProcessor()
	hsp = NewHTMLHeadingSplitParser(gp, []string{"p"})

	n, err = hsp.ProcessNode(doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))
	assert.Equal(t, "Header", gp.Objects[0]["Title"])
	assert.Equal(t, "h1", gp.Objects[0]["Tag"])
	assert.Equal(t, "Some text\n\t\tSome text2\n\t\tSome text3", gp.Objects[0]["Body"])

	assert.Equal(t, "Subheader", gp.Objects[1]["Title"])
	assert.Equal(t, "h2", gp.Objects[1]["Tag"])
	assert.Equal(t, "Some text", gp.Objects[1]["Body"])
}

func TestStripTags(t *testing.T) {
	gp := NewTestProcessor()
	html_ := `
<html>
	<head>
		<title>Foobar</title>
	</head>
	<body>
		<h1>Header</h1>
		<p>Some text<b>Foobar</b> <span><strong>Test</strong></span></p>
		<h2>Subheader <strong>Foobar</strong><span>Test</span></h2>
		<p>Some text</p>
	</body>
</html>
`
	doc, err := html.Parse(strings.NewReader(html_))
	require.NoError(t, err)

	hsp := NewHTMLHeadingSplitParser(gp, []string{"span", "p"})

	n, err := hsp.ProcessNode(doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))
	assert.Equal(t, "Header", gp.Objects[0]["Title"])
	assert.Equal(t, "h1", gp.Objects[0]["Tag"])
	assert.Equal(t, "Some text<b>Foobar</b> <strong>Test</strong>", gp.Objects[0]["Body"])

	assert.Equal(t, "Subheader <strong>Foobar</strong>Test", gp.Objects[1]["Title"])
	assert.Equal(t, "h2", gp.Objects[1]["Tag"])
	assert.Equal(t, "Some text", gp.Objects[1]["Body"])
}

func TestSplitOtherTags(t *testing.T) {
	gp := NewTestProcessor()
	html_ := `
<p>Foobar</p>
<p>Test</p>
`
	doc, err := html.Parse(strings.NewReader(html_))
	require.NoError(t, err)

	hsp := NewHTMLSplitParser(gp, []string{"p"}, []string{"p"}, true)

	n, err := hsp.ProcessNode(doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))
	assert.Equal(t, "Foobar", gp.Objects[0]["Title"])
	assert.Equal(t, "p", gp.Objects[0]["Tag"])
	assert.Equal(t, "Test", gp.Objects[1]["Title"])
	assert.Equal(t, "p", gp.Objects[1]["Tag"])
}

func TestSplitOtherTagsWithoutTitle(t *testing.T) {
	gp := NewTestProcessor()
	html_ := `
<p>Foobar</p>
<p>Test</p>
`
	doc, err := html.Parse(strings.NewReader(html_))
	require.NoError(t, err)

	hsp := NewHTMLSplitParser(gp, []string{"p"}, []string{"p"}, false)

	n, err := hsp.ProcessNode(doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))
	assert.Equal(t, "", gp.Objects[0]["Title"])
	assert.Equal(t, "p", gp.Objects[0]["Tag"])
	assert.Equal(t, "Foobar", gp.Objects[0]["Body"])
	assert.Equal(t, "", gp.Objects[1]["Title"])
	assert.Equal(t, "p", gp.Objects[1]["Tag"])
	assert.Equal(t, "Test", gp.Objects[1]["Body"])

	gp = NewTestProcessor()
	hsp = NewHTMLSplitParser(gp, []string{}, []string{"p"}, false)

	n, err = hsp.ProcessNode(doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))
	assert.Equal(t, "", gp.Objects[0]["Title"])
	assert.Equal(t, "p", gp.Objects[0]["Tag"])
	assert.Equal(t, "<p>Foobar</p>", gp.Objects[0]["Body"])
	assert.Equal(t, "", gp.Objects[1]["Title"])
	assert.Equal(t, "p", gp.Objects[1]["Tag"])
	assert.Equal(t, "<p>Test</p>", gp.Objects[1]["Body"])
}
