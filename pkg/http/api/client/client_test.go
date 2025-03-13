package client

import (
	"testing"

	"github.com/b3scale/b3scale/pkg/http/api"
)

func TestClientImplementsInterface(t *testing.T) {
	var c api.Client = New("foo", "bar", "agent/1.0")
	_ = c
}
