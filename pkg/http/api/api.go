package api

/*
B3Scale API v1

Administrative API for B3Scale. See /docs/rest_api.md for
details.
*/

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/b3scale/b3scale/pkg/store/schema"
)

const (
	// PrefixInternalID indicates that the ID is an 'internal' ID.
	PrefixInternalID = "internal:"
)

// InternalMeetingID returns the internal id for
// accessing via the API.
func InternalMeetingID(id string) string {
	return PrefixInternalID + id
}

// API extends the context and provides methods
// for handling the current user.
type API struct {
	// Authorization
	Scopes []string
	Ref    string
	// Database
	Conn *pgxpool.Conn

	echo.Context
}

// HasScope checks if the authentication scope claim
// contains a scope by name.
// The scope claim is a space separated list of scopes
// according to RFC8693, Section 4.2, (OAuth 2).
func (api *API) HasScope(s string) (found bool) {
	for _, sc := range api.Scopes {
		if sc == s {
			return true
		}
	}
	return false
}

// Ctx is a shortcut to access the request context
func (api *API) Ctx() context.Context {
	return api.Request().Context()
}

// ParamID is a shortcut to access the ID parameter.
// If the parameter is prefixed with `internal:`, the
// prefix will be stripped and the ID will be returned.
func (api *API) ParamID() (string, bool) {
	id := api.Param("id")
	if strings.HasPrefix(id, PrefixInternalID) {
		return id[len(PrefixInternalID):], true
	}
	return id, false
}

// ContextMiddleware initializes the context with
// auth information and a database connection.
func ContextMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		// Add authorization to context
		token, ok := c.Get("user").(*jwt.Token)
		if !ok {
			return errors.New("JWT missing")
		}
		claims, ok := token.Claims.(*auth.Claims)
		if !ok {
			return errors.New("invalid token claims")
		}
		scopes := strings.Split(claims.Scope, " ")
		ref := claims.RegisteredClaims.Subject

		// Acquire connection
		conn, err := store.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()

		// Create API context
		ac := &API{
			Scopes: scopes,
			Conn:   conn,
			Ref:    ref,

			Context: c,
		}

		return next(ac)
	}
}

// Init sets up a group with authentication
// for a restful management interface.
func Init(e *echo.Echo) error {
	// Register routes
	log.Info().Str("path", "/api/v1").Msg("initializing http api v1")
	v1 := e.Group("/api/v1")

	// API Auth and Context Middlewares
	v1.Use(ErrorHandler)
	v1.Use(auth.NewJWTAuthMiddleware())
	v1.Use(ContextMiddleware)

	// Status
	v1.GET("", Endpoint(apiStatusShow))
	v1.GET("/status", Endpoint(apiStatusShow))

	// API resources
	ResourceFrontends.Mount(v1, "/frontends")
	ResourceBackends.Mount(v1, "/backends")
	ResourceMeetings.Mount(v1, "/meetings")
	ResourceCommands.Mount(v1, "/commands")
	ResourceRecordingsImport.Mount(v1, "/recordings-import")
	ResourceAgentRPC.Mount(v1, "/agent/rpc")
	ResourceAgentBackend.Mount(v1, "/agent/backend")
	ResourceAgentHeartbeat.Mount(v1, "/agent/heartbeat")
	ResourceCtlMigrate.Mount(v1, "/ctrl/migrate")

	// Protected Recordings
	protected := e.Group("/api/v1/protected")
	protected.Use(ErrorHandler)
	protected.GET("/recordings/:token", apiProtectedRecordingsShow)
	protected.GET("/recordings/auth", apiProtectedRecordingsAuth)

	// Backend Callbacks
	publicV1 := e.Group("/api/v1")
	publicV1.Use(ErrorHandler)

	publicV1.POST("/recordings/ready/:token", apiOnRecordingReady)

	return nil
}

// StatusResponse returns information about the
// API implementation and the current user.
type StatusResponse struct {
	Version    string         `json:"version" doc:"The current b3scale server version."`
	Build      string         `json:"build" doc:"Build identifier of the server."`
	API        string         `json:"api" doc:"The API version." example:"v1"`
	AccountRef string         `json:"account_ref" doc:"The currently authenticated subject."`
	IsAdmin    bool           `json:"is_admin" doc:"True if the subject has admin privileges."`
	Database   *schema.Status `json:"database" doc:"Status of the database" api:"SchemaStatus"`
}

// apiStatusShow will respond with the api version and b3scale
// version.
func apiStatusShow(ctx context.Context, api *API) error {
	dbURL := config.EnvOpt(config.EnvDbURL, config.EnvDbURLDefault)
	m := schema.NewManager(dbURL)
	status := &StatusResponse{
		Version:    config.Version,
		Build:      config.Build,
		API:        "v1",
		AccountRef: api.Ref,
		IsAdmin:    api.HasScope(auth.ScopeAdmin),
		Database:   m.Status(ctx),
	}
	return api.JSON(http.StatusOK, status)
}

// RequireScope creates a middleware to ensure the presence of
// at least one required scope.
func RequireScope(scopes ...string) ResourceMiddleware {
	return func(next ResourceHandler) ResourceHandler {
		return func(ctx context.Context, api *API) error {
			hasScope := false
			for _, sc := range scopes {
				if api.HasScope(sc) {
					hasScope = true
					break
				}
			}
			if !hasScope {
				return auth.ErrScopeRequired(scopes...)
			}
			return next(ctx, api) // We are good to go.
		}
	}
}
