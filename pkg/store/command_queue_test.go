package store

import (
	"context"
	"fmt"
	"testing"
)

func TestSafeExecHandler(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx) //nolint

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

	result, err := safeExecHandler(ctx, cmd, h1)
	if result != nil {
		t.Error("unexpected result")
	}
	if err == nil {
		t.Error("error should not be nil")
	}
	t.Log(err)

	result, err = safeExecHandler(ctx, cmd, h2)
	if result == nil {
		t.Error("unexpected result")
	}
	if err == nil {
		t.Error("error should not be nil")
	}

	result, err = safeExecHandler(ctx, cmd, h3)
	if result != true {
		t.Error("unexpected result")
	}
	if err != nil {
		t.Error("error should be nil")
	}
}

func TestCountCommandsRequested(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx) //nolint

	_, err := tx.Exec(ctx, "DELETE FROM commands")
	if err != nil {
		t.Fatal(err)
	}

	count, err := CountCommandsRequested(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Error("did not expect anything in the queue")
	}

	// Queue some commands
	cmd := &Command{}
	if err := QueueCommand(ctx, tx, cmd); err != nil {
		t.Fatal(err)
	}
	if err := QueueCommand(ctx, tx, cmd); err != nil {
		t.Fatal(err)
	}

	count, err = CountCommandsRequested(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Error("did not expect len(q)=", count)
	}

}
