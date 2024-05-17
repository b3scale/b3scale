package bbb

import "encoding/xml"

// Recordings have a metadata xml.

// RecordingMetadata can be parsed from a metadata.xml
type RecordingMetadata struct {
	XMLName      xml.Name                   `xml:"recording"`
	RecordID     string                     `xml:"id"`
	State        string                     `xml:"state"`
	Published    bool                       `xml:"published"`
	StartTime    Timestamp                  `xml:"start_time"`
	EndTime      Timestamp                  `xml:"end_time"`
	Participants int                        `xml:"participants"`
	Meeting      *RecordingMetadataMeeting  `xml:"meeting"`
	Meta         Metadata                   `xml:"meta"`
	Playback     *RecordingMetadataPlayback `xml:"playback"`
	RawSize      int                        `xml:"raw_size"`
}

// UnmarshalRecordingMetadata deserializes bytes
func UnmarshalRecordingMetadata(
	data []byte,
) (*RecordingMetadata, error) {
	res := &RecordingMetadata{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// ToRecording converts a recording metadata into a recording
func (m *RecordingMetadata) ToRecording() *Recording {
	var meetingID, internalMeetingID, name string
	var isBreakout bool

	if m.Meeting != nil {
		meetingID = m.Meeting.MeetingID
		internalMeetingID = m.Meeting.InternalMeetingID
		name = m.Meeting.Name
		isBreakout = m.Meeting.Breakout
	}

	r := &Recording{
		RecordID:          m.RecordID,
		MeetingID:         meetingID,
		InternalMeetingID: internalMeetingID,
		Name:              name,
		IsBreakout:        isBreakout,

		Published:    m.Published,
		State:        m.State,
		StartTime:    m.StartTime,
		EndTime:      m.EndTime,
		Participants: m.Participants,
		Metadata:     m.Meta,
		Formats: []*Format{
			{
				Type:           m.Playback.Format,
				URL:            m.Playback.Link,
				ProcessingTime: m.Playback.ProcessingTime,
				Length:         m.Playback.Duration / 1000 / 60,
			},
		},
	}
	return r
}

// RecordingMetadataMeeting encodes the meeting information
// from the recordings metadata.xml
type RecordingMetadataMeeting struct {
	InternalMeetingID string `xml:"id,attr"`
	MeetingID         string `xml:"externalId,attr"`
	Name              string `xml:"name,attr"`
	Breakout          bool   `xml:"breakout,attr"`
}

// RecordingMetadataPlayback contains the playback format
type RecordingMetadataPlayback struct {
	XMLName        xml.Name `xml:"playback"`
	Format         string   `xml:"format"`
	Link           string   `xml:"link"`
	ProcessingTime int      `xml:"processing_time"`
	Duration       int      `xml:"duration"`
	Size           int      `xml:"size"`
}
