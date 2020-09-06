package cluster

import (
	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// A Request is a request to the cluster, containing
// the BBB api request.
type Request struct {
	*bbb.Request
	Labels
}
