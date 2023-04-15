package table

import (
	"github.com/jedib0t/go-pretty/table"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestStyleToYAMLFromYAML(t *testing.T) {
	yamlString := strings.Builder{}

	style := prettyStyleToStyle(&table.StyleDefault)

	assert.Equal(t, style.Name, "StyleDefault")

	err := styleToYAML(&yamlString, style)
	require.NoError(t, err)

	s := yamlString.String()

	defaultStyle := table.StyleDefault
	_ = defaultStyle
	styleFromYAML, err := styleFromYAML(strings.NewReader(s))
	require.NoError(t, err)

	assert.Equal(t, table.StyleDefault, *styleFromYAML)
}
