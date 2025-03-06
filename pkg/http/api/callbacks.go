package api

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/http/callbacks"
	"github.com/b3scale/b3scale/pkg/middlewares/requests"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/labstack/echo/v4"
)

// Update callback query parameters:
// Some callbacks invoked by the bbb node will pass query parameters
// (e.g. recordingmarks=true|false to the callback URL.
func updateCallbackQuery(c echo.Context, callbackURL string) (string, error) {
	cbURL, err := url.Parse(callbackURL)
	if err != nil {
		return "", err
	}
	q := cbURL.Query()

	// Update query parameters: This will replace
	// the original query parameter if present.
	params := c.QueryParams()
	for k, v := range params {
		q[k] = v
	}
	cbURL.RawQuery = q.Encode()

	return cbURL.String(), nil
}

// Rewrite meetingID in query parameters: unpack an
// internal frontendKey meetingID pair and replace the
// query param.
func rewriteMeetingID(key string, query url.Values) {
	mID := query.Get(key)
	if mID == "" {
		return
	}

	// Try to decode meetingID
	fkmID := requests.DecodeFrontendKeyMeetingID(mID)
	if fkmID == nil {
		return // nothing to do here
	}

	// Rewrite meetingID
	query.Set(key, fkmID.MeetingID)
}

// Rewrite query parameters: The meetingID parameter might be
// set by the bbb node and contains a rewritten meetingID.
//
// For the callback we need to decode and patch the parameter.
func rewriteQueryParams(callbackURL string) (string, error) {
	cbURL, err := url.Parse(callbackURL)
	if err != nil {
		return "", err
	}
	q := cbURL.Query()

	// Because of the way the BBB api is implemented and
	// "designed", let's just assume the worst here.
	rewriteMeetingID("meeting_id", q)
	rewriteMeetingID("meetingID", q)
	rewriteMeetingID("meetingId", q)
	rewriteMeetingID("meetingid", q)

	cbURL.RawQuery = q.Encode()
	callbackURL = cbURL.String()

	return callbackURL, nil
}

// OnProxyGet accepts a rewritten GET callback from a BBB node,
// unpacks the original callback from the token and invokes it.
//
// There is no request body or JWT to reissue and the callback
// is invoked directly.
func apiOnProxyGet(c echo.Context) error {
	secret := config.MustEnv(config.EnvJWTSecret)

	rawToken := c.Param("token")
	token, err := auth.ParseAPIToken(rawToken, secret)
	if err != nil {
		return err
	}

	if !token.HasScope(auth.ScopeCallback) {
		return auth.ErrScopeRequired(auth.ScopeCallback)
	}

	callbackURL := token.Audience()
	if callbackURL == "" {
		return fmt.Errorf("callback url is required in request token")
	}
	callbackURL, err = updateCallbackQuery(c, callbackURL)
	if err != nil {
		return err
	}
	callbackURL, err = rewriteQueryParams(callbackURL)
	if err != nil {
		return err
	}

	// Invoke callback. This will happen in the background.
	callbacks.Dispatch(callbacks.Get(callbackURL))

	return c.NoContent(http.StatusOK)
}

// OnProxyPost accepts a rewritten POST callback from a BBB node,
// unpacks the original callback from the token and invokes it.
//
// This is used by the meta_recording-ready-callback-url.
// MeetingEnded callbacks are invoked via GET. Because.
//
// The endpoint accepts a "token" which is a JWT encoding the
// original URL.
//
// The received content is a JWT, which needs to be decoded,
// and signed with the secret of the frontend.
func apiOnProxyPost(c echo.Context) error {
	ctx := c.Request().Context()
	secret := config.MustEnv(config.EnvJWTSecret)

	// Acquire connection and begin database transaction
	conn, err := store.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint

	// Get request token and get frontend and original
	// callback URL. The token must be signed.
	rawToken := c.Param("token")
	token, err := auth.ParseAPIToken(rawToken, secret)
	if err != nil {
		return err
	}

	if !token.HasScope(auth.ScopeCallback) {
		return auth.ErrScopeRequired(auth.ScopeCallback)
	}

	callbackURL := token.Audience()
	if callbackURL == "" {
		return fmt.Errorf("callback url is required in request token")
	}
	callbackURL, err = updateCallbackQuery(c, callbackURL)
	if err != nil {
		return err
	}
	callbackURL, err = rewriteQueryParams(callbackURL)
	if err != nil {
		return err
	}

	// Get frontend and frontend secret
	frontendID := token.Subject()
	frontend, err := store.GetFrontendState(
		ctx, tx, store.Q().Where(
			"frontends.id = ?", frontendID))
	if err != nil {
		return err
	}
	if frontend == nil {
		return fmt.Errorf("could not find frontend")
	}

	// Read request body form data. The token will be in the
	// `signed_parameters` field.
	req := &callbacks.SignedBody{}
	if err := c.Bind(req); err != nil {
		return err
	}
	if err := req.Validate(); err != nil {
		return err
	}

	// Decode JWT. We do not care about the signature,
	// the secret varies by backend. We reissue the token
	// with the frontend secret.
	//
	// This is ok, because the request is authenticated
	// with the token in the URL and the callback POST
	// from the backend will include this token.
	cbClaims, err := auth.ParseUnverifiedRawToken(req.SignedParameters)
	if err != nil {
		return err
	}

	// The payload must contain a `meeting_id` parameters, which
	// needs to be rewritten.
	cbMeetingID, ok := cbClaims["meeting_id"].(string)
	if !ok {
		return fmt.Errorf("meeting_id not found in callback payload")
	}
	fkmID := requests.DecodeFrontendKeyMeetingID(cbMeetingID)
	if fkmID == nil {
		return fmt.Errorf("could not decode meetingID")
	}
	cbClaims["meeting_id"] = fkmID.MeetingID

	feSecret := frontend.Frontend.Secret
	cbToken, err := auth.SignRawToken(cbClaims, feSecret)
	if err != nil {
		return err
	}

	// Invoke callback. This will happen in the background.
	callbacks.Dispatch(callbacks.Post(
		callbackURL,
		&callbacks.SignedBody{
			SignedParameters: cbToken,
		}))

	return c.NoContent(http.StatusOK)
}
