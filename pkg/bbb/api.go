package bbb

// BBB API

// ApiResources
const (
	ResJoin                   = "join"
	ResCreate                 = "create"
	ResIsMeetingRunning       = "isMeetingRunning"
	ResEnd                    = "end"
	ResGetMeetingInfo         = "getMeetingInfo"
	ResGetMeetings            = "getMeetings"
	ResGetRecordings          = "getRecordings"
	ResPublishRecordings      = "publishRecordings"
	ResDeleteRecordings       = "deleteRecordings"
	ResUpdateRecordings       = "updateRecordings"
	ResGetDefaultConfigXML    = "getDefaultConfigXML"
	ResSetConfigXML           = "setConfigXML"
	ResGetRecordingTextTracks = "getRecordingTextTracks"
	ResPutRecordingTextTrack  = "putRecordingTextTrack"
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
