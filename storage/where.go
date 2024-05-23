package storage

import (
	"github.com/uptrace/bun"
)

type Where struct {
	Column   bun.Ident
	Operator bun.Safe
	Value    any
}

func addWhere(query *bun.SelectQuery, where []Where) *bun.SelectQuery {
	for _, w := range where {
		if w.Operator == bun.Safe("IN") {
			query = query.Where("? ? (?)", w.Column, w.Operator, bun.In(w.Value))
		} else {
			query = query.Where("? ? ?", w.Column, w.Operator, w.Value)
		}

	}
	return query
}
