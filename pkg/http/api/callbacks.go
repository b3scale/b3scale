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

	// Update query parameters
	params := c.QueryParams()
	for k, v := range params {
		for _, value := range v {
			q.Add(k, value)
		}
	}
	cbURL.RawQuery = q.Encode()

	return cbURL.String(), nil
}

// OnProxyCallback accepts a rewritten callback from a BBB node,
// unpacks the original callback from the token and invokes it.
//
// The `rewrite_meta_callback_urls` middleware updates the parameters
// of the request with a new callback URL.
// The endpoint accepts a "token" which is a JWT encoding the
// original URL.
//
// The received content is a JWT, which needs to be decoded,
// and signed with the secret of the frontend.
func apiOnProxyCallback(c echo.Context) error {
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
	req := &callbacks.Callback{}
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

	// The payload may contain a `meeting_id` parameters, which
	// needs to be rewritten.
	cbMeetingID, ok := cbClaims["meeting_id"].(string)
	if ok {
		fkmID := requests.DecodeFrontendKeyMeetingID(cbMeetingID)
		if fkmID == nil {
			return fmt.Errorf("could not decode meetingID")
		}
		cbClaims["meeting_id"] = fkmID.MeetingID
	}

	feSecret := frontend.Frontend.Secret
	cbToken, err := auth.SignRawToken(cbClaims, feSecret)
	if err != nil {
		return err
	}

	// Invoke callback. This will happen in the background.
	callbacks.Dispatch(callbacks.NewRequest(
		callbackURL,
		&callbacks.Callback{
			SignedParameters: cbToken,
		}))

	return c.NoContent(http.StatusOK)
}
