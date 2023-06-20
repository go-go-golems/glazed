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
	ProcessInputObject(ctx context.Context, obj map[string]interface{}) error
	OutputFormatter() formatters.OutputFormatter
}

type GlazeProcessor struct {
	of  formatters.OutputFormatter
	oms []middlewares.ObjectMiddleware
}

func (gp *GlazeProcessor) OutputFormatter() formatters.OutputFormatter {
	return gp.of
}

type GlazeProcessorOption func(*GlazeProcessor)

func WithAppendObjectMiddleware(om ...middlewares.ObjectMiddleware) GlazeProcessorOption {
	return func(gp *GlazeProcessor) {
		gp.oms = append(gp.oms, om...)
	}
}

func WithPrependObjectMiddleware(om ...middlewares.ObjectMiddleware) GlazeProcessorOption {
	return func(gp *GlazeProcessor) {
		gp.oms = append(om, gp.oms...)
	}
}

func WithOutputFormatter(of formatters.OutputFormatter) GlazeProcessorOption {
	return func(gp *GlazeProcessor) {
		gp.of = of
	}
}

func NewGlazeProcessor(of formatters.OutputFormatter, options ...GlazeProcessorOption) *GlazeProcessor {
	ret := &GlazeProcessor{
		of:  of,
		oms: []middlewares.ObjectMiddleware{},
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
func (gp *GlazeProcessor) ProcessInputObject(ctx context.Context, obj map[string]interface{}) error {
	currentObjects := []map[string]interface{}{obj}

	for _, om := range gp.oms {
		nextObjects := []map[string]interface{}{}
		for _, obj := range currentObjects {
			objs, err := om.Process(obj)
			if err != nil {
				return err
			}
			nextObjects = append(nextObjects, objs...)
		}
		currentObjects = nextObjects
	}

	for _, obj := range currentObjects {
		gp.of.AddRow(&types.SimpleRow{Hash: obj})
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

func (gp *SimpleGlazeProcessor) GetTable() *types.Table {
	gp.formatter.Table.Finalize()
	return gp.formatter.Table
}
