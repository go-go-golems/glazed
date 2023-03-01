package cmds

import (
	"github.com/go-go-golems/glazed/pkg/formatters"
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

func NewGlazeProcessor(of formatters.OutputFormatter, oms []middlewares.ObjectMiddleware) *GlazeProcessor {
	ret := &GlazeProcessor{
		of:  of,
		oms: oms,
	}

	return ret
}

// TODO(2022-12-18, manuel) we should actually make it possible to order the columns
// https://github.com/go-go-golems/glazed/issues/56
func (gp *GlazeProcessor) ProcessInputObject(obj map[string]interface{}) error {
	for _, om := range gp.oms {
		obj2, err := om.Process(obj)
		if err != nil {
			return err
		}
		obj = obj2
	}

	gp.of.AddRow(&types.SimpleRow{Hash: obj})
	return nil
}
