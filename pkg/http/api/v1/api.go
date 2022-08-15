package v1

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

	// Until the echo middleware is updated, we have to use the
	// old repo of the jwt module.
	// "github.com/golang-jwt/jwt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/store"
)

// Errors
var (
	// ErrMissingJWTSecret will be returned if a JWT secret
	// could not be found in the environment.
	ErrMissingJWTSecret = errors.New("missing JWT secret")
)

// APIContext extends the context and provides methods
// for handling the current user.
type APIContext struct {
	echo.Context
}

// Release will free any acquired resources of this context
func (ctx *APIContext) Release() {
	conn := store.ConnectionFromContext(ctx.Ctx())
	if conn != nil {
		conn.Release()
	}
}

// HasScope checks if the authentication scope claim
// contains a scope by name.
// The scope claim is a space separated list of scopes
// according to RFC8693, Section 4.2, (OAuth 2).
func (ctx *APIContext) HasScope(s string) (found bool) {
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(*APIAuthClaims)
	scopes := strings.Split(claims.Scope, " ")
	for _, sc := range scopes {
		if sc == s {
			return true
		}
	}
	return false
}

// AccountRef retrievs the subject from the JWT
// as the account reference.
func (ctx *APIContext) AccountRef() string {
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(*APIAuthClaims)
	return claims.StandardClaims.Subject
}

// FilterAccountRef when the b3scale:admin scope is
// present, this function retrieves the value
// of the query param `ref`. The value will be nil
// in absence of the parameter.
//
// When the admin scope is not present, the requesting
// subject will be used.
func (ctx *APIContext) FilterAccountRef() *string {
	if ctx.HasScope(ScopeAdmin) {
		ref := ctx.Context.QueryParam("subject_ref")
		if ref == "" {
			return nil
		}
		return &ref
	}
	ref := ctx.AccountRef()
	return &ref
}

// Ctx is a shortcut to access the request context
func (ctx *APIContext) Ctx() context.Context {
	return ctx.Request().Context()
}

// Init sets up a group with authentication
// for a restful management interface.
func Init(e *echo.Echo) error {
	// Initialize JWT middleware config
	jwtConfig, err := NewAPIJWTConfig()
	if err != nil {
		return err
	}

	// Register routes
	log.Info().Str("path", "/api/v1").Msg("initializing http api v1")
	a := e.Group("/api/v1")

	// API Auth and Context Middlewares
	a.Use(middleware.JWTWithConfig(jwtConfig))
	a.Use(APIErrorHandler)
	a.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ac := &APIContext{c}

			// Check presence of required scopes
			if !ac.HasScope(ScopeUser) && !ac.HasScope(ScopeAdmin) {
				return ErrorInvalidCredentials(c)
			}

			req := c.Request()
			ctx := req.Context()

			// Acquire connection
			conn, err := store.Acquire(ctx)
			if err != nil {
				return err
			}
			defer conn.Release()

			ctx = store.ContextWithConnection(ctx, conn)
			req = req.WithContext(ctx)
			c.SetRequest(req)

			return next(ac)
		}
	})

	// Status
	a.GET("", Status)

	// Frontends
	a.GET("/frontends", FrontendsList)
	a.POST("/frontends", FrontendCreate)
	a.GET("/frontends/:id", FrontendRetrieve)
	a.DELETE("/frontends/:id", FrontendDestroy)
	a.PATCH("/frontends/:id", FrontendUpdate)

	// Backends
	a.GET("/backends", RequireAdminScope(BackendsList))
	a.POST("/backends", RequireAdminScope(BackendCreate))
	a.GET("/backends/:id", RequireAdminScope(BackendRetrieve))
	a.DELETE("/backends/:id", RequireAdminScope(BackendDestroy))
	a.PATCH("/backends/:id", RequireAdminScope(BackendUpdate))

	// Recordings
	a.POST("/recordings-import", RequireNodeScope(RecordingsImportMeta))

	// Meetings at backend. The backend is required because
	// the returned response set might be really big.
	// However, the backend might be specified either through
	// the backend ID or by host.
	a.GET("/meetings", RequireAdminScope(BackendMeetingsList))
	a.DELETE("/meetings", RequireAdminScope(BackendMeetingsEnd))

	return nil
}

// StatusResponse returns information about the
// API implementation and the current user.
type StatusResponse struct {
	Version    string `json:"version"`
	Build      string `json:"build"`
	API        string `json:"api"`
	AccountRef string `json:"account_ref"`
	IsAdmin    bool   `json:"is_admin"`
}

// Status will respond with the api version and b3scale
// version.
func Status(c echo.Context) error {
	ctx := c.(*APIContext)
	status := &StatusResponse{
		Version:    config.Version,
		Build:      config.Build,
		API:        "v1",
		AccountRef: ctx.AccountRef(),
		IsAdmin:    ctx.HasScope(ScopeAdmin),
	}
	return c.JSON(http.StatusOK, status)
}
