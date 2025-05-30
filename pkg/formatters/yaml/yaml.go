package yaml

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type OutputFormatter struct {
	OutputFile           string
	OutputFileTemplate   string
	OutputMultipleFiles  bool
	OutputIndividualRows bool
}

func (f *OutputFormatter) OutputRow(ctx context.Context, row types.Row, w io.Writer) error {
	m := types.RowToMap(row)

	// Convert *time.Time to time.Time for YAML encoding, otherwise the yaml lib crashes
	for k, v := range m {
		if timePtr, ok := v.(*time.Time); ok {
			if timePtr != nil {
				m[k] = *timePtr
			} else {
				delete(m, k)
			}
		}
	}

	encoder := yaml.NewEncoder(w)
	err := encoder.Encode(m)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "\n---\n\n")
	if err != nil {
		return err
	}

	return nil
}

var _ formatters.TableOutputFormatter = (*OutputFormatter)(nil)

func (f *OutputFormatter) Close(ctx context.Context, w io.Writer) error {
	return nil
}

func (f *OutputFormatter) RegisterTableMiddlewares(mw *middlewares.TableProcessor) error {
	return nil
}

func (f *OutputFormatter) RegisterRowMiddlewares(mw *middlewares.TableProcessor) error {
	return nil
}

func (f *OutputFormatter) OutputTable(ctx context.Context, table_ *types.Table, w io.Writer) error {
	if f.OutputMultipleFiles {
		if f.OutputFileTemplate == "" && f.OutputFile == "" {
			return errors.New("neither output file or output file template is set")
		}

		for i, row := range table_.Rows {
			outputFileName, err := formatters.ComputeOutputFilename(f.OutputFile, f.OutputFileTemplate, row, i)
			if err != nil {
				return err
			}

			f_, err := os.Create(outputFileName)
			if err != nil {
				return err
			}

			defer func(f_ *os.File) {
				_ = f_.Close()
			}(f_)

			encoder := yaml.NewEncoder(f_)
			err = encoder.Encode(row)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w, "Wrote output to %s\n", outputFileName)
		}

		return nil
	}

	if f.OutputIndividualRows {
		if len(table_.Rows) > 1 {
			return errors.New("output individual rows is set but there are multiple rows in the table")
		}

		if f.OutputFile != "" {
			f_, err := os.Create(f.OutputFile)
			if err != nil {
				return err
			}
			w = f_
			defer func(f_ *os.File) {
				_ = f_.Close()
			}(f_)

			if len(table_.Rows) == 0 {
				_, _ = fmt.Fprintln(w, "Empty table, an empty file was created")
				return nil
			}
		}

		encoder := yaml.NewEncoder(w)
		err := encoder.Encode(table_.Rows[0])
		if err != nil {
			return err
		}

		return nil
	} else {
		var rows []types.Row
		rows = append(rows, table_.Rows...)

		encoder := yaml.NewEncoder(w)
		err := encoder.Encode(rows)
		if err != nil {
			return err
		}

		return nil
	}
}

func (f *OutputFormatter) ContentType() string {
	return "application/yaml"
}

type OutputFormatterOption func(*OutputFormatter)

func WithYAMLOutputFile(outputFile string) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.OutputFile = outputFile
	}
}

func WithOutputFileTemplate(outputFileTemplate string) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.OutputFileTemplate = outputFileTemplate
	}
}

func WithOutputMultipleFiles(outputMultipleFiles bool) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.OutputMultipleFiles = outputMultipleFiles
	}
}

func WithOutputIndividualRows(outputIndividualRows bool) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.OutputIndividualRows = outputIndividualRows
	}
}

func NewOutputFormatter(opts ...OutputFormatterOption) *OutputFormatter {
	f := &OutputFormatter{}

	for _, opt := range opts {
		opt(f)
	}

	return f
}
