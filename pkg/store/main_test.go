package store

import (
	"context"
	"os"
	"testing"
)

// TestMain is the test entrypoint, ensuring that a migrated
// database is present.
func TestMain(m *testing.M) {
	// Setup pool connection and migrate database
	ctx := context.Background()
	if err := ConnectTest(ctx); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
