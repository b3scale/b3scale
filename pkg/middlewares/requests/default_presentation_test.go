package requests

import (
	"testing"
)

func TestMakePresenationRequestBody(t *testing.T) {
	body := makePresentationRequestBody("http://foo.jpg")
	t.Log(string(body))
}
