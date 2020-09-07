package middleware

import (
	"fmt"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// NewAPIBackend creates the request handler
// middleware for invoking API calls on a backend.
//
// The backend is retrieved from the request context
// and must implement the bbb.API interface.
func NewAPIBackend() cluster.MiddlewareFunc {
	return func(_next cluster.HandlerFunc) cluster.HandlerFunc {
		return func(req *cluster.Request) (cluster.Response, error) {
			// Get backend by id
			backend, ok := req.Context.Load("backend")
			if !ok {
				return nil, fmt.Errorf("no backend in context")
			}
			api := backend.(bbb.API)

			// Dispatch API resources
			switch req.Resource {
			case bbb.ResJoin:
				return api.Join(req.Request)
			case bbb.ResCreate:
				return api.Create(req.Request)
			case bbb.ResIsMeetingRunning:
				return api.IsMeetingRunning(req.Request)
			case bbb.ResEnd:
				return api.End(req.Request)
			case bbb.ResGetMeetingInfo:
				return api.GetMeetingInfo(req.Request)
			case bbb.ResGetMeetings:
				return api.GetMeetings(req.Request)
			case bbb.ResGetRecordings:
				return api.GetRecordings(req.Request)
			case bbb.ResPublishRecordings:
				return api.PublishRecordings(req.Request)
			case bbb.ResDeleteRecordings:
				return api.DeleteRecordings(req.Request)
			case bbb.ResUpdateRecordings:
				return api.UpdateRecordings(req.Request)
			case bbb.ResGetDefaultConfigXML:
				return api.GetDefaultConfigXML(req.Request)
			case bbb.ResSetConfigXML:
				return api.SetConfigXML(req.Request)
			case bbb.ResGetRecordingTextTracks:
				return api.GetRecordingTextTracks(req.Request)
			case bbb.ResPutRecordingTextTrack:
				return api.PutRecordingTextTrack(req.Request)
			}
			// We could not dispatch this
			return nil, fmt.Errorf("unknown resource: %s",
				req.Resource)
		}
	}
}
