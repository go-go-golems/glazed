package html

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/formatters"
	assert2 "github.com/go-go-golems/glazed/pkg/helpers/assert"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
	"io"
	"strings"
	"testing"
)

type TestProcessor struct {
	Objects   []types.Row
	formatter formatters.OutputFormatter
	processor *middlewares.Processor
}

func NewTestProcessor() *TestProcessor {
	return &TestProcessor{
		formatter: &TestFormatter{},
		processor: middlewares.NewProcessor(),
	}
}

type TestFormatter struct{}

func (t TestFormatter) ContentType() string {
	return "text/plain"
}

func (t TestFormatter) Output(context.Context, *types.Table, io.Writer) error {
	return nil
}

func (t *TestProcessor) ProcessInputObject(ctx context.Context, obj types.Row) error {
	t.Objects = append(t.Objects, obj)
	return nil
}

func (t *TestProcessor) OutputFormatter() formatters.OutputFormatter {
	return nil
}

func (t *TestProcessor) Processor() *middlewares.Processor {
	return t.processor
}

func TestSimpleHeaderParse(t *testing.T) {
	gp := NewTestProcessor()
	html_ := "<h1>Header</h1>"

	doc, err := html.Parse(strings.NewReader(html_))
	require.NoError(t, err)

	hsp := NewHTMLHeadingSplitParser(gp, []string{})

	ctx := context.Background()

	n, err := hsp.ProcessNode(ctx, doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	require.Equal(t, 1, len(gp.Objects))
	assert2.EqualRowValue(t, "Header", gp.Objects[0], "Title")
	assert2.EqualRowValue(t, "h1", gp.Objects[0], "Tag")
}

func TestTwoHeadersParse(t *testing.T) {
	gp := NewTestProcessor()
	html_ := "<h1>Header</h1><h2>Subheader</h2>"

	doc, err := html.Parse(strings.NewReader(html_))
	require.NoError(t, err)

	hsp := NewHTMLHeadingSplitParser(gp, []string{})

	ctx := context.Background()

	n, err := hsp.ProcessNode(ctx, doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))

	assert2.EqualRowValue(t, "Header", gp.Objects[0], "Title")
	assert2.EqualRowValue(t, "h1", gp.Objects[0], "Tag")

	assert2.EqualRowValue(t, "Subheader", gp.Objects[1], "Title")
	assert2.EqualRowValue(t, "h2", gp.Objects[1], "Tag")
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

	ctx := context.Background()
	n, err := hsp.ProcessNode(ctx, doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))

	assert2.EqualRowValue(t, "Header", gp.Objects[0], "Title")
	assert2.EqualRowValue(t, "h1", gp.Objects[0], "Tag")

	assert2.EqualRowValue(t, "Subheader", gp.Objects[1], "Title")
	assert2.EqualRowValue(t, "h2", gp.Objects[1], "Tag")
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

	ctx := context.Background()
	n, err := hsp.ProcessNode(ctx, doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))

	assert2.EqualRowValue(t, "Header", gp.Objects[0], "Title")
	assert2.EqualRowValue(t, "h1", gp.Objects[0], "Tag")
	assert2.EqualRowValue(t, "<p>Some text</p>", gp.Objects[0], "Body")

	assert2.EqualRowValue(t, "Subheader", gp.Objects[1], "Title")
	assert2.EqualRowValue(t, "h2", gp.Objects[1], "Tag")
	assert2.EqualRowValue(t, "<p>Some text</p>", gp.Objects[1], "Body")
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

	ctx := context.Background()
	n, err := hsp.ProcessNode(ctx, doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))

	assert2.EqualRowValue(t, "Header", gp.Objects[0], "Title")
	assert2.EqualRowValue(t, "h1", gp.Objects[0], "Tag")
	assert2.EqualRowValue(t, "<p>Some text</p>\n\t\t<p>Some text2</p>\n\t\t<p>Some text3</p>", gp.Objects[0], "Body")

	assert2.EqualRowValue(t, "Subheader", gp.Objects[1], "Title")
	assert2.EqualRowValue(t, "h2", gp.Objects[1], "Tag")
	assert2.EqualRowValue(t, "<p>Some text</p>", gp.Objects[1], "Body")

	gp = NewTestProcessor()
	hsp = NewHTMLHeadingSplitParser(gp, []string{"p"})

	n, err = hsp.ProcessNode(ctx, doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))

	assert2.EqualRowValue(t, "Header", gp.Objects[0], "Title")
	assert2.EqualRowValue(t, "h1", gp.Objects[0], "Tag")
	assert2.EqualRowValue(t, "Some text\n\t\tSome text2\n\t\tSome text3", gp.Objects[0], "Body")

	assert2.EqualRowValue(t, "Subheader", gp.Objects[1], "Title")
	assert2.EqualRowValue(t, "h2", gp.Objects[1], "Tag")
	assert2.EqualRowValue(t, "Some text", gp.Objects[1], "Body")
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

	ctx := context.Background()

	n, err := hsp.ProcessNode(ctx, doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))

	assert2.EqualRowValue(t, "Header", gp.Objects[0], "Title")
	assert2.EqualRowValue(t, "h1", gp.Objects[0], "Tag")
	assert2.EqualRowValue(t, "Some text<b>Foobar</b> <strong>Test</strong>", gp.Objects[0], "Body")

	assert2.EqualRowValue(t, "Subheader <strong>Foobar</strong>Test", gp.Objects[1], "Title")
	assert2.EqualRowValue(t, "h2", gp.Objects[1], "Tag")
	assert2.EqualRowValue(t, "Some text", gp.Objects[1], "Body")
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

	ctx := context.Background()
	n, err := hsp.ProcessNode(ctx, doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))

	assert2.EqualRowValue(t, "Foobar", gp.Objects[0], "Title")
	assert2.EqualRowValue(t, "p", gp.Objects[0], "Tag")

	assert2.EqualRowValue(t, "Test", gp.Objects[1], "Title")
	assert2.EqualRowValue(t, "p", gp.Objects[1], "Tag")
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

	ctx := context.Background()
	n, err := hsp.ProcessNode(ctx, doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))

	assert2.EqualRowValue(t, "", gp.Objects[0], "Title")
	assert2.EqualRowValue(t, "p", gp.Objects[0], "Tag")
	assert2.EqualRowValue(t, "Foobar", gp.Objects[0], "Body")

	assert2.EqualRowValue(t, "", gp.Objects[1], "Title")
	assert2.EqualRowValue(t, "p", gp.Objects[1], "Tag")
	assert2.EqualRowValue(t, "Test", gp.Objects[1], "Body")

	gp = NewTestProcessor()
	hsp = NewHTMLSplitParser(gp, []string{}, []string{"p"}, false)

	n, err = hsp.ProcessNode(ctx, doc)
	require.NoError(t, err)
	assert.Nil(t, n)

	assert.Equal(t, 2, len(gp.Objects))

	assert2.EqualRowValue(t, "", gp.Objects[0], "Title")
	assert2.EqualRowValue(t, "p", gp.Objects[0], "Tag")
	assert2.EqualRowValue(t, "<p>Foobar</p>", gp.Objects[0], "Body")

	assert2.EqualRowValue(t, "", gp.Objects[1], "Title")
	assert2.EqualRowValue(t, "p", gp.Objects[1], "Tag")
	assert2.EqualRowValue(t, "<p>Test</p>", gp.Objects[1], "Body")
}
