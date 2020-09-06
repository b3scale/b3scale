package middleware

import (
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// BackendAPI Handler will invoke API calls
// on backends. The backend is selected
// by popping a backend label.
type BackendAPI struct {
	cluster *cluster.State
}

// ID of the handler
func (api *BackendAPI) ID() string {
	return "backendapi"
}

// Schema retrieves the update schema
func (api *BackendAPI) Schema() cluster.Schema {
	return cluster.Schema{}
}
