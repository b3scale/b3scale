package http

/*
B3Scale API v1

Administrative API for B3Scale. See /docs/rest_api.md for
details.
*/

import (
	"errors"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
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
	defer func() {
		if r := recover(); r != nil {
			found = false
		}
	}()
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

// UserID retrievs the current user ID from the JWT
func (ctx *APIContext) UserID() (uid string) {
	defer func() {
		if r := recover(); r != nil {
			uid = ""
		}
	}()
	user := ctx.Get("user").(*jwt.Token)
	claims := user.Claims.(*APIAuthClaims)
	return claims.StandardClaims.Subject
}

// API is the v1 implementation of the b3scale
// administrative API.
type API struct{}

// InitAPI sets up a group with authentication
// for a restful management interface.
func InitAPI(e *echo.Echo) error {
	// Initialize JWT middleware config
	jwtConfig, err := NewAPIJWTConfig()
	if err != nil {
		return err
	}

	// Initialize API
	api := &API{}

	// Register routes
	a := e.Group("/api/v1")

	// API Auth and Context Middlewares
	a.Use(middleware.JWTWithConfig(jwtConfig))
	a.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ac := &APIContext{c}

			// Check presence of required scopes
			if !ac.HasScope(ScopeUser) && !ac.HasScope(ScopeAdmin) {
				return api.ErrorInvalidCredentials(c)
			}

			return next(ac)
		}
	})

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

// FrontendsList will list all frontends known
// to the cluster or within the user scope.
func (a *API) FrontendsList(c echo.Context) error {
	return nil
}

// FrontendCreate will add a new frontend to the cluster.
func (a *API) FrontendCreate(c echo.Context) error {
	return nil
}

// FrontendRetrieve will retrieve a single frontend
// identified by ID.
func (a *API) FrontendRetrieve(c echo.Context) error {
	return nil
}

// FrontendDestroy will remove a frontend from the cluster.
// The frontend is identified by ID.
func (a *API) FrontendDestroy(c echo.Context) error {
	return nil
}

// FrontendUpdate will update the frontend with values
// provided by the request. Only keys provided will
// be updated.
func (a *API) FrontendUpdate(c echo.Context) error {
	return nil
}

// BackendsList will list all frontends known
// to the cluster or within the user scope.
func (a *API) BackendsList(c echo.Context) error {
	return nil
}

// BackendCreate will add a new frontend to the cluster.
func (a *API) BackendCreate(c echo.Context) error {
	return nil
}

// BackendRetrieve will retrieve a single frontend
// identified by ID.
func (a *API) BackendRetrieve(c echo.Context) error {
	return nil
}

// BackendDestroy will remove a frontend from the cluster.
// The frontend is identified by ID.
func (a *API) BackendDestroy(c echo.Context) error {
	return nil
}

// BackendUpdate will update the frontend with values
// provided by the request. Only keys provided will
// be updated.
func (a *API) BackendUpdate(c echo.Context) error {
	return nil
}

// Error responses

// ErrorInvalidCredentials will create an API response
// for an unauthorized request
func (a *API) ErrorInvalidCredentials(c echo.Context) error {
	return c.JSON(http.StatusForbidden, map[string]string{
		"error": "invalid_credentials",
		"message": "the credentials provided are lacking the " +
			"scope: b3scale or b3scale:admin",
	})
}
