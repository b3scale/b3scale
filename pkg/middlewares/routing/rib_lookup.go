package routing

import (
	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// RIBLookup middleware for retriving a
// backend for a given meeting id from the RIB
func RIBLookup(rib cluster.RIB) cluster.RouterMiddleware {
	return func(next cluster.RouterHandler) cluster.RouterHandler {
		return func(
			backends []*cluster.Backend, req *bbb.Request,
		) ([]*cluster.Backend, error) {
			// Get meeting id from params. If none is present,
			// there is nothing to do for us here.
			id, ok := req.Params.GetMeetingID()
			if !ok {
				return backends, nil
			}

			// Lookup meeting id in RIB, use backend
			// if there is one associated - otherwise return
			// all possible backends.
			meeting := bbb.Meeting{MeetingID: id}
			backend, err := rib.GetBackend(meeting)
			if err != nil {
				return nil, err
			}
			if meeting == nil {
				// No specific backend was associated with the ID
				return backends, nil
			}
			// Check if backend is still available, otherwise
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
				log.Warn(
					"Requested backend", backend, "is not longer available",
					"as selectable routing target.",
					"Reassigning meeting:", meeting)
				return backends, nil
			}

			// Return backend only
			return []*cluster.Backend{backend}, nil
		}
	}
}
