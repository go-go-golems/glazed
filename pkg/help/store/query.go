package store

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
)

func (s *Store) Find(ctx context.Context, pred query.Predicate) ([]*model.Section, error) {
	c := &query.Compiler{}
	pred(c)
	sqlStr, args := c.SQL()
	rows, err := s.db.QueryContext(ctx, sqlStr, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*model.Section
	for rows.Next() {
		sec := &model.Section{}
		// Only scan the main fields for now
		if err := rows.Scan(
			&sec.ID, &sec.Slug, &sec.Title, &sec.Subtitle, &sec.Short, &sec.Content,
			&sec.SectionType, &sec.IsTopLevel, &sec.IsTemplate, &sec.ShowPerDefault, &sec.Order,
		); err != nil {
			return nil, err
		}
		result = append(result, sec)
	}
	return result, nil
}
