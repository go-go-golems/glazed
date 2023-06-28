package processor

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/formatters/table"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

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

// ProcessInputObject takes an input object and processes it through the object middleware
// chain.
func (gp *GlazeProcessor) ProcessInputObject(ctx context.Context, obj types.Row) error {
	err := gp.processor.AddRow(ctx, obj)
	if err != nil {
		return err
	}
	return nil
}

func NewSimpleGlazeProcessor(options ...GlazeProcessorOption) *GlazeProcessor {
	formatter := table.NewOutputFormatter("csv")
	return NewGlazeProcessor(formatter, options...)
}

func (gp *GlazeProcessor) GetTable(ctx context.Context) (*types.Table, error) {
	err := gp.processor.FinalizeTable(ctx)
	if err != nil {
		return nil, err
	}
	return gp.processor.Table, nil
}
