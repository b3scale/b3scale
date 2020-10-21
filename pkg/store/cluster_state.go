package store

import ()

// The ClusterState holds the current backends
// and frontends in the cluster.
type ClusterState struct {
}

// GetBackendsOpts provides filtering options
// for the GetBackends operation
type GetBackendsOpts struct {
	FilterState string
}

// GetBackends retrives backends
func (c *ClusterState) GetBackends(opts *GetBackendsOpts) {

}

// GetBackend .... . .. . ..  provides filtering options
// for the GetBackend operation
func (c *ClusterState) GetBackend() {

}
