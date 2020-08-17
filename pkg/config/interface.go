package config

import "strings"

// Backend is the configuration of a cluster node
type Backend struct {
	Host   string
	Secret string
}

// NewBackend creates a new Backend configuration
// and normalizes the Host url
func NewBackend(host, secret string) *Backend {
	// Normalize host URL by appending a trailing slash
	// in case it was not provided in the config.
	if !strings.HasSuffix(host, "/") {
		host = host + "/"
	}
	return &Backend{
		Host:   host,
		Secret: secret,
	}
}

// Frontend is the configuration of a
// a cluster consumer.
type Frontend struct {
	Key    string
	Secret string
}

// NewFrontend creates a new frontend config
func NewFrontend(key, secret string) *Frontend {
	return &Frontend{
		Key:    key,
		Secret: secret,
	}
}

// BackendsConfig interface
type BackendsConfig interface {
	Load() ([]*Backend, error)
}

// FrontendsConfig interface
type FrontendsConfig interface {
	Load() ([]*Frontend, error)
}
