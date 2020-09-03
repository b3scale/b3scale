package bbb

import (
	"encoding/xml"
	"testing"
)

func TestMetadataUnmarshalXML(t *testing.T) {

	type Foo struct {
		XMLName xml.Name `xml:"foo"`
		Meta    Metadata `xml:"meta"`
	}

	data := []byte("<foo><meta><a>23</a><b>bar</b><c>true</c></meta></foo>")

	foo := &Foo{}
	err := xml.Unmarshal(data, foo)
	if err != nil {
		t.Error(err)
	}

	t.Log(foo)
}
