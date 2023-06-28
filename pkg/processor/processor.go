package processor

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/formatters/table"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// TODO(manuel, 2023-04-27) This is probably a good location for With* constructors for middlewares

type Processor interface {
	ProcessInputObject(ctx context.Context, obj types.Row) error
	OutputFormatter() formatters.OutputFormatter
	Processor() *middlewares.Processor
}

type GlazeProcessor struct {
	of        formatters.OutputFormatter
	processor *middlewares.Processor
}

func (gp *GlazeProcessor) Processor() *middlewares.Processor {
	return gp.processor
}

func (gp *GlazeProcessor) OutputFormatter() formatters.OutputFormatter {
	return gp.of
}

type GlazeProcessorOption func(*GlazeProcessor)

func WithOutputFormatter(of formatters.OutputFormatter) GlazeProcessorOption {
	return func(gp *GlazeProcessor) {
		gp.of = of
	}
}

func WithMiddlewareProcessor(processor *middlewares.Processor) GlazeProcessorOption {
	return func(gp *GlazeProcessor) {
		gp.processor = processor
	}
}

func NewGlazeProcessor(of formatters.OutputFormatter, options ...GlazeProcessorOption) *GlazeProcessor {
	ret := &GlazeProcessor{
		of:        of,
		processor: middlewares.NewProcessor(),
	}

	for _, option := range options {
		option(ret)
	}

	return ret
}

// TODO(2022-12-18, manuel) we should actually make it possible to order the columns
// https://github.com/go-go-golems/glazed/issues/56

// ProcessInputObject takes an input object and processes it through the object middleware
// chain.
//
// The final output is added to the output formatter as a single row.
func (gp *GlazeProcessor) ProcessInputObject(ctx context.Context, obj types.Row) error {
	err := gp.processor.AddRow(ctx, obj)
	if err != nil {
		return err
	}
	return nil
}

// SimpleGlazeProcessor only collects the output and returns it as a types.Table
type SimpleGlazeProcessor struct {
	*GlazeProcessor
	formatter *table.OutputFormatter
}

func NewSimpleGlazeProcessor(options ...GlazeProcessorOption) *SimpleGlazeProcessor {
	formatter := table.NewOutputFormatter("csv")
	return &SimpleGlazeProcessor{
		GlazeProcessor: NewGlazeProcessor(formatter, options...),
		formatter:      formatter,
	}
}

func (gp *SimpleGlazeProcessor) GetTable(ctx context.Context) (*types.Table, error) {
	err := gp.processor.FinalizeTable(ctx)
	if err != nil {
		return nil, err
	}
	return gp.processor.Table, nil
}
