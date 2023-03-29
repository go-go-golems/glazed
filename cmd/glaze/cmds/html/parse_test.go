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
