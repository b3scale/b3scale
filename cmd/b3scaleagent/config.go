package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/store"
)

// Errors
var (
	ErrServerURLNotInConfig = errors.New(
		"bigbluebutton.web.serverURL property not found in config")
	ErrSecretNotInConfig = errors.New(
		"securitySalt property not found in config")
)

// Well known config params
const (
	CfgWebServerURL = "bigbluebutton.web.serverURL"
	CfgSecret       = "securitySalt"
)

// Make a redis url from the BBB config
func configRedisURL(conf config.Properties) string {
	host, ok := conf.Get("redisHost")
	if !ok {
		host = "localhost"
	}
	port, ok := conf.Get("redisPort")
	if !ok {
		port = "6379"
	}
	pass, _ := conf.Get("redisPassword")

	return fmt.Sprintf(
		"redis://:%s@%s:%s/1",
		pass, host, port)
}

// Make a bbb BackendState from a BBB config
func backendFromConfig(
	conf config.Properties,
) (*store.BackendState, error) {
	serverURL, ok := conf.Get(CfgWebServerURL)
	if !ok {
		return nil, ErrServerURLNotInConfig
	}
	secret, ok := conf.Get(CfgSecret)
	if !ok {
		return nil, ErrSecretNotInConfig
	}

	// Append the api endpoint to the server URL
	apiURL := serverURL
	if !strings.HasSuffix(apiURL, "/") {
		apiURL += "/"
	}
	apiURL += "bigbluebutton/api/"

	// Make new backend
	state := store.InitBackendState(&store.BackendState{
		Backend: &bbb.Backend{
			Host:   apiURL,
			Secret: secret,
		},
		AdminState: "init",
		LoadFactor: config.GetLoadFactor(),
	})
	return state, nil
}
