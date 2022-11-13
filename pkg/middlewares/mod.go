package middlewares

import (
	"dd-cli/pkg/types"
)

type TableMiddleware interface {
	// Process transforms a full table
	Process(table *types.Table) (*types.Table, error)
}

type ObjectMiddleware interface {
	// Process transforms each individual object
	Process(object interface{}) (interface{}, error)
}
