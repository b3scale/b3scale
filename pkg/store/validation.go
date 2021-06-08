package store

import (
	"fmt"
	"strings"
)

// A ValidationError is a mapping between the field
// name (same as json field name), and a list of error strings.
type ValidationError map[string][]string

// Error implements the error interface
func (e ValidationError) Error() string {
	fieldErrs := []string{}
	for field, err := range e {
		fieldErrs = append(fieldErrs, fmt.Sprintf("%s%v", field, err))
	}
	errs := strings.Join(fieldErrs, ", ")
	return "validation faild for fields: " + errs
}

// Add a validation error to the collection
func (e ValidationError) Add(field, err string) {
	if _, ok := e[field]; !ok {
		e[field] = []string{err}
	} else {
		e[field] = append(e[field], err)
	}
}
