package api

import (
	"encoding/json"
	"testing"
)

func TestSerializeOpenAPISpec(t *testing.T) {
	spec := NewAPISpec()
	result, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(result))
}
