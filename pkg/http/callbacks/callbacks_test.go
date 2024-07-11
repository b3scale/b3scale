package callbacks

import (
	"testing"
)

func TestOnRecordingReadyValidate(t *testing.T) {
	cb := &OnRecordingReady{
		SignedParameters: "PARAMS",
	}
	if err := cb.Validate(); err != nil {
		t.Error(err)
	}

	cb = &OnRecordingReady{}
	if err := cb.Validate(); err == nil {
		t.Error("expected error")
	}
}

func TestOnRecordingReadyEncode(t *testing.T) {
	cb := &OnRecordingReady{
		SignedParameters: "PARAMS",
	}
	data := cb.Encode()
	if data != "signed_parameters=PARAMS" {
		t.Error("unexpected data:", data)
	}

	t.Log(data)
}
