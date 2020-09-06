package bbb

// BBB API

// ApiResources
const (
	APIJoin                   = "join"
	APICreate                 = "create"
	APIIsMeetingRunning       = "isMeetingRunning"
	APIEnd                    = "end"
	APIGetMeetingInfo         = "getMeetingInfo"
	APIGetMeetings            = "getMeetings"
	APIGetRecordings          = "getRecordings"
	APIPublishRecordings      = "publishRecordings"
	APIDeleteRecordings       = "deleteRecordings"
	APIUpdateRecordings       = "updateRecordings"
	APIGetDefaultConfigXML    = "getDefaultConfigXML"
	APISetConfigXML           = "setConfigXML"
	APIGetRecordingTextTracks = "getRecordingTextTrack"
	APIPutRecordingTextTrack  = "putRecordingTextTrack"
)

// API is the bbb api interface
type API interface {
	Join(*Request) (*JoinResponse, error)
	Create(*Request) (*CreateResponse, error)
	IsMeetingRunning(*Request) (*IsMeetingRunningResponse, error)
	End(*Request) (*EndResponse, error)
	GetMeetingInfo(*Request) (*GetMeetingInfoResponse, error)
	GetMeetings(*Request) (*GetMeetingsResponse, error)
	GetRecordings(*Request) (*GetRecordingsResponse, error)
	PublishRecordings(*Request) (*PublishRecordingsResponse, error)
	DeleteRecordings(*Request) (*DeleteRecordingsResponse, error)
	UpdateRecordings(*Request) (*UpdateRecordingsResponse, error)
	GetDefaultConfigXML(*Request) (*GetDefaultConfigXMLResponse, error)
	SetConfigXML(*Request) (*SetConfigXMLResponse, error)
	GetRecordingTextTracks(*Request) (*GetRecordingTextTracksResponse, error)
	PutRecordingTextTrack(*Request) (*PutRecordingTextTrackResponse, error)
}
