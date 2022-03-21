package v1

import (
	// Until the echo middleware is updated, we have to use the
	// old repo of the jwt module.
	// "github.com/golang-jwt/jwt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4/middleware"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

// Scopes
const (
	ScopeUser  = "b3scale"
	ScopeAdmin = "b3scale:admin"
	ScopeNode  = "b3scale:node"
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
	claims := &APIAuthClaims{
		Scope: ScopeAdmin,
		StandardClaims: jwt.StandardClaims{
			Subject: sub,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	return token.SignedString(secret)
}
