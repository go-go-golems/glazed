package middlewares

import (
	"github.com/go-go-golems/glazed/pkg/types"
)

type TableMiddleware interface {
	// Process transforms a full table
	Process(table *types.Table) (*types.Table, error)
}

type ObjectMiddleware interface {
	// Process transforms each individual object. Each object can return multiple
	// objects which will get processed individually downstream.
	Process(object map[string]interface{}) ([]map[string]interface{}, error)
}
