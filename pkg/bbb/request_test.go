package bbb

import (
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

func TestParamsString(t *testing.T) {
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
		result := params.String()
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

func TestString(t *testing.T) {
	// Request create to backend
	req := &Request{
		Backend: &config.Backend{
			Host:   "https://bbbackend",
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

	// Call stringer
	reqURL := req.String()
	expected := "https://bbbackend/create" +
		"?attendeePW=111222&meetingID=abc123" +
		"&moderatorPW=333444&name=Test+Meeting&" +
		"checksum=0b89c2ebcfefb76772cbcf19386c33561f66f6ae"
	if reqURL != expected {
		t.Error("Unexpected request URL:", reqURL)
	}

	// No params
	req.Params = Params{}
	reqURL = req.String()
	expected = "https://bbbackend/create" +
		"?checksum=8a21c9b7e3b18541974c9e78c0d0bfa790c665eb"
	if reqURL != expected {
		t.Error("Unexpected request URL:", reqURL)
	}

}
