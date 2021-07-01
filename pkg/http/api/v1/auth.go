package v1

import (
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4/middleware"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
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

// NewAPIJWTConfig creates a new JWT middleware config.
// Parameters like shared secrets, public keys, etc..
// are retrieved from the environment.
func NewAPIJWTConfig() (middleware.JWTConfig, error) {
	secret := config.EnvOpt(config.EnvJWTSecret, "")
	if secret == "" {
		return middleware.JWTConfig{}, ErrMissingJWTSecret
	}

	cfg := middleware.DefaultJWTConfig
	cfg.Claims = &APIAuthClaims{}
	cfg.SigningKey = []byte(secret)

	return cfg, nil
}

// SignAccessToken creates a new authorized JWT.
func SignAccessToken() string {
	return "meeeeh"
}
