package cmds

import (
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/formatters/table"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type Processor interface {
	ProcessInputObject(obj map[string]interface{}) error
	OutputFormatter() formatters.OutputFormatter
}

type GlazeProcessor struct {
	of  formatters.OutputFormatter
	oms []middlewares.ObjectMiddleware
}

func (gp *GlazeProcessor) OutputFormatter() formatters.OutputFormatter {
	return gp.of
}

func NewGlazeProcessor(of formatters.OutputFormatter, oms ...middlewares.ObjectMiddleware) *GlazeProcessor {
	ret := &GlazeProcessor{
		of:  of,
		oms: oms,
	}

	return ret
}

// TODO(2022-12-18, manuel) we should actually make it possible to order the columns
// https://github.com/go-go-golems/glazed/issues/56

// ProcessInputObject takes an input object and processes it through the object middleware
// chain.
//
// The final output is added to the output formatter as a single row.
func (gp *GlazeProcessor) ProcessInputObject(obj map[string]interface{}) error {
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

func NewSimpleGlazeProcessor(oms ...middlewares.ObjectMiddleware) *SimpleGlazeProcessor {
	formatter := table.NewOutputFormatter("csv")
	return &SimpleGlazeProcessor{
		GlazeProcessor: NewGlazeProcessor(formatter, oms...),
		formatter:      formatter,
	}
}

func (gp *SimpleGlazeProcessor) GetTable() *types.Table {
	gp.formatter.Table.Finalize()
	return gp.formatter.Table
}
