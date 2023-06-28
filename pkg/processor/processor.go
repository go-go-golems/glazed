package processor

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/formatters/table"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type TableProcessor interface {
	AddRow(ctx context.Context, obj types.Row) error
	OutputFormatter() formatters.TableOutputFormatter
	Finalize(ctx context.Context) error
	GetTable() *types.Table
}

// GlazeProcessor is a simple wrapper around a middlewares.Processor that also handles
// an OutputFormatter.
//
// NOTE(manuel, 2023-06-28) At the end of tackling the big refactor of both ordered map rows
// and a new row middleware concept, and introducing the dedicated middlewares.Processor,
// this interface seems a bit unnecessary / confusing.
//
// https://github.com/go-go-golems/glazed/issues/310
//
// Because I am planning to add streaming output so that we can cut down on memory usage in sqleton,
// I might revisit this and find a better way to:
// - connect a middleares processor to an output formatter
//   - for example, an output formatter could actually be a final middleware step
//   - streaming output formatters could be registered as a row middleware (and we could have multiple
//     streaming row level output formatters for debugging and monitoring purposes)
//   - normal output formatters could be registered as a table middleware that gets called once the entire
//     input has been processed.
//
// Approaching everything as middlewares would allow us to work long running glazed commands,
// and if no TableMiddleware is registered, we can discard values altogether after they've been run
// through the row middlewares.
type GlazeProcessor struct {
	*middlewares.Processor
	of formatters.TableOutputFormatter
}

func (gp *GlazeProcessor) OutputFormatter() formatters.TableOutputFormatter {
	return gp.of
}

type GlazeProcessorOption func(*GlazeProcessor)

func NewGlazeProcessor(of formatters.TableOutputFormatter, options ...GlazeProcessorOption) (*GlazeProcessor, error) {
	ret := &GlazeProcessor{
		of:        of,
		Processor: middlewares.NewProcessor(),
	}

	for _, option := range options {
		option(ret)
	}

	err := of.RegisterMiddlewares(ret.Processor)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func NewSimpleGlazeProcessor(options ...GlazeProcessorOption) (*GlazeProcessor, error) {
	formatter := table.NewOutputFormatter("csv")
	return NewGlazeProcessor(formatter, options...)
}
