package formatters

import (
	"bytes"
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

// TableOutputFormatter is an output formatter that requires an entire table to be computed up
// front before it can be output.
//
// NOTE(manuel, 2023-06-28) Since this is actually the first type of Formatter that was implemented,
// it is the current de facto standard. RowOutputFormatter has been added later and is thus not
// in wide spread use.
type TableOutputFormatter interface {
	// RegisterMiddlewares allows individual OutputFormatters to register middlewares that might be
	// necessary for them to do the proper output. For example, table and excel output require
	// flattening the row objects before output.
	//
	// TODO(manuel, 2023-06-28) We could add the indication if this output formatter can stream here
	// Not all output formatters should have to take a full table, but could instead output a single row
	RegisterMiddlewares(mw *middlewares.Processor) error

	Output(ctx context.Context, table *types.Table, w io.Writer) error
	ContentType() string
}

type RowOutputFormatter interface {
	RegisterMiddlewares(mw *middlewares.Processor) error
	Output(ctx context.Context, row types.Row, w io.Writer) error
	ContentType() string
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

// StartFormatIntoChannel outputs the data from an TableOutputFormatter into a channel.
// This is useful to render a table into a stream, for example when rendering larger outputs
// into HTML when serving.
func StartFormatIntoChannel[T interface{ ~string }](
	ctx context.Context,
	table *types.Table,
	formatter TableOutputFormatter,
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
		err := formatter.Output(ctx2, table, writer)
		defer func(writer *io.PipeWriter) {
			_ = writer.Close()
		}(writer)
		if err != nil {
			_ = writer.CloseWithError(err)
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

func StreamRows(
	ctx context.Context,
	formatter RowOutputFormatter,
	inputChannel <-chan types.Row,
	outputChannel <-chan string,
) error {
	for {
		select {
		case <-ctx.Done():
			return nil

		case row, ok := <-inputChannel:
			if !ok {
				return nil
			}

			var buf bytes.Buffer
			err := formatter.Output(ctx, row, &buf)
			if err != nil {
				return err
			}
		}
	}
}

// StartStreamRows starts a background goroutine that returns two channels. The first
// channel is used to send rows to the background goroutine, the second channel is used
// to receive the formatted output.
//
// This background goroutine will close the output channel when it is done.
func StartStreamRows[T interface{ ~string }](
	ctx context.Context,
	formatter RowOutputFormatter,
) (chan<- types.Row, <-chan T) {
	// the inputChannel is buffered, so that a slow output stream still allows
	// the producer to move forward, to avoid having both latencies queue
	inputChannel := make(chan types.Row, 100)
	outputChannel := make(chan T)

	eg, ctx2 := errgroup.WithContext(ctx)

	eg.Go(func() error {
		defer close(outputChannel)

		for {
			select {
			case <-ctx2.Done():
				return nil

			case row, ok := <-inputChannel:
				if !ok {
					return nil
				}
				var buf bytes.Buffer
				err := formatter.Output(ctx2, row, &buf)
				if err != nil {
					return err
				}

				outputChannel <- T(buf.String())
			}
		}
	})

	return inputChannel, outputChannel
}
