package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/b3scale/b3scale/pkg/config"
)

// Scopes is a list of scopes
type Scopes []string

// Scopes
const (
	ScopeUser       = "b3scale"
	ScopeAdmin      = "b3scale:admin"
	ScopeNode       = "b3scale:node"
	ScopeRecordings = "b3scale:recordings"
)

// ErrScopeRequired will be returned when a scope is missing
// from the response.
func ErrScopeRequired(scopes ...string) *echo.HTTPError {
	scope := strings.Join(scopes, ", ")
	return echo.NewHTTPError(
		http.StatusForbidden,
		scope+" scope required")
}

// AuthClaims extends the JWT standard claims
// with a well-known `scope` claim.
type AuthClaims struct {
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
	cfg.Claims = &AuthClaims{}
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
	id := GenerateNonce(24)
	claims := &AuthClaims{
		Scope: scope,
		StandardClaims: jwt.StandardClaims{
			Subject: sub,
			Id:      id,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	return token.SignedString(secret)
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
				return ErrScopeRequired(scopes...)
			}
			return next(ctx, api) // We are good to go.
		}
	}
}
