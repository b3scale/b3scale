package routing

import (
	"log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// Lookup middleware for retriving an already associated
// backend for a given meeting.
func Lookup(ctrl *cluster.Controller) cluster.RouterMiddleware {
	return func(next cluster.RouterHandler) cluster.RouterHandler {
		return func(
			backends []*cluster.Backend, req *bbb.Request,
		) ([]*cluster.Backend, error) {
			// Get meeting id from params. If none is present,
			// there is nothing to do for us here.
			meetingID, ok := req.Params.MeetingID()
			if !ok {
				return backends, nil
			}

			// Lookup backend for meeting in cluster, use backend
			// if there is one associated - otherwise return
			// all possible backends.
			backend, err := ctrl.GetBackend(store.Q().
				Join("meetings ON meetings.backend_id = backends.id").
				Where("meetings.id = ?", meetingID))
			if err != nil {
				return nil, err
			}
			if backend == nil {
				// No specific backend was associated with the ID
				return backends, nil
			}
			// Check if backend was not already filtered out
			// by a previous middleware stil, otherwise
			// delete, association and return all possible backends.
			found := false
			for _, b := range backends {
				if b == backend {
					found = true
					break
				}
			}
			if !found {
				// Emit a warning
				log.Println(
					"WARNING: Requested backend", backend,
					"is no longer available",
					"as selectable routing target.",
					"Reassigning meeting:", meetingID)
				return backends, nil
			}

			// Return backend only
			return []*cluster.Backend{backend}, nil
		}
	}
}
