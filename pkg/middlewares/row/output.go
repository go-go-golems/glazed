package row

import (
	"bytes"
	"context"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/types"
	"io"
)

type OutputMiddleware struct {
	formatter formatters.RowOutputFormatter
	writer    io.Writer
}

func (o OutputMiddleware) Close(ctx context.Context) error {
	return o.formatter.Close(ctx)
}

func NewOutputMiddleware(formatter formatters.RowOutputFormatter, writer io.Writer) *OutputMiddleware {
	return &OutputMiddleware{
		formatter: formatter,
		writer:    writer,
	}
}

func (o OutputMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	err := o.formatter.OutputRow(ctx, row, o.writer)
	if err != nil {
		return nil, err
	}

	return []types.Row{row}, nil
}

// OutputChannelMiddleware sends each row it receives to a channel after formatting it with the given formatter.
// This will block the pipeline until the channel is read from (or the buffer of the channel is full).
type OutputChannelMiddleware[T interface{ ~string }] struct {
	formatter formatters.RowOutputFormatter
	c         chan<- T
}

func (o *OutputChannelMiddleware[T]) Close(ctx context.Context) error {
	return o.formatter.Close(ctx)
}

func (o *OutputChannelMiddleware[T]) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	var buf bytes.Buffer
	err := o.formatter.OutputRow(ctx, row, &buf)
	if err != nil {
		return nil, err
	}

	// We don't run the sending in its own goroutine, because we want to rate limit on actual rows being processed.
	// This is different from the ColumnsChannelMiddleware, which is not going to process a lot of data and is
	// just sending column names.
	o.c <- T(buf.String())

	return []types.Row{row}, nil
}

func NewOutputChannelMiddleware[T interface{ ~string }](
	formatter formatters.RowOutputFormatter,
	c chan<- T,
) *OutputChannelMiddleware[T] {
	return &OutputChannelMiddleware[T]{
		formatter: formatter,
		c:         c,
	}
}

// ColumnsChannelMiddleware sends the column names of each row it receives to a channel.
// The column names are sent in a separate goroutine so as not to block the pipeline.
type ColumnsChannelMiddleware struct {
	c            chan<- []types.FieldName
	seenColumns  map[types.FieldName]interface{}
	onlyFirstRow bool
	seenFirstRow bool
}

func NewColumnsChannelMiddleware(c chan<- []types.FieldName, onlyFirstRow bool) *ColumnsChannelMiddleware {
	return &ColumnsChannelMiddleware{
		c:            c,
		onlyFirstRow: onlyFirstRow,
		seenColumns:  map[types.FieldName]interface{}{},
	}
}

func (c *ColumnsChannelMiddleware) Close(ctx context.Context) error {
	return nil
}

func (c *ColumnsChannelMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	if c.onlyFirstRow && c.seenFirstRow {
		return []types.Row{row}, nil
	}
	fields := types.GetFields(row)
	newFields := []types.FieldName{}
	for _, field := range fields {
		if _, ok := c.seenColumns[field]; !ok {
			c.seenColumns[field] = nil
			newFields = append(newFields, field)
		}
	}

	c.seenFirstRow = true

	if len(newFields) > 0 {
		// send the columns to the channel in a goroutine so that we don't block the pipeline
		go func() {
			select {
			case <-ctx.Done():
				return
			case c.c <- newFields:
			}
		}()
	}

	return []types.Row{row}, nil
}
