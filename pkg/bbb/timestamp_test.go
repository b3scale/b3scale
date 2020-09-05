package bbb

import (
	"encoding/xml"
	"testing"
	"time"
)

func TestTimestampUnmarshalXML(t *testing.T) {
	// Let's just use the create success response
	// and the createTime attribute.
	data := readTestResponse("createSuccess.xml")
	res, err := UnmarshalCreateResponse(data)
	if err != nil {
		t.Error(err)
	}

	cTime := time.Time(res.CreateTime)
	if cTime.Minute() != 3 {
		t.Error("Unexpected minute:", cTime.Minute())
	}
	if cTime.Hour() != 17 {
		t.Error("Unexpected hour:", cTime.Hour())
	}
	if cTime.Second() != 29 {
		t.Error("Unexpected second:", cTime.Second())
	}
	if cTime.Year() != 2018 {
		t.Error("Unexpected year:", cTime.Year())
	}
	if cTime.Month() != 7 {
		t.Error("Unexpected Jul:", cTime.Month())
	}
}

func TestTimestampMarshalXML(t *testing.T) {
	repr := struct {
		XMLName xml.Name  `xml:"Time"`
		T       Timestamp `xml:"Ts"`
	}{
		T: Timestamp(time.Date(2018, 7, 9, 17, 3, 29, 613000000, time.UTC)),
	}
	data, err := xml.Marshal(repr)
	if err != nil {
		t.Error(err)
	}
	if string(data) != "<Time><Ts>1531155809613</Ts></Time>" {
		t.Error("Unexpected:", string(data))
	}
}
