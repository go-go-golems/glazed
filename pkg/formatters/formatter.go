package formatters

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"io"
	"path/filepath"
	"strings"
)

// This part of the library contains helper functionality to do output formatting
// for data.
//
// We want to do the following:
//    - print a Table with a header
//    - print the Table as csv
//    - render raw data as json
//    - render data as sqlite (potentially into multiple tables)
//    - support multiple tables
//        - transform tree like structures into flattened tables
//    - make it easy for the user to add data
//    - make it easy for the user to specify filters and fields
//    - provide a middleware like structure to chain filters and transformers
//    - provide a way to add documentation to the output / data schema
//    - support go templating
//    - load formatting values from a config file
//    - streaming functionality (i.e., output as values come in)
//
// Advanced functionality:
//    - excel output
//    - pager and search
//    - highlight certain values
//    - filter the input structure / output structure using a jq like query language

// The following is all geared towards tabulated output

type OutputFormatter interface {
	// RegisterTableMiddlewares allows individual OutputFormatters to register middlewares that might be
	// necessary for them to do the proper output. For example, table and excel output require
	// flattening the row objects before output.
	RegisterTableMiddlewares(mw *middlewares.TableProcessor) error
	RegisterRowMiddlewares(mw *middlewares.TableProcessor) error
	ContentType() string
	Close(ctx context.Context, w io.Writer) error
}

// TableOutputFormatter is an output formatter that requires an entire table to be computed up
// front before it can be output.
//
// NOTE(manuel, 2023-06-28) Since this is actually the first type of Formatter that was implemented,
// it is the current de facto standard. RowOutputFormatter has been added later and is thus not
// in wide spread use.
type TableOutputFormatter interface {
	OutputFormatter
	OutputTable(ctx context.Context, table *types.Table, w io.Writer) error
}

type RowOutputFormatter interface {
	OutputFormatter
	OutputRow(ctx context.Context, row types.Row, w io.Writer) error
}

func ComputeOutputFilename(
	outputFile string,
	outputFileTemplate string,
	row types.Row,
	index int,
) (string, error) {
	var outputFileName string
	if outputFileTemplate != "" {
		data := map[string]interface{}{}
		values := row

		for pair := values.Oldest(); pair != nil; pair = pair.Next() {
			k, v := pair.Key, pair.Value
			data[k] = v
		}
		data["rowIndex"] = index
		t, err := templating.RenderTemplateString(outputFileTemplate, data)
		if err != nil {
			return "", err
		}
		outputFileName = t
	} else {
		fileType := filepath.Ext(outputFile)
		baseName := strings.TrimSuffix(outputFile, fileType)

		outputFileName = fmt.Sprintf("%s-%d%s", baseName, index, fileType)

	}
	return outputFileName, nil
}
