package middlewares

import (
	"glazed/pkg/types"
)

type TableMiddleware interface {
	// Process transforms a full table
	Process(table *types.Table) (*types.Table, error)
}

type ObjectMiddleware interface {
	// Process transforms each individual object
	//
	// TODO(manuel, 2022-11-20) Make the Process monadic, to return one or more new objects
	// this way we can build filtering interfaces
	//
	// Although maybe this should just be the interface for a single object,
	// which in our standard case would be all the rows at once.
	// A single object JSON manipulation would be just a single "row"
	//
	// Furthermore, we could maybe use the `Row` type here instead of a map,
	// although I am not sure how much we would impact efficiency, and maybe this is
	// all premature.
	Process(object map[string]interface{}) (map[string]interface{}, error)
}
