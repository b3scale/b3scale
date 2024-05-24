package auth

import (
	"testing"
	"time"
)

func TestAuthClaims(t *testing.T) {
	claims := NewAuthClaims("sub").WithScopes(ScopeRecordings)
	t.Log(claims)

	claims = claims.
		WithLifetime(time.Duration(1) * time.Hour).
		WithAudience("resource42")
	if claims.RegisteredClaims.ExpiresAt.IsZero() {
		t.Error("Expected ExpiresAt to be set")
	}
	t.Log(claims)
}

func TestSignAPIToken(t *testing.T) {
	claims := NewAuthClaims("frontend42").
		WithScopes(ScopeRecordings).
		WithLifetime(time.Duration(1) * time.Hour).
		WithAudience("resource42")
	token, err := claims.Sign("secret")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)
}

func TestParseAPIToken(t *testing.T) {
	secret := "secret42"
	token, _ := NewAuthClaims("frontend42").
		WithScopes(ScopeRecordings).
		WithLifetime(time.Duration(1) * time.Hour).
		WithAudience("resource42").
		Sign(secret)
	t.Log(token)

	claims, err := ParseAPIToken(token, secret)
	if err != nil {
		t.Fatal(err)
	}
	if claims.RegisteredClaims.Subject != "frontend42" {
		t.Error("Expected subject to be frontend42")
	}
	if claims.RegisteredClaims.Audience[0] != "resource42" {
		t.Error("Expected audience to be resource42")
	}
	if claims.Scope != ScopeRecordings {
		t.Error("Expected scope to be recordings")
	}
}

func TestParseAPITokenInvalid(t *testing.T) {
	token, _ := NewAuthClaims("frontend42").
		WithScopes(ScopeRecordings).
		Sign("secret42")

	_, err := ParseAPIToken(token, "not_secret42")
	if err == nil {
		t.Fatal("parse api token should fail with wrong secret")
	}
}
