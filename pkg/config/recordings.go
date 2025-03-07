package config

import (
	"encoding/json"
	"fmt"
)

// RecordingVisibility is an enum represeting the visibility
// of the recording: Published, Unpublishe, Protected
type RecordingVisibility int

// The recording visibility affects the state of the recording
// as in 'published' / 'unpublished', the 'protection' and
// the 'gl-listed' meta-parameter ('public').
const (
	RecordingVisibilityUnpublished RecordingVisibility = iota
	RecordingVisibilityPublished
	RecordingVisibilityProtected
	RecordingVisibilityPublic
	RecordingVisibilityPublicProtected
)

var recordingVisiblityKeys = map[RecordingVisibility]string{
	RecordingVisibilityUnpublished:     "unpublished",
	RecordingVisibilityPublished:       "published",
	RecordingVisibilityProtected:       "protected",
	RecordingVisibilityPublic:          "public",
	RecordingVisibilityPublicProtected: "public_protected",
}

// String implements the stringer interface for recording
// visibilty.
func (v RecordingVisibility) String() string {
	return recordingVisiblityKeys[v]
}

// Parse resolves the recording visibility key into the enum value
func ParseRecordingVisibility(s string) (RecordingVisibility, error) {
	for value, key := range recordingVisiblityKeys {
		if s == key {
			return value, nil
		}
	}

	return 0, fmt.Errorf("unknown recording visibility: '%s'", s)
}

// MarshalJSON implements the Marshaler interface
// for serializing a recording visibility.
func (v RecordingVisibility) MarshalJSON() ([]byte, error) {
	repr, ok := recordingVisiblityKeys[v]
	if !ok {
		return nil, fmt.Errorf("unknown recording visibility: '%d'", v)
	}

	return json.Marshal(repr)
}

// UnmarshalJSON implements the Unmarshaler interface
// for deserializing a recording visibility.
func (v *RecordingVisibility) UnmarshalJSON(b []byte) error {
	var repr string
	err := json.Unmarshal(b, &repr)
	if err != nil {
		return err
	}

	val, err := ParseRecordingVisibility(repr)
	if err != nil {
		return err
	}

	*v = val

	return nil
}
