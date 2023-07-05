package table

import (
	"bytes"
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

type OutputChannelMiddleware[T interface{ ~string }] struct {
	formatter formatters.RowOutputFormatter
	c         chan<- T
}

func (o *OutputChannelMiddleware[T]) Close(ctx context.Context) error {
	return o.formatter.Close(ctx)
}

func NewOutputChannelMiddleware[T interface{ ~string }](formatter formatters.RowOutputFormatter,
	c chan<- T) *OutputChannelMiddleware[T] {
	return &OutputChannelMiddleware[T]{
		formatter: formatter,
		c:         c,
	}
}

func (o *OutputChannelMiddleware[T]) Process(ctx context.Context, table *types.Table) (*types.Table, error) {
	for _, row_ := range table.Rows {
		var buf bytes.Buffer

		err := o.formatter.OutputRow(ctx, row_, &buf)
		if err != nil {
			return nil, err
		}

		o.c <- T(buf.String())
	}

	return table, nil
}

type ColumnsChannelMiddleware struct {
	c chan<- []types.FieldName
}

func NewColumnsChannelMiddleware(c chan<- []types.FieldName) *ColumnsChannelMiddleware {
	return &ColumnsChannelMiddleware{
		c: c,
	}
}

func (c *ColumnsChannelMiddleware) Close(ctx context.Context) error {
	return nil
}

func (c *ColumnsChannelMiddleware) Process(ctx context.Context, table *types.Table) (*types.Table, error) {
	c.c <- table.Columns
	return table, nil
}
