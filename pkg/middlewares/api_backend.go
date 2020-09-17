package middleware

import (
	"context"
	"fmt"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// NewAPIBackend creates the request handler
// middleware for invoking API calls on a backend.
//
// The backend is retrieved from the request context
// and must implement the bbb.API interface.
func NewAPIBackend() cluster.RequestMiddleware {
	return func(_next cluster.RequestHandler) cluster.RequestHandler {
		return func(ctx context.Context, req *bbb.Request) (bbb.Response, error) {
			// Get backend by id
			backend := cluster.BackendFromContext(ctx)
			if backend == nil {
				return nil, fmt.Errorf("no backend in context")
			}

			// Dispatch API resources
			switch req.Resource {
			case bbb.ResJoin:
				return backend.Join(req.Request)
			case bbb.ResCreate:
				return backend.Create(req.Request)
			case bbb.ResIsMeetingRunning:
				return backend.IsMeetingRunning(req.Request)
			case bbb.ResEnd:
				return backend.End(req.Request)
			case bbb.ResGetMeetingInfo:
				return backend.GetMeetingInfo(req.Request)
			case bbb.ResGetMeetings:
				return backend.GetMeetings(req.Request)
			case bbb.ResGetRecordings:
				return backend.GetRecordings(req.Request)
			case bbb.ResPublishRecordings:
				return backend.PublishRecordings(req.Request)
			case bbb.ResDeleteRecordings:
				return backend.DeleteRecordings(req.Request)
			case bbb.ResUpdateRecordings:
				return backend.UpdateRecordings(req.Request)
			case bbb.ResGetDefaultConfigXML:
				return backend.GetDefaultConfigXML(req.Request)
			case bbb.ResSetConfigXML:
				return backend.SetConfigXML(req.Request)
			case bbb.ResGetRecordingTextTracks:
				return backend.GetRecordingTextTracks(req.Request)
			case bbb.ResPutRecordingTextTrack:
				return backend.PutRecordingTextTrack(req.Request)
			}
			// We could not dispatch this
			return nil, fmt.Errorf("unknown resource: %s",
				req.Resource)
		}
	}
}
