package v1

import (
	// Until the echo middleware is updated, we have to use the
	// old repo of the jwt module.
	// "github.com/golang-jwt/jwt"
	"context"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/b3scale/b3scale/pkg/config"
)

// Scopes is a list of scopes
type Scopes []string

// Scopes
const (
	ScopeUser  = "b3scale"
	ScopeAdmin = "b3scale:admin"
	ScopeNode  = "b3scale:node"
)

// ErrScopeRequired will be returned when a scope is missing
// from the response.
func ErrScopeRequired(scopes ...string) *echo.HTTPError {
	scope := strings.Join(scopes, ", ")
	return echo.NewHTTPError(
		http.StatusForbidden,
		scope+" scope required")
}

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

// RequireScope creates a middleware to ensure the presence of
// at least one required scope.
func RequireScope(scopes ...string) APIEndpointMiddleware {
	return func(next APIEndpointHandler) APIEndpointHandler {
		return func(ctx context.Context, api *APIContext) error {
			hasScope := false
			for _, sc := range scopes {
				if api.HasScope(sc) {
					hasScope = true
					break
				}
			}
			if !hasScope {
				return ErrScopeRequired(scopes...)
			}
			return next(ctx, api) // We are good to go.
		}
	}
}
