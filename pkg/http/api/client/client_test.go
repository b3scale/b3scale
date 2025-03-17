package client

import (
	"testing"

	"github.com/b3scale/b3scale/pkg/http/api"
)

func TestClientImplementsInterface(t *testing.T) {
	var c api.Client = New("foo", "bar").WithUserAgent("useragent/1.0")
	_ = c
}
