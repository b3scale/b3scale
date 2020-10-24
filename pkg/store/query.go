package store

import (
	"fmt"
	"strings"
)

// Filter provides query options for states
type Filter struct {
	q     *Query
	idx   int
	attr  string
	op    string
	param interface{}
}

// Make string from filter
func (f *Filter) String() string {
	return fmt.Sprintf(
		"%s %s $%d",
		f.attr, f.op, f.idx+f.q.offset+1)
}

// Query is a representation of the WHERE clause
type Query struct {
	offset  int
	concat  string
	filters []*Filter
	joins   []string
}

// NewQuery creates a new query
func NewQuery() *Query {
	return &Query{
		offset:  0,
		concat:  " AND ",
		filters: []*Filter{},
		joins:   []string{},
	}
}

// Q is an alias for NewQuery
func Q() *Query {
	return NewQuery()
}

// Offset updates the offset of the query
func (q *Query) Offset(o int) *Query {
	q.offset = o
	return q
}

// Join includeds a join in the query
func (q *Query) Join(table, on string) *Query {
	join := fmt.Sprintf("JOIN %s ON %s", table, on)
	q.joins = append(q.joins, join)
	return q
}

// Filter adds a test to the query
func (q *Query) Filter(attr, op string, param interface{}) *Query {
	idx := len(q.filters)
	f := &Filter{
		q:     q,
		idx:   idx,
		attr:  attr,
		op:    op,
		param: param,
	}
	q.filters = append(q.filters, f)
	return q
}

// Eq adds an equality test to the query
func (q *Query) Eq(attr string, param interface{}) *Query {
	return q.Filter(attr, "=", param)
}

// Internal where returns the WHERE query string
func (q *Query) where() string {
	where := make([]string, 0, len(q.filters))
	for _, f := range q.filters {
		where = append(where, f.String())
	}

	return strings.Join(where, q.concat)
}

// Internal params, gets a list of all params
func (q *Query) params() []interface{} {
	params := make([]interface{}, 0, len(q.filters))
	for _, f := range q.filters {
		params = append(params, f.param)
	}
	return params
}

// Internal fetch from and related
func (q *Query) related() string {
	return strings.Join(q.joins, " ")
}
