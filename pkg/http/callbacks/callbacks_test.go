package callbacks

import (
	"testing"
)

func TestCallbackValidate(t *testing.T) {
	cb := &Callback{
		SignedParameters: "PARAMS",
	}
	if err := cb.Validate(); err != nil {
		t.Error(err)
	}

	cb = &Callback{}
	if err := cb.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestCallbackEncode(t *testing.T) {
	cb := &Callback{
		SignedParameters: "PARAMS",
	}
	data := cb.Encode()
	if data != "signed_parameters=PARAMS" {
		t.Error("unexpected data:", data)
	}

	t.Log(data)
}
