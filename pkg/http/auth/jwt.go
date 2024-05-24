package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

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
	jwt.RegisteredClaims
}

func NewAuthClaims(sub string) *AuthClaims {
	id := GenerateNonce(24)
	return &AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:       id,
			Subject:  sub,
			IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		},
	}
}

// Scopes returns the list of scopes.
func (c *AuthClaims) Scopes() []string {
	return strings.Split(c.Scope, " ")
}

// WithScopes adds a list of scopes to the claims.
func (c *AuthClaims) WithScopes(scopes ...string) *AuthClaims {
	c.Scope = strings.Join(scopes, " ")
	return c
}

// WithScopesCSV adds a list of scopes to the claims,
// separated by a delimiter.
func (c *AuthClaims) WithScopesCSV(scopes string) *AuthClaims {
	tokens := strings.Split(scopes, ",")
	trimmed := make([]string, 0, len(tokens))
	for _, t := range tokens {
		trimmed = append(trimmed, strings.TrimSpace(t))
	}

	return c.WithScopes(trimmed...)
}

// WithLifetime adds a lifetime to the claims.
func (c *AuthClaims) WithLifetime(ttl time.Duration) *AuthClaims {
	expiresAt := c.RegisteredClaims.IssuedAt.Add(ttl)
	c.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(expiresAt)
	return c
}

// WithAudience adds an audience to the claims.
func (c *AuthClaims) WithAudience(aud string) *AuthClaims {
	c.RegisteredClaims.Audience = jwt.ClaimStrings{aud}
	return c
}

// Sign will create a new JWT from the claims.
func (c *AuthClaims) Sign(secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, c)
	return token.SignedString([]byte(secret))
}

// ParseAPIToken validates and parses a JWT token.
func ParseAPIToken(data string, secret string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(data, &AuthClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*AuthClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// NewJWTAuthMiddleware creates a new instance of the
// echojwt middleware.
// Parameters like shared secrets, public keys, etc..
// are retrieved from the environment.
func NewJWTAuthMiddleware() echo.MiddlewareFunc {
	secret := config.MustEnv(config.EnvJWTSecret)
	cfg := echojwt.Config{
		SigningKey:    []byte(secret),
		SigningMethod: "HS384",
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &AuthClaims{}
		},
	}
	return echojwt.WithConfig(cfg)
}
