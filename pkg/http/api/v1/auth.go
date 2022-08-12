package v1

import (
	// Until the echo middleware is updated, we have to use the
	// old repo of the jwt module.
	// "github.com/golang-jwt/jwt"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/b3scale/b3scale/pkg/config"
)

// Scopes
const (
	ScopeUser  = "b3scale"
	ScopeAdmin = "b3scale:admin"
	ScopeNode  = "b3scale:node"
)

// Errors
var (
	// ErrAdminScopeRequired will be returned if the token
	// has insuficient rights.
	ErrAdminScopeRequired = echo.NewHTTPError(
		http.StatusForbidden,
		ScopeAdmin+" scope required")

	// ErrNodeScopeRequired will be returned if the token
	// has insuficient rights.
	ErrNodeScopeRequired = echo.NewHTTPError(
		http.StatusForbidden,
		ScopeNode+" scope required")
)

// APIAuthClaims extends the JWT standard claims
// with a well-known `scope` claim.
type APIAuthClaims struct {
	Scope string `json:"scope"`
	jwt.StandardClaims
}

// NewAPIJWTConfig creates a new JWT middleware config.
// Parameters like shared secrets, public keys, etc..
// are retrieved from the environment.
func NewAPIJWTConfig() (middleware.JWTConfig, error) {
	secret := config.EnvOpt(config.EnvJWTSecret, "")
	if secret == "" {
		return middleware.JWTConfig{}, ErrMissingJWTSecret
	}

	cfg := middleware.DefaultJWTConfig
	cfg.SigningMethod = "HS384"
	cfg.Claims = &APIAuthClaims{}
	cfg.SigningKey = []byte(secret)

	return cfg, nil
}

// SignAdminAccessToken creates a new authorized
// JWT with an admin scope.
func SignAdminAccessToken(sub string, secret []byte) (string, error) {
	return SignAccessToken(sub, ScopeAdmin, secret)
}

// SignAccessToken creates an authorized JWT
func SignAccessToken(sub string, scope string, secret []byte) (string, error) {
	claims := &APIAuthClaims{
		Scope: scope,
		StandardClaims: jwt.StandardClaims{
			Subject: sub,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	return token.SignedString(secret)
}

// RequireAdminScope wraps a handler func and checks for
// the presence of the AdminScope before invoking the
// decorated function.
func RequireAdminScope(fn echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*APIContext)
		if !ctx.HasScope(ScopeAdmin) {
			return ErrAdminScopeRequired
		}
		return fn(c)
	}
}

// RequireNodeScope checks if the node scope is present
// and wraps a handler func.
func RequireNodeScope(fn echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.(*APIContext)
		if !ctx.HasScope(ScopeNode) {
			return ErrNodeScopeRequired
		}
		return fn(c)
	}
}
