package callbacks

import (
	"testing"
)

func TestSignedBodyValidate(t *testing.T) {
	cb := &SignedBody{
		SignedParameters: "PARAMS",
	}
	if err := cb.Validate(); err != nil {
		t.Error(err)
	}

	cb = &SignedBody{}
	if err := cb.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestSignedBodyEncode(t *testing.T) {
	cb := &SignedBody{
		SignedParameters: "PARAMS",
	}
	data := cb.Encode()
	if data != "signed_parameters=PARAMS" {
		t.Error("unexpected data:", data)
	}

	t.Log(data)
}
