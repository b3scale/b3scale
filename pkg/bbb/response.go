package bbb

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
)

// ErrCantBeMerged is the error when two responses
// of the same type can not be merged, e.g. when
// the data is not a collection.
var ErrCantBeMerged = errors.New(
	"responses of this type can not be merged")

// ErrMergeConflict will be returned when two
// responses differ in fields, where they should not.
// Eg. a successful and a failed return code
var ErrMergeConflict = errors.New(
	"responses have conflicting values")

// Response interface
type Response interface {
	Marshal() ([]byte, error)
	Merge(response Response) error
}

// A XMLResponse from the server
type XMLResponse struct {
	XMLName    xml.Name `xml:"response"`
	Returncode string   `xml:"returncode"`
	Message    string   `xml:"message,omitempty"`
	MessageKey string   `xml:"messageKey,omitempty"`
}

// MergeXMLResponse is a specific merge
func (res *XMLResponse) MergeXMLResponse(other *XMLResponse) error {
	if res.Returncode != other.Returncode {
		return ErrMergeConflict
	}
	if res.Message != "" && res.Message != other.Message {
		return ErrMergeConflict
	}
	if res.MessageKey != "" && res.MessageKey != other.MessageKey {
		return ErrMergeConflict
	}

	res.Message = other.Message
	res.MessageKey = other.MessageKey
	return nil
}

// Merge XMLResponses.
// However, in general this should not be merged.
func (res *XMLResponse) Merge(other Response) error {
	return ErrCantBeMerged
}

// Marshal a XMLResponse to XML
func (res *XMLResponse) Marshal() ([]byte, error) {
	data, err := xml.Marshal(res)
	return data, err
}

// CreateResponse is the resonse for the `create` API resource.
type CreateResponse struct {
	*XMLResponse
	*Meeting
}

// UnmarshalCreateResponse decodes the resonse XML data.
func UnmarshalCreateResponse(data []byte) (*CreateResponse, error) {
	res := &CreateResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal a CreateResponse to XML
func (res *CreateResponse) Marshal() ([]byte, error) {
	data, err := xml.Marshal(res)
	return data, err
}

// Merge another response
func (res *CreateResponse) Merge(other Response) error {
	return ErrCantBeMerged
}

// JoinResponse of the join resource
type JoinResponse struct {
	*XMLResponse
	MeetingID    string `xml:"meeting_id"`
	UserID       string `xml:"user_id"`
	AuthToken    string `xml:"auth_token"`
	SessionToken string `xml:"session_token"`
	URL          string `xml:"url"`
}

// UnmarshalJoinResponse decodes the serialized XML data
func UnmarshalJoinResponse(data []byte) (*JoinResponse, error) {
	res := &JoinResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal encodes a JoinResponse as XML
func (res *JoinResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Merge another response
func (res *JoinResponse) Merge(other Response) error {
	return ErrCantBeMerged
}

// IsMeetingRunningResponse is a meeting status resonse
type IsMeetingRunningResponse struct {
	*XMLResponse
	Running bool `xml:"running"`
}

// UnmarshalIsMeetingRunningResponse decodes the XML data.
func UnmarshalIsMeetingRunningResponse(
	data []byte,
) (*IsMeetingRunningResponse, error) {
	res := &IsMeetingRunningResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal a IsMeetingRunningResponse to XML
func (res *IsMeetingRunningResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Merge IsMeetingRunning responses
func (res *IsMeetingRunningResponse) Merge(other Response) error {
	return ErrCantBeMerged
}

// EndResponse is the resonse of the end resource
type EndResponse struct {
	*XMLResponse
}

// UnmarshalEndResponse decodes the xml resonse
func UnmarshalEndResponse(data []byte) (*EndResponse, error) {
	res := &EndResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal EndResponse to XML
func (res *EndResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Merge EndResponses
func (res *EndResponse) Merge(other Response) error {
	return ErrCantBeMerged
}

// GetMeetingInfoResponse contains detailed meeting information
type GetMeetingInfoResponse struct {
	*XMLResponse
	*Meeting
}

// UnmarshalGetMeetingInfoResponse decodes the xml response
func UnmarshalGetMeetingInfoResponse(
	data []byte,
) (*GetMeetingInfoResponse, error) {
	res := &GetMeetingInfoResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal GetMeetingInfoResponse to XML
func (res *GetMeetingInfoResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Merge GetMeetingInfoResponse
func (res *GetMeetingInfoResponse) Merge(other Response) error {
	return ErrCantBeMerged
}

// GetMeetingsResponse contains a list of meetings.
type GetMeetingsResponse struct {
	*XMLResponse
	Meetings MeetingsCollection `xml:"meetings>meeting"`
}

// UnmarshalGetMeetingsResponse decodes the xml response
func UnmarshalGetMeetingsResponse(
	data []byte,
) (*GetMeetingsResponse, error) {
	res := &GetMeetingsResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal serializes the response as XML
func (res *GetMeetingsResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Merge get meetings responses
func (res *GetMeetingsResponse) Merge(other Response) error {
	otherRes, ok := other.(*GetMeetingsResponse)
	if !ok {
		return ErrCantBeMerged
	}

	// Check envelope
	err := res.XMLResponse.MergeXMLResponse(otherRes.XMLResponse)
	if err != nil {
		return err
	}
	// Merge meetings lists by appending
	res.Meetings = append(res.Meetings, otherRes.Meetings...)
	return nil
}

// GetRecordingsResponse is the response of the getRecordings resource
type GetRecordingsResponse struct {
	*XMLResponse
	Recordings []*Recording `xml:"recordings>recording"`
}

// UnmarshalGetRecordingsResponse deserializes the response XML
func UnmarshalGetRecordingsResponse(
	data []byte,
) (*GetRecordingsResponse, error) {
	res := &GetRecordingsResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal a GetRecordingsResponse to XML
func (res *GetRecordingsResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Merge another GetRecordingsResponse
func (res *GetRecordingsResponse) Merge(other Response) error {
	otherRes, ok := other.(*GetRecordingsResponse)
	if !ok {
		return ErrCantBeMerged
	}
	err := res.XMLResponse.Merge(otherRes.XMLResponse)
	if err != nil {
		return err
	}
	res.Recordings = append(res.Recordings, otherRes.Recordings...)
	return nil
}

// PublishRecordingsResponse indicates if the recordings
// were published. This also has the potential for
// tasks failed successfully.
// Also the endpoint is designed badly because you can send
// a set of recordings and receive just a single published
// true or false.
type PublishRecordingsResponse struct {
	*XMLResponse
	Published bool `xml:"published"`
}

// UnmarshalPublishRecordingsResponse decodes the XML response
func UnmarshalPublishRecordingsResponse(
	data []byte,
) (*PublishRecordingsResponse, error) {
	res := &PublishRecordingsResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal a publishRecodingsResponse to XML
func (res *PublishRecordingsResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Merge a PublishRecordingsResponse
func (res *PublishRecordingsResponse) Merge(other Response) error {
	// This is kind of meh... I guess this is mergable
	// as it needs to be dispatched to other instances...
	otherRes, ok := other.(*PublishRecordingsResponse)
	if !ok {
		return ErrCantBeMerged
	}
	// Envelope
	err := res.XMLResponse.Merge(otherRes.XMLResponse)
	if err != nil {
		return err
	}
	// Payload
	if res.Published != otherRes.Published {
		return ErrMergeConflict
	}

	return nil
}

// DeleteRecordingsResponse indicates if the recording
// was correctly deleted. Might fail successfully.
// Same crap as with the publish resource
type DeleteRecordingsResponse struct {
	*XMLResponse
	Deleted bool `xml:"deleted"`
}

// UnmarshalDeleteRecordingsResponse decodes XML resource response
func UnmarshalDeleteRecordingsResponse(
	data []byte,
) (*DeleteRecordingsResponse, error) {
	res := &DeleteRecordingsResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal encodes the delete recordings response as XML
func (res *DeleteRecordingsResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Merge a DeleteRecordingsResponse
func (res *DeleteRecordingsResponse) Merge(other Response) error {
	otherRes, ok := other.(*DeleteRecordingsResponse)
	if !ok {
		return ErrCantBeMerged
	}
	// Envelope
	err := res.XMLResponse.Merge(otherRes.XMLResponse)
	if err != nil {
		return err
	}
	// Payload
	if res.Deleted != otherRes.Deleted {
		return ErrMergeConflict
	}
	return nil
}

// UpdateRecordingsResponse indicates if the update was successful
// in the attribute updated. Might be different from Returncode.
// I guess.
type UpdateRecordingsResponse struct {
	*XMLResponse
	Updated bool `xml:"updated"`
}

// UnmarshalUpdateRecordingsResponse decodes the XML data
func UnmarshalUpdateRecordingsResponse(
	data []byte,
) (*UpdateRecordingsResponse, error) {
	res := &UpdateRecordingsResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal UpdateRecordingsResponse to XML
func (res *UpdateRecordingsResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Merge a UpdateRecordingsResponse
func (res *UpdateRecordingsResponse) Merge(other Response) error {
	otherRes, ok := other.(*UpdateRecordingsResponse)
	if !ok {
		return ErrCantBeMerged
	}
	// Envelope
	err := res.XMLResponse.Merge(otherRes.XMLResponse)
	if err != nil {
		return err
	}
	// Payload
	if res.Updated != otherRes.Updated {
		return ErrMergeConflict
	}
	return nil
}

// GetDefaultConfigXMLResponse has the raw config xml data
type GetDefaultConfigXMLResponse struct {
	Config []byte
}

// UnmarshalGetDefaultConfigXMLResponse creates a new response
// from the data.
func UnmarshalGetDefaultConfigXMLResponse(
	data []byte,
) (*GetDefaultConfigXMLResponse, error) {
	return &GetDefaultConfigXMLResponse{
		Config: data,
	}, nil
}

// Marshal GetDefaultConfigXMLResponse encodes the response
// body which is just the data.
func (res *GetDefaultConfigXMLResponse) Marshal() ([]byte, error) {
	if res.Config == nil {
		return nil, fmt.Errorf("no config is set in response")
	}
	return res.Config, nil
}

// Merge GetDefaultConfigXMLResponse
func (res *GetDefaultConfigXMLResponse) Merge(other Response) error {
	return ErrCantBeMerged
}

// SetConfigXMLResponse encodes the result of setting the config
type SetConfigXMLResponse struct {
	*XMLResponse
	Token string `xml:"token"`
}

// UnmarshalSetConfigXMLResponse decodes the XML data
func UnmarshalSetConfigXMLResponse(
	data []byte,
) (*SetConfigXMLResponse, error) {
	res := &SetConfigXMLResponse{}
	err := xml.Unmarshal(data, res)
	return res, err
}

// Marshal encodes a SetConfigXMLResponse as XML
func (res *SetConfigXMLResponse) Marshal() ([]byte, error) {
	return xml.Marshal(res)
}

// Merge SetConfigXMLResponse
func (res *SetConfigXMLResponse) Merge(other Response) error {
	return ErrCantBeMerged
}

// JSONResponse encapsulates a json reponse
type JSONResponse struct {
	Response interface{} `json:"response"`
}

// GetRecordingTextTracksResponse lists all tracks
type GetRecordingTextTracksResponse struct {
	Returncode string       `json:"returncode"`
	MessageKey string       `json:"messageKey,omitempty"`
	Message    string       `json:"message,omitempty"`
	Tracks     []*TextTrack `json:"tracks"`
}

// UnmarshalGetRecordingTextTracksResponse decodes the json
func UnmarshalGetRecordingTextTracksResponse(
	data []byte,
) (*GetRecordingTextTracksResponse, error) {
	res := &JSONResponse{
		Response: &GetRecordingTextTracksResponse{},
	}
	err := json.Unmarshal(data, res)
	return res.Response.(*GetRecordingTextTracksResponse), err
}

// Marshal GetRecordingTextTracksResponse to JSON
func (res *GetRecordingTextTracksResponse) Marshal() ([]byte, error) {
	wrap := &JSONResponse{Response: res}
	return json.Marshal(wrap)
}

// Merge GetRecordingTextTracksResponse
func (res *GetRecordingTextTracksResponse) Merge(other Response) error {

	otherRes, ok := other.(*GetRecordingTextTracksResponse)
	if !ok {
		return ErrCantBeMerged
	}
	// Envelope
	if res.Returncode != otherRes.Returncode {
		return ErrMergeConflict
	}
	if res.Message != "" && res.Message != otherRes.Message {
		return ErrMergeConflict
	}
	if res.MessageKey != "" && res.MessageKey != otherRes.MessageKey {
		return ErrMergeConflict
	}
	res.Message = otherRes.Message
	res.MessageKey = otherRes.MessageKey
	// Payload
	res.Tracks = append(res.Tracks, otherRes.Tracks...)
	return nil
}

// PutRecordingTextTrackResponse is the response when uploading
// a text track. Response is in JSON.
type PutRecordingTextTrackResponse struct {
	Returncode string `json:"returncode"`
	MessageKey string `json:"messageKey,omitempty"`
	Message    string `json:"message,omitempty"`
	RecordID   string `json:"recordId,omitempty"`
}

// UnmarshalPutRecordingTextTrackResponse decodes the json response
func UnmarshalPutRecordingTextTrackResponse(
	data []byte,
) (*PutRecordingTextTrackResponse, error) {
	res := &JSONResponse{
		Response: &PutRecordingTextTrackResponse{},
	}
	err := json.Unmarshal(data, res)
	return res.Response.(*PutRecordingTextTrackResponse), err
}

// Marshal a PutRecordingTextTrackResponse to JSON
func (res *PutRecordingTextTrackResponse) Marshal() ([]byte, error) {
	wrap := &JSONResponse{Response: res}
	return json.Marshal(wrap)
}

// Merge a put recording text track
func (res *PutRecordingTextTrackResponse) Merge(other Response) error {
	return ErrCantBeMerged
}

// Breakout info
type Breakout struct {
	XMLName         xml.Name `xml:"breakout"`
	ParentMeetingID string   `xml:"parentMeetingID"`
	Sequence        int      `xml:"sequence"`
	FreeJoin        bool     `xml:"freeJoin"`
}

// Attendee of a meeting
type Attendee struct {
	XMLName         xml.Name `xml:"attendee"`
	UserID          string   `xml:"userID"`
	FullName        string   `xml:"fullName"`
	Role            string   `xml:"role"`
	IsPresenter     bool     `xml:"isPresenter"`
	IsListeningOnly bool     `xml:"isListeningOnly"`
	HasJoinedVoice  bool     `xml:"hasJoinedVoice"`
	HasVideo        bool     `xml:"hasVideo"`
	ClientType      string   `xml:"clientType"`
}

// Meeting information
type Meeting struct {
	XMLName               xml.Name  `xml:"meeting"`
	MeetingName           string    `xml:"meetingName"`
	MeetingID             string    `xml:"meetingID"`
	InternalMeetingID     string    `xml:"internalMeetingID"`
	CreateTime            Timestamp `xml:"createTime"`
	CreateDate            string    `xml:"createDate"`
	VoiceBridge           string    `xml:"voiceBridge"`
	DialNumber            string    `xml:"dialNumber"`
	AttendeePW            string    `xml:"attendeePW"`
	ModeratorPW           string    `xml:"moderatorPW"`
	Running               string    `xml:"running"`
	Duration              int       `xml:"duration"`
	Recording             string    `xml:"recording"`
	HasBeenForciblyEnded  string    `xml:"hasBeenForciblyEnded"`
	StartTime             Timestamp `xml:"startTime"`
	EndTime               Timestamp `xml:"endTime"`
	ParticipantCount      int       `xml:"participantCount"`
	ListenerCount         int       `xml:"listenerCount"`
	VoiceParticipantCount int       `xml:"voiceParticipantCount"`
	VideoCount            int       `xml:"videoCount"`
	MaxUsers              int       `xml:"maxUsers"`
	ModeratorCount        int       `xml:"moderatorCount"`
	IsBreakout            bool      `xml:"isBreakout"`

	Metadata Metadata `xml:"metadata"`

	Attendees     []*Attendee `xml:"attendees>attendee"`
	BreakoutRooms []string    `xml:"breakoutRooms>breakout"`
	Breakout      *Breakout   `xml:"breakout"`
}

func (m *Meeting) String() string {
	return fmt.Sprintf(
		"[Meeting id: %v, pc: %v, mc: %v, running: %v]",
		m.MeetingID, m.ParticipantCount, m.ModeratorCount, m.Running,
	)
}

// MeetingsCollection is a slice of meetings
type MeetingsCollection []*Meeting

// GetMeetingByID retrieves a meeting by its id
func (col MeetingsCollection) GetMeetingByID(id string) *Meeting {
	for _, m := range col {
		if m.MeetingID == id {
			return m
		}
	}
	return nil
}

// Update the collection will either replace a
// present meeting with the new one identified by ID,
// or will append the meeting.
func (col MeetingsCollection) Update(m *Meeting) MeetingsCollection {
	if col.GetMeetingByID(m.MeetingID) == nil {
		return append(col, m)
	}

	// Replace meeting
	meetings := make(MeetingsCollection, 0, len(col))
	for _, cur := range col {
		if cur.MeetingID == m.MeetingID {
			meetings = append(meetings, m) // Replace
		} else {
			meetings = append(meetings, cur)
		}
	}
	return meetings
}

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
	Alt     string   `xml:"alt,attr"`
	Height  int      `xml:"height,attr"`
	Width   int      `xml:"width,attr"`
}

// TextTrack of a Recording
type TextTrack struct {
	Href   string `json:"href"`
	Kind   string `json:"kind"`
	Label  string `json:"label"`
	Source string `json:"source"`
}
