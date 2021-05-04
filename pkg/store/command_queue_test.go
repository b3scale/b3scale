package store

import (
	"context"
	"fmt"
	"testing"
)

func TestSafeExecHandler(t *testing.T) {
	h1 := func(ctx context.Context, cmd *Command) (interface{}, error) {
		panic("panic with fatal error")
	}
	h2 := func(ctx context.Context, cmd *Command) (interface{}, error) {
		return false, fmt.Errorf("error ret")
	}
	h3 := func(ctx context.Context, cmd *Command) (interface{}, error) {
		return true, nil
	}

	cmd := &Command{}

	result, err := safeExecHandler(cmd, h1)
	if result != nil {
		t.Error("unexpected result")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	t.Log(err)

	result, err = safeExecHandler(cmd, h2)
	if result == nil {
		t.Error("unexpected result")
	}
	if err == nil {
		t.Error("error should not be nil")
	}

	result, err = safeExecHandler(cmd, h3)
	if result != true {
		t.Error("unexpected result")
	}
	if err != nil {
		t.Error("error should be nil")
	}
}
