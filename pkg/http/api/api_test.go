package api

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

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store"
)

func init() {
	ctx := context.Background()
	if err := store.ConnectTest(ctx); err != nil {
		fmt.Println("WARNING: can not connect to DB. tests will fail.")
	}
}

// ResponseRecorder is a response recorder with
// convenience functions for testing
type ResponseRecorder struct {
	*httptest.ResponseRecorder
}

// AssertOK checks the http status code of the response
func (rec *ResponseRecorder) StatusOK() error {
	res := rec.Result()
	code := res.StatusCode
	if code != http.StatusOK && code != http.StatusAccepted {
		return fmt.Errorf("unexpected status code: %v", res.StatusCode)
	}
	return nil
}

func (rec *ResponseRecorder) Body() string {
	res := rec.Result()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	return string(body)
}

func (rec *ResponseRecorder) JSON() map[string]interface{} {
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
func MakeTestContext(req *http.Request) (*API, *ResponseRecorder) {
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

	rec := &ResponseRecorder{httptest.NewRecorder()}
	e := echo.New()

	context := e.NewContext(req, rec)

	return &API{
		Conn:    conn,
		Context: context,
	}, rec
}

type TestRequest struct {
	body        []byte
	contentType string
	sub         string
	scopes      []string
	query       string
	keep        bool
}

func NewTestRequest() *TestRequest {
	return &TestRequest{
		contentType: "application/json", // default
	}
}

// Authorize adds scope and subject to request
func (req *TestRequest) Authorize(
	sub string,
	scopes ...string,
) *TestRequest {
	req.sub = sub
	req.scopes = scopes
	return req
}

func (req *TestRequest) Query(q string) *TestRequest {
	req.query = q
	return req
}

// KeepState will prevent invoking a state reset after
// the context was created
func (req *TestRequest) KeepState() *TestRequest {
	req.keep = true
	return req
}

// JSON adds a request body
func (req *TestRequest) JSON(payload interface{}) *TestRequest {
	body, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	req.body = body
	req.contentType = "application/json"
	return req
}

// Binary adds a binary request body
func (req *TestRequest) Binary(blob []byte) *TestRequest {
	req.body = blob
	req.contentType = "application/octet-stream"
	return req
}

// Context creates the Context for a test request
func (req *TestRequest) Context() (*API, *ResponseRecorder) {
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

func (api *API) Release() {
	if api.Conn != nil {
		api.Conn.Release()
	}
}

// Invoke the endpoint handler in the api context
func (api *API) Handle(endpoint ResourceHandler) error {
	return endpoint(api.Ctx(), api)
}

// Authorize authorizes the context
func (api *API) Authorize(
	sub string,
	scopes []string,
) *API {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &auth.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
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

func (api *API) ClearState() error {
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

func TestContextHasScope(t *testing.T) {
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

func TestStatus(t *testing.T) {
	endpoint := Endpoint(apiStatusShow)

	api, rec := MakeTestContext(nil)
	defer api.Release()

	if err := endpoint(api); err != nil {
		t.Fatal(err)
	}

	t.Log(rec)
}

func TestParamID(t *testing.T) {
	api, _ := MakeTestContext(nil)
	defer api.Release()

	api.SetParamNames("id")
	api.SetParamValues("foo-bar")

	id, internal := api.ParamID()
	if id != "foo-bar" {
		t.Error("unexpected id", id)
	}
	if internal == true {
		t.Error("should not have been an internal ID")
	}
}

func TestParamIDInternal(t *testing.T) {
	api, _ := MakeTestContext(nil)
	defer api.Release()

	api.SetParamNames("id")
	api.SetParamValues("internal:fnord")

	id, internal := api.ParamID()
	if id != "fnord" {
		t.Error("unexpected id", id)
	}
	if !internal {
		t.Error("should have been an internal ID")
	}
}
