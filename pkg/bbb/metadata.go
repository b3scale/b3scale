package bbb

import (
	"encoding/xml"
)

// Metadata about the BBB instance, this is not exactly
// specified in the docs, so we are using a map with
// string keys and an empty interface for the values.
type Metadata map[string]string

// UnmarshalXML decodes an unordered key, value mapping
// from XML.
func (meta *Metadata) UnmarshalXML(
	d *xml.Decoder, start xml.StartElement,
) error {
	if *meta == nil {
		*meta = make(Metadata)
	}

	var (
		key   string
		value string
	)

loop:
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch t.(type) {
		case xml.StartElement:
			elem := t.(xml.StartElement)
			key = string(elem.Name.Local)
			err = d.DecodeElement(&value, &elem)
			if err != nil {
				return err
			}
			(*meta)[key] = value

		case xml.EndElement:
			if t.(xml.EndElement) == start.End() {
				break loop // We reached our end
			}
			break
		}
	}
	return nil
}
