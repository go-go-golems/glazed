package middlewares

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/types"
)

type Processor interface {
	AddRow(ctx context.Context, obj types.Row) error
}

type TableProcessor struct {
	TableMiddlewares  []TableMiddleware
	ObjectMiddlewares []ObjectMiddleware
	RowMiddlewares    []RowMiddleware

	Table *types.Table
}

type TableProcessorOption func(*TableProcessor)

func WithTableMiddleware(tm ...TableMiddleware) TableProcessorOption {
	return func(p *TableProcessor) {
		p.TableMiddlewares = append(p.TableMiddlewares, tm...)
	}
}

func WIthPrependTableMiddleware(tm ...TableMiddleware) TableProcessorOption {
	return func(p *TableProcessor) {
		p.TableMiddlewares = append(tm, p.TableMiddlewares...)
	}
}

func WithObjectMiddleware(om ...ObjectMiddleware) TableProcessorOption {
	return func(p *TableProcessor) {
		p.ObjectMiddlewares = append(p.ObjectMiddlewares, om...)
	}
}

func WithPrependObjectMiddleware(om ...ObjectMiddleware) TableProcessorOption {
	return func(p *TableProcessor) {
		p.ObjectMiddlewares = append(om, p.ObjectMiddlewares...)
	}
}

func WithRowMiddleware(rm ...RowMiddleware) TableProcessorOption {
	return func(p *TableProcessor) {
		p.RowMiddlewares = append(p.RowMiddlewares, rm...)
	}
}

func WithPrependRowMiddleware(rm ...RowMiddleware) TableProcessorOption {
	return func(p *TableProcessor) {
		p.RowMiddlewares = append(rm, p.RowMiddlewares...)
	}
}

func NewTableProcessor(options ...TableProcessorOption) *TableProcessor {
	ret := &TableProcessor{
		Table: types.NewTable(),
	}

	for _, option := range options {
		option(ret)
	}

	return ret
}

func (p *TableProcessor) GetTable() *types.Table {
	return p.Table
}

// NOTE(manuel, 2023-06-30) should this maybe rather be called close?
func (p *TableProcessor) RunTableMiddlewares(ctx context.Context) error {
	// close the object and row middlewares

	for _, tm := range p.TableMiddlewares {
		table, err := tm.Process(ctx, p.Table)
		if err != nil {
			return err
		}
		p.Table = table

		defer func(tm TableMiddleware, ctx context.Context) {
			_ = tm.Close(ctx)
		}(tm, ctx)
	}

	return nil
}

// AddRow runs row through the chain of ObjectMiddlewares, then RowMiddlewares and
// adds the resulting rows to the table.
func (p *TableProcessor) AddRow(ctx context.Context, row types.Row) error {
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

func (p *TableProcessor) AddObjectMiddleware(mw ...ObjectMiddleware) {
	p.ObjectMiddlewares = append(p.ObjectMiddlewares, mw...)
}

func (p *TableProcessor) AddObjectMiddlewareInFront(mw ...ObjectMiddleware) {
	p.ObjectMiddlewares = append(mw, p.ObjectMiddlewares...)
}

func (p *TableProcessor) AddObjectMiddlewareAtIndex(i int, mw ...ObjectMiddleware) {
	p.ObjectMiddlewares = append(p.ObjectMiddlewares[:i], append(mw, p.ObjectMiddlewares[i:]...)...)
}

func (p *TableProcessor) AddRowMiddleware(mw ...RowMiddleware) {
	p.RowMiddlewares = append(p.RowMiddlewares, mw...)
}

func (p *TableProcessor) AddRowMiddlewareInFront(mw ...RowMiddleware) {
	p.RowMiddlewares = append(mw, p.RowMiddlewares...)
}

func (p *TableProcessor) AddRowMiddlewareAtIndex(i int, mw ...RowMiddleware) {
	p.RowMiddlewares = append(p.RowMiddlewares[:i], append(mw, p.RowMiddlewares[i:]...)...)
}

func (p *TableProcessor) AddTableMiddleware(mw ...TableMiddleware) {
	p.TableMiddlewares = append(p.TableMiddlewares, mw...)
}

func (p *TableProcessor) AddTableMiddlewareInFront(mw ...TableMiddleware) {
	p.TableMiddlewares = append(mw, p.TableMiddlewares...)
}

func (p *TableProcessor) AddTableMiddlewareAtIndex(i int, mw ...TableMiddleware) {
	p.TableMiddlewares = append(p.TableMiddlewares[:i], append(mw, p.TableMiddlewares[i:]...)...)
}

func (p *TableProcessor) ReplaceTableMiddleware(mw ...TableMiddleware) {
	p.TableMiddlewares = mw
}
