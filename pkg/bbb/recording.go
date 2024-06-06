package bbb

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"time"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/rs/zerolog/log"
)

// Recording is a recorded bbb session
type Recording struct {
	XMLName           xml.Name  `xml:"recording"`
	RecordID          string    `xml:"recordID"`
	MeetingID         string    `xml:"meetingID"`
	InternalMeetingID string    `xml:"internalMeetingID"`
	Name              string    `xml:"name"`
	IsBreakout        bool      `xml:"isBreakout"`
	Published         bool      `xml:"published"`
	State             string    `xml:"state"`
	StartTime         Timestamp `xml:"startTime"`
	EndTime           Timestamp `xml:"endTime"`
	Participants      int       `xml:"participants"`
	Metadata          Metadata  `xml:"metadata"`
	Formats           []*Format `xml:"playback>format"`
}

// updateHostURL replaces the host and schema of a URL
func updateHostURL(target, base string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return target // nothing we can do here
	}

	targetURL, err := url.Parse(target)
	if err != nil {
		return target // same
	}

	targetURL.Scheme = baseURL.Scheme
	targetURL.Host = baseURL.Host

	return targetURL.String()
}

// SetPlaybackHost will update the link to the presentation
// and preview thumbnails
func (r *Recording) SetPlaybackHost(host string) {
	for _, f := range r.Formats {

		// Update recording host
		f.URL = updateHostURL(f.URL, host)

		if f.Preview == nil || f.Preview.Images == nil {
			continue
		}

		// Update preview host
		for _, img := range f.Preview.Images.All {
			img.URL = updateHostURL(img.URL, host)
		}
	}
}

// Protect will update the link to the presentation
// to point back to the b3scale instance, with a request
// token that will be exchanged into an access token.
//
// The default lifetime is an hour.
//
// As a subject, the frontendID will most likely be used,
// but it could be any identifier.
func (r *Recording) Protect(subject string) {
	apiURL := config.MustEnv(config.EnvAPIURL)
	secret := config.MustEnv(config.EnvJWTSecret)

	for _, f := range r.Formats {
		// Create resource token and update target URL.
		// A note on the token lifetime:
		//  - The link is actually shareable for this time.
		//  - Scalelite uses a default of 60 minutes.
		//  - Maybe make configurable.
		resource := auth.EncodeResource(f.Type, r.RecordID)
		token, err := auth.NewClaims(subject).
			WithLifetime(60 * time.Minute).
			WithAudience(resource).
			Sign(secret)
		if err != nil {
			panic(err)
		}

		// Protected URL
		f.URL = fmt.Sprintf(
			"%s/api/v1/protected/recordings/%s",
			apiURL,
			token)
	}
}

// GetFormat returns the format with the given type
func (r *Recording) GetFormat(format string) *Format {
	for _, f := range r.Formats {
		if f.Type == format {
			return f
		}
	}
	return nil
}

// Merge two recordings
func (r *Recording) Merge(other *Recording) error {
	if other.RecordID != r.RecordID {
		return fmt.Errorf("RecordID must match for merge")
	}
	if other.MeetingID != "" {
		r.MeetingID = other.MeetingID
		r.IsBreakout = other.IsBreakout
	}
	if other.InternalMeetingID != "" {
		r.InternalMeetingID = other.InternalMeetingID
	}
	if other.Name != "" {
		r.Name = other.Name
	}
	r.State = other.State
	r.Published = other.Published
	if other.Participants > 0 {
		r.Participants = other.Participants
	}
	r.Metadata = other.Metadata

	// And append all formats from the other state, if not
	// already present.
	for _, f := range other.Formats {
		present := false
		for _, f2 := range r.Formats {
			if f2.Type == f.Type {
				present = true
				break
			}
		}
		if !present {
			r.Formats = append(r.Formats, f)
		}
	}

	return nil
}

// Format contains a link to the playable media
type Format struct {
	XMLName        xml.Name `xml:"format"`
	Type           string   `xml:"type"`
	URL            string   `xml:"url"`
	ProcessingTime int      `xml:"processingTime"` // No idea. The example is 7177.
	Length         int      `xml:"length"`
	Preview        *Preview `xml:"preview"`
}

// Preview contains a list of images
type Preview struct {
	XMLName xml.Name `xml:"preview"`
	Images  *Images  `xml:"images"`
}

// Images is a collection of Image
type Images struct {
	XMLName xml.Name `xml:"images"`
	All     []*Image `xml:"image"`
}

// Image is a preview image of the format
type Image struct {
	XMLName xml.Name `xml:"image"`
	Alt     string   `xml:"alt,attr,omitempty"`
	Height  int      `xml:"height,attr,omitempty"`
	Width   int      `xml:"width,attr,omitempty"`
	URL     string   `xml:",chardata"`
}

// TextTrack of a Recording
type TextTrack struct {
	Href   string `json:"href"`
	Kind   string `json:"kind"`
	Label  string `json:"label"`
	Source string `json:"source"`
}

// RecordingMetadata can be parsed from a metadata.xml
// by posting it to the API endpoint in the bbb recordings
// hook.
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
	meetingID, ok := m.Meta["meetingId"]
	if !ok {
		log.Warn().Msg("RecordingMetadata: meetingId not found in metadata")
	}
	name, ok := m.Meta["meetingName"]
	if !ok {
		log.Warn().Msg("RecordingMetadata: meetingName not found in metadata")
	}
	isBreakoutStr, ok := m.Meta["isBreakout"]
	if !ok {
		log.Warn().Msg("RecordingMetadata: isBreakout not found in metadata")
	}
	isBreakout := isBreakoutStr == "true"

	r := &Recording{
		RecordID:          m.RecordID,
		MeetingID:         meetingID,
		InternalMeetingID: "DEPRECATED:" + meetingID, // Satisfy indices
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
