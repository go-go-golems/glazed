package middlewares

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/types"
)

type Processor struct {
	TableMiddlewares  []TableMiddleware
	ObjectMiddlewares []ObjectMiddleware
	RowMiddlewares    []RowMiddleware

	Table *types.Table
}

type ProcessorOption func(*Processor)

func WithTableMiddleware(tm ...TableMiddleware) ProcessorOption {
	return func(p *Processor) {
		p.TableMiddlewares = append(p.TableMiddlewares, tm...)
	}
}

func WIthPrependTableMiddleware(tm ...TableMiddleware) ProcessorOption {
	return func(p *Processor) {
		p.TableMiddlewares = append(tm, p.TableMiddlewares...)
	}
}

func WithObjectMiddleware(om ...ObjectMiddleware) ProcessorOption {
	return func(p *Processor) {
		p.ObjectMiddlewares = append(p.ObjectMiddlewares, om...)
	}
}

func WithPrependObjectMiddleware(om ...ObjectMiddleware) ProcessorOption {
	return func(p *Processor) {
		p.ObjectMiddlewares = append(om, p.ObjectMiddlewares...)
	}
}

func WithRowMiddleware(rm ...RowMiddleware) ProcessorOption {
	return func(p *Processor) {
		p.RowMiddlewares = append(p.RowMiddlewares, rm...)
	}
}

func WithPrependRowMiddleware(rm ...RowMiddleware) ProcessorOption {
	return func(p *Processor) {
		p.RowMiddlewares = append(rm, p.RowMiddlewares...)
	}
}

func NewProcessor(options ...ProcessorOption) *Processor {
	ret := &Processor{
		Table: types.NewTable(),
	}

	for _, option := range options {
		option(ret)
	}

	return ret
}

func (p *Processor) GetTable() *types.Table {
	return p.Table
}

func (p *Processor) FinalizeTable(ctx context.Context) error {
	for _, tm := range p.TableMiddlewares {
		table, err := tm.Process(ctx, p.Table)
		if err != nil {
			return err
		}
		p.Table = table
	}
	return nil
}

// AddRow runs row through the chain of ObjectMiddlewares, then RowMiddlewares and
// adds the resulting rows to the table.
func (p *Processor) AddRow(ctx context.Context, row types.Row) error {
	rows := []types.Row{row}

	for _, ow := range p.ObjectMiddlewares {
		newRows := []types.Row{}
		for _, row_ := range rows {
			rows_, err := ow.Process(ctx, row_)
			if err != nil {
				return err
			}
			newRows = append(newRows, rows_...)
		}

		rows = newRows
	}

	for _, mw := range p.RowMiddlewares {
		newRows := []types.Row{}
		for _, row_ := range rows {
			rows_, err := mw.Process(ctx, row_)
			if err != nil {
				return err
			}
			newRows = append(newRows, rows_...)
		}

		rows = newRows
	}

	p.Table.AddRows(rows...)

	return nil
}

func (p *Processor) AddObjectMiddleware(mw ...ObjectMiddleware) {
	p.ObjectMiddlewares = append(p.ObjectMiddlewares, mw...)
}

func (p *Processor) AddObjectMiddlewareInFront(mw ...ObjectMiddleware) {
	p.ObjectMiddlewares = append(mw, p.ObjectMiddlewares...)
}

func (p *Processor) AddObjectMiddlewareAtIndex(i int, mw ...ObjectMiddleware) {
	p.ObjectMiddlewares = append(p.ObjectMiddlewares[:i], append(mw, p.ObjectMiddlewares[i:]...)...)
}

func (p *Processor) AddRowMiddleware(mw ...RowMiddleware) {
	p.RowMiddlewares = append(p.RowMiddlewares, mw...)
}

func (p *Processor) AddRowMiddlewareInFront(mw ...RowMiddleware) {
	p.RowMiddlewares = append(mw, p.RowMiddlewares...)
}

func (p *Processor) AddRowMiddlewareAtIndex(i int, mw ...RowMiddleware) {
	p.RowMiddlewares = append(p.RowMiddlewares[:i], append(mw, p.RowMiddlewares[i:]...)...)
}

func (p *Processor) AddTableMiddleware(mw ...TableMiddleware) {
	p.TableMiddlewares = append(p.TableMiddlewares, mw...)
}

func (p *Processor) AddTableMiddlewareInFront(mw ...TableMiddleware) {
	p.TableMiddlewares = append(mw, p.TableMiddlewares...)
}

func (p *Processor) AddTableMiddlewareAtIndex(i int, mw ...TableMiddleware) {
	p.TableMiddlewares = append(p.TableMiddlewares[:i], append(mw, p.TableMiddlewares[i:]...)...)
}
