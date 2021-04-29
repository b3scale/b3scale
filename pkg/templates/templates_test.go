package templates

import (
	"bytes"
	"testing"
)

// Test rendering templates

func TestTmplRedirect(t *testing.T) {
	url := "http://foo-bar-test"
	res := Redirect(url)
	t.Log(string(res))

	if !bytes.Contains(res, []byte(url)) {
		t.Error("result should contain the URL")
	}
}

func TestTmplRenderConcurrent(t *testing.T) {
	for i := 0; i < 1000; i++ {
		go func() {
			res := Redirect("foooo")
			if res == nil {
				t.Error("unexepted result")
			}
		}()
	}
}

func TestTmplRetryJoin(t *testing.T) {
	url := "http://foo-bar-test"
	res := RetryJoin(url)
	t.Log(string(res))

	if !bytes.Contains(res, []byte(url)) {
		t.Error("result should contain the URL")
	}
}
