package row

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/types"
	"gopkg.in/yaml.v3"
	"regexp"
)

type RenameColumnMiddleware struct {
	Renames map[types.FieldName]types.FieldName
	// orderedmap *regexp.Regexp -> string
	RegexpRenames RegexpReplacements

	// renamedColumns keeps tracks of columns that are being renamed. To avoid
	// going through all the Renames and RegexpRenames on every row,
	// we cache affected columns in renamedColumns.
	renamedColumns map[types.FieldName]types.FieldName
}

func NewFieldRenameColumnMiddleware(renames map[types.FieldName]types.FieldName) *RenameColumnMiddleware {
	return &RenameColumnMiddleware{
		Renames:        renames,
		RegexpRenames:  RegexpReplacements{},
		renamedColumns: map[types.FieldName]types.FieldName{},
	}
}

func NewRegexpRenameColumnMiddleware(renames RegexpReplacements) *RenameColumnMiddleware {
	return &RenameColumnMiddleware{
		Renames:        map[types.FieldName]types.FieldName{},
		RegexpRenames:  renames,
		renamedColumns: map[types.FieldName]types.FieldName{},
	}
}

func NewRenameColumnMiddleware(
	renames map[types.FieldName]types.FieldName,
	regexpRenames RegexpReplacements,
) *RenameColumnMiddleware {
	return &RenameColumnMiddleware{
		Renames:        renames,
		RegexpRenames:  regexpRenames,
		renamedColumns: map[types.FieldName]types.FieldName{},
	}
}

type RegexpReplacement struct {
	Regexp      *regexp.Regexp
	Replacement string
}

type RegexpReplacements []*RegexpReplacement

func (rr *RegexpReplacements) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("expected a mapping node, got %v", value.Kind)
	}
	*rr = RegexpReplacements{}
	for i := 0; i < len(value.Content); i += 2 {
		key := value.Content[i]
		val := value.Content[i+1]

		if key.Kind != yaml.ScalarNode {
			return fmt.Errorf("expected a scalar node, got %v", key.Kind)
		}
		if val.Kind != yaml.ScalarNode {
			return fmt.Errorf("expected a scalar node, got %v", val.Kind)
		}
		re, err := regexp.Compile(key.Value)
		if err != nil {
			return err
		}
		*rr = append(*rr, &RegexpReplacement{
			Regexp:      re,
			Replacement: val.Value,
		})
	}

	return nil
}

type ColumnMiddlewareConfig struct {
	FieldRenames map[types.FieldName]types.FieldName `yaml:"renames"`
	// FIXME regex renames actually need to ordered
	RegexpRenames RegexpReplacements `yaml:"regexpRenames"`
}

func NewRenameColumnMiddlewareFromYAML(decoder *yaml.Decoder) (*RenameColumnMiddleware, error) {
	var config ColumnMiddlewareConfig
	err := decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return NewRenameColumnMiddleware(config.FieldRenames, config.RegexpRenames), nil
}

// renameColumn takes a single column fields, looks up if it has already been processed previously,
// and otherwise runs it through the renames and regexp renames to compute the renamed column name.
func (r *RenameColumnMiddleware) renameColumn(
	column types.FieldName,
) types.FieldName {
	if rename, ok := r.renamedColumns[column]; ok {
		return rename
	}

	// we run string renames first, as we consider them more exhaustive matches
	for match, rename := range r.Renames {
		if column == match {
			r.renamedColumns[column] = rename
			return rename
		}
	}

	for _, rr := range r.RegexpRenames {
		rename := rr.Regexp.ReplaceAllString(column, rr.Replacement)
		if rename != column {
			r.renamedColumns[column] = rename
			return rename
		}
	}

	r.renamedColumns[column] = column
	return column
}

func (r *RenameColumnMiddleware) Process(ctx context.Context, row types.Row) ([]types.Row, error) {
	newRow := types.NewMapRow()
	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		newRow.Set(r.renameColumn(pair.Key), pair.Value)
	}

	ret := []types.Row{newRow}
	return ret, nil
}
