package store

import (
	"testing"
)

func TestValidationErrorInterface(t *testing.T) {
	f := func() error {
		return &ValidationError{
			"field":  []string{"error1", "error2"},
			"field2": []string{"required"},
		}
	}
	err := f()
	t.Log(err.Error())
}
