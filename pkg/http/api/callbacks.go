package api

import (
	"fmt"
	"net/http"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/http/callbacks"
	"github.com/b3scale/b3scale/pkg/middlewares/requests"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/labstack/echo/v4"
)

// OnRecordingReady provides an endpoint, for the backend to
// call, in case the frontend uses `meta_bbb-recording-ready-url`
// to signal that the recording is ready.
//
// The meeting create API request will be modified, to point to
// this endpoint.
//
// The endpoints accepts a "token" which is a JWT encoding the
// original bbb-recording-ready-url, which is then called.
//
// The received content is a JWT, which needs to be decoded,
// and signed with the secret of the frontend.
func apiOnRecordingReady(c echo.Context) error {
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
	defer tx.Rollback(ctx)

	// Get request token and get frontend and original
	// callback URL. The token must be signed.
	rawToken := c.Param("token")
	token, err := auth.ParseAPIToken(rawToken, secret)
	if err != nil {
		return err
	}

	frontendID := token.Subject()
	callbackURL := token.Audience()

	if !token.HasScope(auth.ScopeCallback) {
		return auth.ErrScopeRequired(auth.ScopeCallback)
	}

	if callbackURL == "" {
		return fmt.Errorf("callback url is required in request token")
	}

	// Get frontend and frontend secret
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
	req := &callbacks.OnRecordingReady{}
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

	// The payload contains a `meeting_id` parameters, which
	// needs to be rewritten.
	cbMeetingID, ok := cbClaims["meeting_id"].(string)
	if !ok {
		return fmt.Errorf("meeting_id not found in JWT payload")
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
	callbacks.Dispatch(callbacks.NewRequest(
		callbackURL,
		&callbacks.OnRecordingReady{
			SignedParameters: cbToken,
		}))

	return c.NoContent(http.StatusOK)
}

// OnApiMeetingEnd handles the internal rewritten
// callback endpoint for meeting end events.
//
// The backend will invoke it with additional
// query parameters: `recordingmarks`. The query
// parameters must be appended to the original URL.
func apiOnMeetingEnd(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}
