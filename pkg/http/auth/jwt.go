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
	ScopeCallback   = "b3scale:callback"
)

// ErrScopeRequired will be returned when a scope is missing
// from the response.
func ErrScopeRequired(scopes ...string) *echo.HTTPError {
	scope := strings.Join(scopes, ", ")
	return echo.NewHTTPError(
		http.StatusForbidden,
		scope+" scope required")
}

// Claims extends the JWT standard claims
// with a well-known `scope` claim.
type Claims struct {
	Scope string `json:"scope"`
	jwt.RegisteredClaims
}

// NewClaims creates a new set of claims for use with
// the API. This includes an ID and the subject.
//
// The subject can be any string, however it will be used
// as an identifier for the 'user' making the request.
func NewClaims(sub string) *Claims {
	id := GenerateNonce(24)
	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:       id,
			Subject:  sub,
			IssuedAt: jwt.NewNumericDate(time.Now().UTC()),
		},
	}
}

// Scopes returns the list of scopes.
func (c *Claims) Scopes() []string {
	return strings.Split(c.Scope, " ")
}

// HasScope checks if the token has a scope
func (c *Claims) HasScope(scope string) bool {
	for _, s := range c.Scopes() {
		if s == scope {
			return true
		}
	}
	return false
}

// WithScopes adds a list of scopes to the claims.
func (c *Claims) WithScopes(scopes ...string) *Claims {
	c.Scope = strings.Join(scopes, " ")
	return c
}

// WithScopesCSV adds a list of scopes to the claims,
// separated by a delimiter.
func (c *Claims) WithScopesCSV(scopes string) *Claims {
	tokens := strings.Split(scopes, ",")
	trimmed := make([]string, 0, len(tokens))
	for _, t := range tokens {
		trimmed = append(trimmed, strings.TrimSpace(t))
	}

	return c.WithScopes(trimmed...)
}

// WithLifetime adds a lifetime to the claims.
func (c *Claims) WithLifetime(ttl time.Duration) *Claims {
	expiresAt := c.RegisteredClaims.IssuedAt.Add(ttl)
	c.RegisteredClaims.ExpiresAt = jwt.NewNumericDate(expiresAt)
	return c
}

// WithAudience adds an audience to the claims.
func (c *Claims) WithAudience(aud string) *Claims {
	c.RegisteredClaims.Audience = jwt.ClaimStrings{aud}
	return c
}

// Sign will create a new JWT from the claims.
func (c *Claims) Sign(secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, c)
	return token.SignedString([]byte(secret))
}

// ParseAPIToken validates and parses a JWT token.
func ParseAPIToken(data string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(data, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
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
			return &Claims{}
		},
	}
	return echojwt.WithConfig(cfg)
}
