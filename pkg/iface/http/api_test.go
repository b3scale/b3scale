package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

// MakeTestContext creates a new testing context
func MakeTestContext(req *http.Request) (*APIContext, *httptest.ResponseRecorder) {
	if req == nil {
		req = httptest.NewRequest("GET", "http:///", nil)
	}
	rec := httptest.NewRecorder()
	e := echo.New()
	ctx := e.NewContext(req, rec)
	return &APIContext{ctx}, rec
}

// AuthorizeTestContext authorizes the context
func AuthorizeTestContext(ctx echo.Context, sub string, scopes []string) *APIContext {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &APIAuthClaims{
		StandardClaims: jwt.StandardClaims{
			Subject: sub,
		},
		Scope: strings.Join(scopes, " "),
	})
	ctx.Set("user", token)
	return &APIContext{ctx}
}

func TestAPIContextHasScope(t *testing.T) {
	ctx, _ := MakeTestContext(nil)
	ctx = AuthorizeTestContext(ctx, "user23", []string{"foo", "b3scale"})

	if !ctx.HasScope("b3scale") {
		t.Error("b3scale should be a scope in an authorized context")
	}
	if ctx.HasScope("aaaaaaaa") {
		t.Error("unexpected scope in context")
	}
}
func TestAPIContextSubject(t *testing.T) {
	ctx, _ := MakeTestContext(nil)
	ctx = AuthorizeTestContext(ctx, "user42", []string{})
	if ctx.Subject() != "user42" {
		t.Error("unexpected subject:", ctx.Subject())
	}
}

func TestAPIStatus(t *testing.T) {
	ctx, rec := MakeTestContext(nil)
	ctx = AuthorizeTestContext(ctx, "user42", []string{ScopeAdmin})
	a := &API{}
	if err := a.Status(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
}
