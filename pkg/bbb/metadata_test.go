package bbb

import (
	"encoding/xml"
	"testing"
)

func TestMetadataUnmarshalXML(t *testing.T) {
	type Nested struct {
		Meta Metadata `xml:"meta"`
	}
	type Foo struct {
		XMLName xml.Name `xml:"foo"`
		*Nested
	}
	data := []byte("<foo><meta><a>23</a><b>bar</b><c>true</c></meta></foo>")
	foo := &Foo{}
	err := xml.Unmarshal(data, foo)
	if err != nil {
		t.Error(err)
	}
	t.Log(foo.Nested.Meta)
}

func TestMetadataMarshalXML(t *testing.T) {
	type Foo struct {
		XMLName xml.Name `xml:"foo"`
		Meta    Metadata `xml:"metadata"`
	}

	foo := &Foo{
		Meta: Metadata{
			"key":  "value",
			"test": "42",
		},
	}
	data, err := xml.Marshal(foo)
	if err != nil {
		t.Error(err)
	}

	if len(data) != 63 {
		t.Error("Unexpected data:", string(data), len(data))
	}
}

func TestMetadataUpdate(t *testing.T) {
	m := Metadata{
		"Presenter": "Dough Doe",
		"category":  "FINANCE",
		"TERM":      "Fall1994",
	}
	u := Metadata{
		"category": "LEGAL",
		"TERM":     "",
	}
	m.Update(u)

	if m["Presenter"] != "Dough Doe" {
		t.Error("presenter should not have changed")
	}
	if m["category"] != "LEGAL" {
		t.Error("category should have changed, is:", m["category"])
	}
	if _, ok := m["TERM"]; ok {
		t.Error("TERM should have been unset")
	}

}
