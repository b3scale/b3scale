package bbb

import (
	"encoding/xml"
	"time"
)

// Timestamp is the milliseconds passed since beginning
// of the epoch.
type Timestamp time.Time

// UnmarshalXML decodes the timestamp from XML data
func (t *Timestamp) UnmarshalXML(
	d *xml.Decoder,
	start xml.StartElement,
) error {
	var value int64
	if err := d.DecodeElement(&value, &start); err != nil {
		return err
	}

	// Decode timestamp
	sec := int64(value / 1000)
	nsec := int64((value % 1000) * 1000000)
	*t = Timestamp(time.Unix(sec, nsec).UTC())

	return nil
}

// MarshalXML encodes the timestamp into XML data
func (t Timestamp) MarshalXML(
	e *xml.Encoder,
	start xml.StartElement,
) error {
	timestamp := int64(time.Time(t).UnixNano() / 1000000)
	return e.EncodeElement(timestamp, start)
}

// String of timestamp, use time.Time.String
func (t Timestamp) String() string {
	return time.Time(t).String()
}
