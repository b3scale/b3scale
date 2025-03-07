package config

import (
	"encoding/json"
	"testing"
)

func TestRecordingVisibilityString(t *testing.T) {
	if RecordingVisibilityPublished.String() != "published" {
		t.Error("unexpected result:", RecordingVisibilityPublished.String())
	}

	var f RecordingVisibility = 12
	s := f.String()
	if s != "" {
		t.Error("unexpected string value:", s)
	}
}

func TestRecordingVisibilityParse(t *testing.T) {

	v, err := ParseRecordingVisibility("published")
	if err != nil {
		t.Fatal(err)
	}
	if v != RecordingVisibilityPublished {
		t.Error("unexpected visibility:", v)
	}

	_, err = ParseRecordingVisibility("unknown")
	if err == nil {
		t.Fatal("visibility 'unknown' should not parse")
	}

}

func TestRecordingVisibilityMarshalJSON(t *testing.T) {
	data := map[string]any{
		"visibility": RecordingVisibilityProtected,
	}
	buf, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	res := string(buf)
	ref := `{"visibility":"protected"}`
	if res != ref {
		t.Error("unexpected result:", res)
	}
	t.Log(res)
}

func TestRecordingVisibilityUnmarshalJSON(t *testing.T) {
	buf := `{"foo":"protected", "bar": "published"}`
	res := map[string]RecordingVisibility{}

	if err := json.Unmarshal([]byte(buf), &res); err != nil {
		t.Fatal(err)
	}
	t.Log(res)

	if res["foo"] != RecordingVisibilityProtected {
		t.Error("unexpected result", res)
	}

	if res["bar"] != RecordingVisibilityPublished {
		t.Error("unexpected result", res)
	}
}
