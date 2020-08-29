package bbb

import (
	"testing"
)

func TestTimestampUnmarshalXML(t *testing.T) {
	// Let's just use the create success response
	// and the createTime attribute.
	data := readTestResponse("createSuccess.xml")
	res, err := UnmarshalCreateResponse(data)
	if err != nil {
		t.Error(err)
	}

	t.Log(res.CreateTime)
}
