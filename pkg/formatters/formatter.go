package formatters

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
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
	// TODO(manuel, 2022-11-12) We need to be able to output to a directory / to a stream / to multiple files
	AddRow(row types.Row)

	SetColumnOrder(columnOrder []types.FieldName)

	// AddTableMiddleware adds a middleware at the end of the processing list
	AddTableMiddleware(m middlewares.TableMiddleware)
	AddTableMiddlewareInFront(m middlewares.TableMiddleware)
	AddTableMiddlewareAtIndex(i int, m middlewares.TableMiddleware)

	GetTable() (*types.Table, error)

	Output(ctx context.Context, w io.Writer) error

	ContentType() string
}

func ComputeOutputFilename(outputFile string, outputFileTemplate string, row types.Row, index int) (string, error) {
	var outputFileName string
	if outputFileTemplate != "" {
		data := map[string]interface{}{}
		values := row.GetValues()

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

// StartFormatIntoChannel outputs the data from an OutputFormatter into a channel.
// This is useful to render a table into a stream, for example when rendering larger outputs
// into HTML when serving.
func StartFormatIntoChannel[T interface{ ~string }](
	ctx context.Context,
	formatter OutputFormatter,
) <-chan T {
	reader, writer := io.Pipe()
	c := make(chan T)

	eg, ctx2 := errgroup.WithContext(ctx)

	eg.Go(func() error {
		defer close(c)

		// read 8k chunks from reader
		buf := make([]byte, 8192)
		for {
			select {
			case <-ctx2.Done():
				return nil
			default:
				n, err := reader.Read(buf)
				if err != nil {
					return err
				}

				c <- T(buf[:n])
			}
		}
	})

	eg.Go(func() error {
		err := formatter.Output(ctx2, writer)
		defer writer.Close()
		if err != nil {
			writer.CloseWithError(err)
			return err
		}
		return nil
	})

	go func() {
		err := eg.Wait()
		if err != nil {
			if err != io.EOF {
				log.Error().Err(err).Msg("error in stream formatter")
			}
		}
	}()

	return c
}
