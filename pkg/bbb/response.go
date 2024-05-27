package bbb

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/auth"
)

var (
	// ErrCantBeMerged is the error when two responses
	// of the same type can not be merged, e.g. when
	// the data is not a collection.
	ErrCantBeMerged = errors.New(
		"responses of this type can not be merged")

	// ErrMergeConflict will be returned when two
	// responses differ in fields, where they should not.
	// Eg. a successful and a failed return code
	ErrMergeConflict = errors.New(
		"responses have conflicting values")
)

const (
	// RetSuccess is the success return code
	RetSuccess = "SUCCESS"

	// RetFailed is the failure return code
	RetFailed = "FAILED"
)

const (
	// StatePublished is the state of recording, when published
	StatePublished = "published"

	// StateUnpublished is the state of an unpublished recording
	StateUnpublished = "unpublished"

	// StateAny indicates that a recording may be in any state.
	// This is intended for querying.
	StateAny = "any"
)

// Response interface
type Response interface {
	Marshal() ([]byte, error)
	Merge(response Response) error

	Header() http.Header
	SetHeader(http.Header)

	Status() int
	SetStatus(int)

	IsSuccess() bool
}

// A XMLResponse from the server
type XMLResponse struct {
	XMLName    xml.Name `xml:"response"`
	Returncode string   `xml:"returncode"`
	Message    string   `xml:"message,omitempty"`
	MessageKey string   `xml:"messageKey,omitempty"`
	Version    string   `xml:"version,omitempty"`

	header http.Header
	status int
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

	res.status = other.status
	res.header = other.header
	res.Message = other.Message
	res.MessageKey = other.MessageKey
	res.Version = other.Version
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

// Make a new default header for XML responses
func (res *XMLResponse) makeDefaultHeader() http.Header {
	header := make(http.Header)
	header.Add("Content-Type", "application/xml")
	return header
}

// Header returns the HTTP response headers
func (res *XMLResponse) Header() http.Header {
	if res.header == nil {
		res.header = res.makeDefaultHeader()
	}
	return res.header
}

// SetHeader sets the HTTP response headers
func (res *XMLResponse) SetHeader(h http.Header) {
	res.header = h
}

// Status returns the HTTP response status code
func (res *XMLResponse) Status() int {
	return res.status
}

// SetStatus sets the HTTP response status code
func (res *XMLResponse) SetStatus(s int) {
	res.status = s
}

// IsSuccess checks if the returncode of the response
// is 'SUCCESS'.
func (res *XMLResponse) IsSuccess() bool {
	return res.Returncode == RetSuccess
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

// Header returns the HTTP response headers
func (res *CreateResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *CreateResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *CreateResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *CreateResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// JoinResponse of the join resource.
// WARNING: the join response might be a html page without
// any meaningful data.
type JoinResponse struct {
	*XMLResponse
	MeetingID    string `xml:"meeting_id"`
	UserID       string `xml:"user_id"`
	AuthToken    string `xml:"auth_token"`
	SessionToken string `xml:"session_token"`
	URL          string `xml:"url"`

	// The join response might be a raw
	raw []byte
}

// UnmarshalJoinResponse decodes the serialized XML data
func UnmarshalJoinResponse(data []byte) (*JoinResponse, error) {
	res := &JoinResponse{}
	err := xml.Unmarshal(data, res)
	if err != nil {
		res.XMLResponse = new(XMLResponse)
		res.raw = data
	}
	return res, nil
}

// IsRaw returns true if the response could
// not be decoded from XML data
func (res *JoinResponse) IsRaw() bool {
	return res.raw != nil
}

// SetRaw will set a raw content
func (res *JoinResponse) SetRaw(data []byte) {
	res.raw = data
}

// Marshal encodes a JoinResponse as XML
func (res *JoinResponse) Marshal() ([]byte, error) {
	if res.IsRaw() {
		return res.raw, nil
	}
	return xml.Marshal(res)
}

// Merge another response
func (res *JoinResponse) Merge(other Response) error {
	return ErrCantBeMerged
}

// Header returns the HTTP response headers
func (res *JoinResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *JoinResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *JoinResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *JoinResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
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

// Header returns the HTTP response headers
func (res *IsMeetingRunningResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *IsMeetingRunningResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *IsMeetingRunningResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *IsMeetingRunningResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
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

// Header returns the HTTP response headers
func (res *EndResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *EndResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *EndResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *EndResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
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

// Header returns the HTTP response headers
func (res *GetMeetingInfoResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *GetMeetingInfoResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *GetMeetingInfoResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *GetMeetingInfoResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// GetMeetingsResponse contains a list of meetings.
type GetMeetingsResponse struct {
	*XMLResponse
	Meetings []*Meeting `xml:"meetings>meeting"`
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

// Header returns the HTTP response headers
func (res *GetMeetingsResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *GetMeetingsResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *GetMeetingsResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *GetMeetingsResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
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

// Header returns the HTTP response headers
func (res *GetRecordingsResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *GetRecordingsResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *GetRecordingsResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *GetRecordingsResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
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

// Header returns the HTTP response headers
func (res *PublishRecordingsResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *PublishRecordingsResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *PublishRecordingsResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *PublishRecordingsResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
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

// Header returns the HTTP response headers
func (res *DeleteRecordingsResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *DeleteRecordingsResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *DeleteRecordingsResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *DeleteRecordingsResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
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

// Header returns the HTTP response headers
func (res *UpdateRecordingsResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *UpdateRecordingsResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *UpdateRecordingsResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *UpdateRecordingsResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
}

// GetDefaultConfigXMLResponse has the raw config xml data
type GetDefaultConfigXMLResponse struct {
	Config []byte

	header http.Header
	status int
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

// Header returns the HTTP response headers
func (res *GetDefaultConfigXMLResponse) Header() http.Header {
	return res.Header()
}

// SetHeader sets the HTTP response headers
func (res *GetDefaultConfigXMLResponse) SetHeader(h http.Header) {
	res.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *GetDefaultConfigXMLResponse) Status() int {
	return res.Status()
}

// SetStatus sets the HTTP response status code
func (res *GetDefaultConfigXMLResponse) SetStatus(s int) {
	res.SetStatus(s)
}

// IsSuccess checks if the returncode of the response
// is 'SUCCESS'.
func (res *GetDefaultConfigXMLResponse) IsSuccess() bool {
	return true
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

// Header returns the HTTP response headers
func (res *SetConfigXMLResponse) Header() http.Header {
	return res.XMLResponse.Header()
}

// SetHeader sets the HTTP response headers
func (res *SetConfigXMLResponse) SetHeader(h http.Header) {
	res.XMLResponse.SetHeader(h)
}

// Status returns the HTTP response status code
func (res *SetConfigXMLResponse) Status() int {
	return res.XMLResponse.Status()
}

// SetStatus sets the HTTP response status code
func (res *SetConfigXMLResponse) SetStatus(s int) {
	res.XMLResponse.SetStatus(s)
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

	header http.Header
	status int
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

// Header returns the HTTP response headers
func (res *GetRecordingTextTracksResponse) Header() http.Header {
	return res.header
}

// SetHeader sets the HTTP response header
func (res *GetRecordingTextTracksResponse) SetHeader(h http.Header) {
	res.header = h
}

// Status returns the HTTP response status code
func (res *GetRecordingTextTracksResponse) Status() int {
	return res.status
}

// SetStatus sets the HTTP response status code
func (res *GetRecordingTextTracksResponse) SetStatus(s int) {
	res.status = s
}

// IsSuccess checks if the returncode of the response
// is 'SUCCESS'.
func (res *GetRecordingTextTracksResponse) IsSuccess() bool {
	return res.Returncode == RetSuccess
}

// PutRecordingTextTrackResponse is the response when uploading
// a text track. Response is in JSON.
type PutRecordingTextTrackResponse struct {
	Returncode string `json:"returncode"`
	MessageKey string `json:"messageKey,omitempty"`
	Message    string `json:"message,omitempty"`
	RecordID   string `json:"recordId,omitempty"`

	header http.Header
	status int
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

// Header returns the HTTP response headers
func (res *PutRecordingTextTrackResponse) Header() http.Header {
	return res.header
}

// SetHeader sets the HTTP response header
func (res *PutRecordingTextTrackResponse) SetHeader(h http.Header) {
	res.header = h
}

// Status returns the HTTP response status code
func (res *PutRecordingTextTrackResponse) Status() int {
	return res.status
}

// SetStatus sets the HTTP response status code
func (res *PutRecordingTextTrackResponse) SetStatus(s int) {
	res.status = s
}

// IsSuccess checks if the returncode of the response
// is 'SUCCESS'.
func (res *PutRecordingTextTrackResponse) IsSuccess() bool {
	return res.Returncode == RetSuccess
}

// Breakout info
type Breakout struct {
	XMLName         xml.Name `xml:"breakout" json:"-"`
	ParentMeetingID string   `xml:"parentMeetingID"`
	Sequence        int      `xml:"sequence"`
	FreeJoin        bool     `xml:"freeJoin"`
}

// Attendee of a meeting
type Attendee struct {
	XMLName         xml.Name `xml:"attendee" json:"-"`
	UserID          string   `xml:"userID"`
	InternalUserID  string   `xml:"internalUserID,omit"`
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
	XMLName               xml.Name  `xml:"meeting" json:"-"`
	MeetingName           string    `xml:"meetingName"`
	MeetingID             string    `xml:"meetingID"`
	InternalMeetingID     string    `xml:"internalMeetingID"`
	CreateTime            Timestamp `xml:"createTime"`
	CreateDate            string    `xml:"createDate"`
	VoiceBridge           string    `xml:"voiceBridge"`
	DialNumber            string    `xml:"dialNumber"`
	AttendeePW            string    `xml:"attendeePW"`
	ModeratorPW           string    `xml:"moderatorPW"`
	Running               bool      `xml:"running"`
	Duration              int       `xml:"duration"`
	Recording             bool      `xml:"recording"`
	HasBeenForciblyEnded  bool      `xml:"hasBeenForciblyEnded"`
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

// Update the meeting info with new data
func (m *Meeting) Update(update *Meeting) error {
	if m.MeetingID != update.MeetingID {
		return fmt.Errorf("meeting ids do not match for update")
	}
	if m.InternalMeetingID != update.InternalMeetingID {
		return fmt.Errorf("internal ids do not match for update")
	}
	/*

		if len(update.MeetingName) > 0 {
			m.MeetingName = update.MeetingName
		}
		if len(update.CreateDate) > 0 {
			m.CreateDate = update.CreateDate
		}
		if len(update.VoiceBridge) > 0 {
			m.VoiceBridge = update.VoiceBridge
		}
		if len(update.DialNumber) > 0 {
			m.DialNumber = update.DialNumber
		}
		if len(update.AttendeePW) > 0 {
			m.AttendeePW = update.AttendeePW
		}
		if len(update.ModeratorPW) > 0 {
			m.ModeratorPW = update.ModeratorPW
		}
		m.Running = update.Running
		m.Duration = update.Duration
		m.Recording = update.Recording
		m.HasBeenForciblyEnded = update.HasBeenForciblyEnded
		m.StartTime = update.StartTime
		m.EndTime = update.EndTime
		m.ParticipantCount = update.ParticipantCount
		m.ListenerCount = update.ListenerCount
		m.VoiceParticipantCount = update.VoiceParticipantCount
		m.VideoCount = update.VideoCount
		m.MaxUsers = update.MaxUsers
		m.ModeratorCount = update.ModeratorCount
		m.IsBreakout = update.IsBreakout
		m.Attendees = update.Attendees
		m.BreakoutRooms = update.BreakoutRooms
	*/

	*m = *update

	return nil
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
