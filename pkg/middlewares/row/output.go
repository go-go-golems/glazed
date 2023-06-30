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
	err := o.formatter.Output(ctx, row, o.writer)
	if err != nil {
		return nil, err
	}

	return []types.Row{row}, nil
}

type OutputChannelMiddleware struct {
	formatter formatters.RowOutputFormatter
	c         chan<- string
}

func NewOutputChannelMiddleware(formatter formatters.RowOutputFormatter, c chan<- string) *OutputChannelMiddleware {
	return &OutputChannelMiddleware{
		formatter: formatter,
		c:         c,
	}
}

func (o OutputChannelMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	var buf bytes.Buffer
	err := o.formatter.Output(ctx, row, &buf)
	if err != nil {
		return nil, err
	}

	o.c <- buf.String()

	return []types.Row{row}, nil
}
