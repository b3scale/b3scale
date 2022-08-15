package v1

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// Until the echo middleware is updated, we have to use the
	// old repo of the jwt module.
	// "github.com/golang-jwt/jwt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"

	"github.com/b3scale/b3scale/pkg/store"
)

func init() {
	if err := store.ConnectTest(); err != nil {
		panic(err)
	}
}

// MakeTestContext creates a new testing context
func MakeTestContext(req *http.Request) (*APIContext, *httptest.ResponseRecorder) {
	reqCtx := context.Background()
	// Acquire connection
	conn, err := store.Acquire(reqCtx)
	if err != nil {
		panic(err)
	}

	reqCtx = store.ContextWithConnection(reqCtx, conn)

	// Make request if not present
	if req == nil {
		req = httptest.NewRequest("GET", "http:///", nil)
	}
	req = req.WithContext(reqCtx)

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

func TestAPIContextAccountRef(t *testing.T) {
	ctx, _ := MakeTestContext(nil)
	ctx = AuthorizeTestContext(ctx, "user42", []string{})
	if ctx.AccountRef() != "user42" {
		t.Error("unexpected account ref:", ctx.AccountRef())
	}
}

func TestAPIStatus(t *testing.T) {
	ctx, rec := MakeTestContext(nil)
	ctx = AuthorizeTestContext(ctx, "user42", []string{ScopeUser})
	if err := Status(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
	body, _ := ioutil.ReadAll(res.Body)
	t.Log(string(body))
}

func ClearState() error {
	ctx, _ := MakeTestContext(nil)
	defer ctx.Release()

	reqCtx := ctx.Ctx()
	tx, err := store.ConnectionFromContext(reqCtx).Begin(reqCtx)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(reqCtx, "DELETE FROM commands"); err != nil {
		return err
	}
	if _, err := tx.Exec(reqCtx, "DELETE FROM meetings"); err != nil {
		return err
	}
	if _, err := tx.Exec(reqCtx, "DELETE FROM backends"); err != nil {
		return err
	}
	if _, err := tx.Exec(reqCtx, "DELETE FROM frontends"); err != nil {
		return err
	}
	if err := tx.Commit(reqCtx); err != nil {
		return err
	}
	return nil
}
