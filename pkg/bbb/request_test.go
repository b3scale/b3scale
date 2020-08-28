package bbb

import (
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

func TestParamsEncode(t *testing.T) {
	tests := map[string]Params{
		// Test parameter ordering
		"a=23&b=true&c=foo": Params{
			"c": "foo",
			"a": 23,
			"b": true,
		},

		// URL-safe encoding
		"name=Meeting+Name": Params{
			"name": "Meeting Name",
		},
	}

	for expected, params := range tests {
		result := params.Encode()
		if result != expected {
			t.Error("Unexpected result:", result)
		}
	}
}

func TestSign(t *testing.T) {
	// We use the example from the api documentation.
	// However as we encode our parameters with a deterministic
	// order, different to the example, we will end up with
	// a different checksum.
	req := &Request{
		Backend: &config.Backend{
			Secret: "639259d4-9dd8-4b25-bf01-95f9567eaf4b",
		},
		Resource: "create",
		Params: Params{
			"name":        "Test Meeting",
			"meetingID":   "abc123",
			"attendeePW":  "111222",
			"moderatorPW": "333444",
		},
	}

	checksum := req.Sign()
	if checksum != "0b89c2ebcfefb76772cbcf19386c33561f66f6ae" {
		t.Error("Unexpected checksum:", checksum)
	}
}

func TestValidate(t *testing.T) {
	// We use the example from the api documentation, now
	// for validating against a frontend secret.
	// order, different to the example, we will end up with
	// a different checksum.
	req := &Request{
		Frontend: &config.Frontend{
			Secret: "639259d4-9dd8-4b25-bf01-95f9567eaf4b",
		},
		Resource: "create",
		Params: Params{
			"name":        "Test Meeting",
			"meetingID":   "abc123",
			"attendeePW":  "111222",
			"moderatorPW": "333444",
		},
		Checksum: []byte("0b89c2ebcfefb76772cbcf19386c33561f66f6ae"),
	}

	// Success
	err := req.Validate()
	if err != nil {
		t.Error(err)
	}

	// Error
	req.Checksum = []byte("foob4r")
	err = req.Validate()
	if err == nil {
		t.Error("Expected a checksum error.")
	}
}
