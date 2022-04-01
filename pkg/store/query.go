package store

import (
	"regexp"

	sq "github.com/Masterminds/squirrel"
)

var (
	// ReMatchParamSQLUnsafe will match anything
	// NOT a-Z, 0-9, '_' and '-'.
	ReMatchParamSQLUnsafe = regexp.MustCompile(`[^a-zA-Z0-9_-]`)
)

// SQLSafeParam will remove any unsafe chars from
// a param potentially used without parameter binding.
func SQLSafeParam(p string) string {
	return ReMatchParamSQLUnsafe.ReplaceAllString(p, "")
}

// NewQuery creates a new query
func NewQuery() sq.SelectBuilder {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return sq.SelectBuilder(psql)
}

// NewDelete creates a new deletion query
func NewDelete() sq.DeleteBuilder {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return sq.DeleteBuilder(psql)
}

// Q is an alias for NewQuery
func Q() sq.SelectBuilder {
	return NewQuery()
}
