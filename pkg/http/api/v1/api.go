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

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// Errors
var (
	// ErrMissingJWTSecret will be returned if a JWT secret
	// could not be found in the environment.
	ErrMissingJWTSecret = errors.New("missing JWT secret")
)

// Scopes
const (
	ScopeUser  = "b3scale"
	ScopeAdmin = "b3scale:admin"
)

// APIAuthClaims extends the JWT standard claims
// with a well-known `scope` claim.
type APIAuthClaims struct {
	Scope string `json:"scope"`
	jwt.StandardClaims
}

// APIContext extends the context and provides methods
// for handling the current user.
type APIContext struct {
	echo.Context
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

// Subject retrievs the "current user" from the JWT
func (ctx *APIContext) Subject() string {
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(*APIAuthClaims)
	return claims.StandardClaims.Subject
}

// FilterSubjectRef when the b3scale:admin scope is
// present, this function retrieves the value
// of the query param `ref`. The value will be nil
// in absence of the parameter.
//
// When the admin scope is not present, the requesting
// subject will be used.
func (ctx *APIContext) FilterSubjectRef() *string {
	if ctx.HasScope(ScopeAdmin) {
		ref := ctx.Context.QueryParam("subject_ref")
		if ref == "" {
			return nil
		}
		return &ref
	}
	ref := ctx.Subject()
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
	a := e.Group("/api/v1")

	// API Auth and Context Middlewares
	a.Use(middleware.JWTWithConfig(jwtConfig))
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
	a.GET("", api.Status)

	// Frontends
	a.GET("/frontends", api.FrontendsList)
	a.POST("/frontends", api.FrontendCreate)
	a.GET("/frontends/:id", api.FrontendRetrieve)
	a.DELETE("/frontends/:id", api.FrontendDestroy)
	a.PATCH("/frontends/:id", api.FrontendUpdate)

	// Backends
	a.GET("/backends", api.BackendsList)
	a.POST("/backends", api.BackendCreate)
	a.GET("/backends/:id", api.BackendRetrieve)
	a.DELETE("/backends/:id", api.BackendDestroy)
	a.PATCH("/backends/:id", api.BackendUpdate)

	return nil
}

// NewAPIJWTConfig creates a new JWT middleware config.
// Parameters like shared secrets, public keys, etc..
// are retrieved from the environment.
func NewAPIJWTConfig() (middleware.JWTConfig, error) {
	secret := config.EnvOpt(config.EnvJWTSecret, "")
	if secret == "" {
		return middleware.JWTConfig{}, ErrMissingJWTSecret
	}

	cfg := middleware.JWTConfig{
		Claims:     &APIAuthClaims{},
		SigningKey: []byte(secret),
	}
	return cfg, nil
}

// Status will respond with the api version and b3scale
// version.
func Status(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"version": config.Version,
		"build":   config.Build,
		"api":     "v1",
	})
}

// FrontendsList will list all frontends known
// to the cluster or within the user scope.
func FrontendsList(c echo.Context) error {
	ctx := c.(*APIContext)
	ref := ctx.FilterSubjectRef()
	reqCtx := ctx.Ctx()

	q := store.Q()
	if ref != nil {
		q.Where("subject_ref = ?", *ref)
	}
	tx, err := store.ConnectionFromContext(ctx.Ctx()).Begin(reqCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
	}
	defer tx.Rollback(reqCtx)
	frontends, err := store.GetFrontendStates(reqCtx, tx, q)

	c.JSON(http.StatusOK, frontends)

	return nil
}

// FrontendCreate will add a new frontend to the cluster.
func FrontendCreate(c echo.Context) error {
	return nil
}

// FrontendRetrieve will retrieve a single frontend
// identified by ID.
func FrontendRetrieve(c echo.Context) error {
	return nil
}

// FrontendDestroy will remove a frontend from the cluster.
// The frontend is identified by ID.
func FrontendDestroy(c echo.Context) error {
	return nil
}

// FrontendUpdate will update the frontend with values
// provided by the request. Only keys provided will
// be updated.
func FrontendUpdate(c echo.Context) error {
	return nil
}
