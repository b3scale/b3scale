package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
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

// Try to resolve the backend state in the cluster by
// serverURL und secret we have in the config.
// Update the secret if it was changed.
func configToBackendState(
	pool *pgxpool.Pool,
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

	// Try to get backend
	state, err := store.GetBackendState(pool, store.Q().
		Where("host ILIKE ?", serverURL+"%"))
	if err != nil {
		return nil, err
	}

	if state == nil {
		return nil, nil
	}

	// Make sure the secret is up to date
	if state.Backend.Secret != secret {
		log.Warn().
			Str("backendID", state.ID).
			Str("host", state.Backend.Host).
			Msg("updating changed secret for backend")

		state.Backend.Secret = secret
		if err := state.Save(); err != nil {
			return nil, err
		}
	}

	return state, nil
}

// Register a new backend at the cluster with
// data derived from the config
func configRegisterBackendState(
	pool *pgxpool.Pool,
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

	// Register new backend
	state := store.InitBackendState(pool, &store.BackendState{
		Backend: &bbb.Backend{
			Host:   apiURL,
			Secret: secret,
		},
	})
	if err := state.Save(); err != nil {
		return nil, err
	}

	log.Info().
		Str("backendID", state.ID).
		Str("host", state.Backend.Host).
		Msg("registered new backend")

	return state, nil
}
