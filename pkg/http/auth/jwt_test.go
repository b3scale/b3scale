package auth

import (
	"testing"
	"time"
)

func TestClaims(t *testing.T) {
	claims := NewClaims("sub").WithScopes(ScopeRecordings)
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
	claims := NewClaims("frontend42").
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
	token, _ := NewClaims("frontend42").
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
	token, _ := NewClaims("frontend42").
		WithScopes(ScopeRecordings).
		Sign("secret42")

	_, err := ParseAPIToken(token, "not_secret42")
	if err == nil {
		t.Fatal("parse api token should fail with wrong secret")
	}
}

func TestClaimsHasScope(t *testing.T) {
	claims := NewClaims("frontend42").
		WithScopes(
			ScopeRecordings,
			ScopeAdmin)

	if !claims.HasScope(ScopeRecordings) {
		t.Error("expected scope to be present")
	}
	if !claims.HasScope(ScopeAdmin) {
		t.Error("expected scope to be present")
	}
	if claims.HasScope("fnord") {
		t.Error("expected scope to be not present")
	}
}

func TestParseWithAudience(t *testing.T) {
	token, _ := NewClaims("frontend42").
		WithAudience("resource42").
		WithScopes(ScopeRecordings).
		Sign("secret42")

	claims, err := ParseAPIToken(token, "secret42")
	if err != nil {
		t.Fatal(err)
	}
	if claims.RegisteredClaims.Subject != "frontend42" {
		t.Error("Expected subject to be frontend42")
	}

	if claims.Subject() != "frontend42" {
		t.Error("Expected subject to be frontend42")
	}
	if !claims.HasScope(ScopeRecordings) {
		t.Error("expected scope to be not present")
	}
	if claims.Audience() != "resource42" {
		t.Error("expected audience to be resource42")
	}
	t.Log(claims)
}

func TestParseUnverifiedRaw(t *testing.T) {
	token, _ := NewClaims("frontend42").
		WithAudience("resource42").
		WithScopes(ScopeRecordings).
		Sign("secret42")

	claims, err := ParseUnverifiedRawToken(token)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(claims)
}

func TestSignRawToken(t *testing.T) {
	token, _ := NewClaims("frontend42").
		Sign("secret42")

	claims, err := ParseUnverifiedRawToken(token)
	if err != nil {
		t.Fatal(err)
	}

	newToken, err := SignRawToken(claims, "secret42")
	if err != nil {
		t.Fatal(err)
	}

	_, err = ParseAPIToken(newToken, "secret42")
	if err != nil {
		t.Fatal(err)
	}

	_, err = ParseAPIToken("foo", "secret23")
	if err == nil {
		t.Fatal("expected error")
	}
}
