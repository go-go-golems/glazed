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

func NewOutputChannelMiddleware[T interface{ ~string }](
	formatter formatters.RowOutputFormatter,
	c chan<- T,
) *OutputChannelMiddleware[T] {
	return &OutputChannelMiddleware[T]{
		formatter: formatter,
		c:         c,
	}
}

func (o OutputChannelMiddleware[T]) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
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

// ColumnsChannelMiddleware sends the column names of each row it receives to a channel.
// The column names are sent in a separate goroutine so as not to block the pipeline.
type ColumnsChannelMiddleware struct {
	c           chan<- []types.FieldName
	seenColumns map[types.FieldName]interface{}
}

func NewColumnsChannelMiddleware(c chan<- []types.FieldName) *ColumnsChannelMiddleware {
	return &ColumnsChannelMiddleware{
		c:           c,
		seenColumns: map[types.FieldName]interface{}{},
	}
}

func (c ColumnsChannelMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	fields := types.GetFields(row)
	newFields := []types.FieldName{}
	for _, field := range fields {
		if _, ok := c.seenColumns[field]; !ok {
			c.seenColumns[field] = nil
			newFields = append(newFields, field)
		}
	}

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
