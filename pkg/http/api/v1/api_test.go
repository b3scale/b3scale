package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// Until the echo middleware is updated, we have to use the
	// old repo of the jwt module.
	// "github.com/golang-jwt/jwt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"

	"github.com/b3scale/b3scale/pkg/store"
)

func init() {
	if err := store.ConnectTest(); err != nil {
		panic(err)
	}
}

// APITestResponseRecorder is a response recorder with
// convenience functions for testing
type APITestResponseRecorder struct {
	*httptest.ResponseRecorder
}

// AssertOK checks the http status code of the response
func (rec *APITestResponseRecorder) StatusOK() error {
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", res.StatusCode)
	}
	return nil
}

func (rec *APITestResponseRecorder) Body() string {
	res := rec.Result()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	return string(body)
}

func (rec *APITestResponseRecorder) JSON() map[string]interface{} {
	res := rec.Result()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal(body, &data); err != nil {
		panic(err)
	}
	return data
}

// MakeTestContext creates a new testing context
func MakeTestContext(req *http.Request) (*APIContext, *APITestResponseRecorder) {
	ctx := context.Background()

	// Acquire connection
	conn, err := store.Acquire(ctx)
	if err != nil {
		panic(err)
	}

	// Make request if not present
	if req == nil {
		req = httptest.NewRequest("GET", "http:///", nil)
	}
	req = req.WithContext(ctx)

	rec := &APITestResponseRecorder{httptest.NewRecorder()}
	e := echo.New()

	context := e.NewContext(req, rec)

	return &APIContext{
		Conn:    conn,
		Context: context,
	}, rec
}

type APITestRequest struct {
	body        []byte
	contentType string
	sub         string
	scopes      []string
	query       string
	keep        bool
}

func NewTestRequest() *APITestRequest {
	return &APITestRequest{
		contentType: "application/json", // default
	}
}

// Authorize adds scope and subject to request
func (req *APITestRequest) Authorize(
	sub string,
	scopes ...string,
) *APITestRequest {
	req.sub = sub
	req.scopes = scopes
	return req
}

func (req *APITestRequest) Query(q string) *APITestRequest {
	req.query = q
	return req
}

// KeepState will prevent invoking a state reset after
// the context was created
func (req *APITestRequest) KeepState() *APITestRequest {
	req.keep = true
	return req
}

// JSON adds a request body
func (req *APITestRequest) JSON(payload interface{}) *APITestRequest {
	body, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	req.body = body
	req.contentType = "application/json"
	return req
}

// Binary adds a binary request body
func (req *APITestRequest) Binary(blob []byte) *APITestRequest {
	req.body = blob
	req.contentType = "application/octet-stream"
	return req
}

// Context creates the APIContext for a test request
func (req *APITestRequest) Context() (*APIContext, *APITestResponseRecorder) {
	url := "http:///"
	if req.query != "" {
		url += "?" + req.query
	}
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}
	if req.body != nil {
		httpReq, err = http.NewRequest(
			http.MethodPost, "http:///", bytes.NewBuffer(req.body))
		if err != nil {
			panic(err)
		}
		httpReq.Header.Set("content-type", req.contentType)
	}

	api, rec := MakeTestContext(httpReq)

	if req.sub != "" {
		api.Authorize(req.sub, req.scopes)
	}

	if req.keep == false {
		if err := api.ClearState(); err != nil {
			panic(err)
		}
	}

	return api, rec
}

func (api *APIContext) Release() {
	if api.Conn != nil {
		api.Conn.Release()
	}
}

// Invoke the endpoint handler in the api context
func (api *APIContext) Handle(endpoint APIEndpointHandler) error {
	return endpoint(api.Ctx(), api)
}

// Authorize authorizes the context
func (api *APIContext) Authorize(
	sub string,
	scopes []string,
) *APIContext {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &APIAuthClaims{
		StandardClaims: jwt.StandardClaims{
			Subject: sub,
		},
		Scope: strings.Join(scopes, " "),
	})
	api.Set("user", token)

	// Add authorization to context
	api.Ref = sub
	api.Scopes = scopes

	return api
}

func (api *APIContext) ClearState() error {
	ctx := api.Ctx()
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM commands"); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM meetings"); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM backends"); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM frontends"); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func TestAPIContextHasScope(t *testing.T) {
	api, _ := MakeTestContext(nil)
	defer api.Release()

	api.Authorize("user23", []string{"foo", "b3scale"})

	if !api.HasScope("b3scale") {
		t.Error("b3scale should be a scope in an authorized context")
	}
	if api.HasScope("aaaaaaaa") {
		t.Error("unexpected scope in context")
	}
}

func TestAPIStatus(t *testing.T) {
	endpoint := APIEndpoint(apiStatusShow)

	api, rec := MakeTestContext(nil)
	defer api.Release()

	if err := endpoint(api); err != nil {
		t.Fatal(err)
	}

	t.Log(rec)
}
