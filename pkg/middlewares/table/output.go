package table

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/types"
	"io"
)

type OutputMiddleware struct {
	formatter formatters.TableOutputFormatter
	writer    io.Writer
}

func (o *OutputMiddleware) Close(ctx context.Context) error {
	return o.formatter.Close(ctx)
}

func NewOutputMiddleware(formatter formatters.TableOutputFormatter, writer io.Writer) *OutputMiddleware {
	return &OutputMiddleware{
		formatter: formatter,
		writer:    writer,
	}
}

func (o *OutputMiddleware) Process(ctx context.Context, table *types.Table) (*types.Table, error) {
	err := o.formatter.OutputTable(ctx, table, o.writer)
	if err != nil {
		return nil, err
	}

	return table, nil
}
