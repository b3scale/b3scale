package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/http/callbacks"
	"github.com/b3scale/b3scale/pkg/middlewares/requests"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// Regular Expressions
var (
	// ReMatchRecordID will match the recordID in an URL with
	// the pattern /.../<hash>-<number>/..
	ReMatchRecordID = regexp.MustCompile(`\/([a-f0-9]+-\d+)`)
)

// Cookies
const (
	CookieKeyProtected = "_b3s_protected"
)

// Errors
var (
	ErrRequestBodyRequired = echo.NewHTTPError(
		http.StatusBadRequest,
		"the request must contain content of a metadata.xml")
)

// ResourceRecordingsImport is the recordings import api resource
var ResourceRecordingsImport = &Resource{
	Create: RequireScope(
		auth.ScopeAdmin,
		auth.ScopeNode,
	)(apiRecordingsImport),
}

// RecordingsImportMeta will accept the contents of a
// metadata.xml from a published recording and will import
// the state.
func apiRecordingsImport(
	ctx context.Context,
	api *API,
) error {
	// Parse request body, which should be the content of a
	// metadata.xml
	if api.Request().Body == nil { // Read
		return ErrRequestBodyRequired
	}
	body, err := io.ReadAll(api.Request().Body)
	if err != nil {
		return err
	}
	meta, err := bbb.UnmarshalRecordingMetadata(body)
	if err != nil {
		return err
	}
	rec := meta.ToRecording()

	// Create preview using the provided thumbnails
	storage, err := store.NewRecordingsStorageFromEnv()
	if err != nil {
		log.Error().Err(err).Msg("could not use recordings storage")
	} else {
		preview := storage.MakeRecordingPreview(rec.RecordID)
		// Use the same preview for all formats, for now...
		for _, f := range rec.Formats {
			f.Preview = preview
		}
	}

	// Save to store
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	state := store.NewStateFromRecording(rec)

	// Check if recording exists, if so merge it with the new
	// recording state from the import.
	current, err := store.GetRecordingStateByID(ctx, tx, state.RecordID)
	if err != nil {
		return err
	}
	if current != nil {
		state.Merge(current)
	}

	// Lookup frontendID for this recording
	frontendID, ok, err := store.LookupFrontendIDByMeetingID(
		ctx, tx, state.MeetingID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf(
			"could not find frontendID for meetingID: %s",
			state.MeetingID)
	}
	state.FrontendID = frontendID

	if err := state.Save(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return api.JSON(http.StatusOK, rec)
}

// Associate the temporary request token with
// the user session and redirect to the protected
// recording resource. The URL returned from the
// BBB operation getRecordings?... will point to this
// resource.
func apiProtectedRecordingsShow(c echo.Context) error {
	ctx := c.Request().Context()

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

	// Get configuration and request token
	secret := config.MustEnv(config.EnvJWTSecret)
	playbackHost := config.MustEnv(config.EnvRecordingsPlaybackHost)
	playbackDomain := config.DomainOf(playbackHost)

	rawToken := c.Param("token")
	token, err := auth.ParseAPIToken(rawToken, secret)
	if err != nil {
		log.Error().Err(err).Msg("invalid recording request token")
		return HTMLError(
			c,
			http.StatusForbidden,
			"You are not allowed to access this recording.",
			"The provided link is invalid or has expired.")
	}

	// Get tenant ID from the auth token.
	frontendID := token.RegisteredClaims.Subject
	if frontendID == "" {
		return echo.ErrForbidden
	}
	frontend, err := store.GetFrontendState(
		ctx, tx, store.Q().Where(
			"frontends.id = ?", frontendID))
	if err != nil {
		return err
	}
	if frontend == nil {
		return fmt.Errorf("could not find frontend")
	}

	// Get the requested recording ID from the
	// request token's audience:
	recordingRequest := token.RegisteredClaims.Audience[0]
	if recordingRequest == "" {
		return fmt.Errorf("no recording ID in token")
	}
	recordID, format := auth.MustDecodeResource(recordingRequest)

	recordingState, err := store.GetRecordingState(ctx, tx, store.Q().
		Where("recordings.record_id = ?", recordID))
	if err != nil {
		return err
	}
	if recordingState == nil {
		return echo.ErrNotFound
	}
	if recordingState.FrontendID != frontend.ID {
		return echo.ErrForbidden
	}

	rec := recordingState.Recording
	rec.SetPlaybackHost(playbackHost)

	recFormat := rec.GetFormat(format)

	// Create access token and store it in the session.
	// The default lifetime is 8 hours.
	tokenTTL := 8 * time.Hour
	accessToken, err := auth.NewClaims(frontendID).
		WithScopes(auth.ScopeRecordings).
		WithLifetime(tokenTTL).
		Sign(secret)
	if err != nil {
		return err
	}

	// Set cookie for top-level domain
	c.SetCookie(&http.Cookie{
		Name:   CookieKeyProtected,
		Value:  accessToken,
		Path:   "/",
		MaxAge: int(tokenTTL.Seconds()),
		Domain: playbackDomain,
	})

	// Redirect to recording URL
	return c.Redirect(http.StatusFound, recFormat.URL)
}

// Parse the recording ID from the resource path.
func parseRecordIDPath(path string) (string, bool) {
	matches := ReMatchRecordID.FindStringSubmatch(path)
	if len(matches) != 2 {
		return "", false
	}
	return matches[1], true
}

// Authenticate the request using the provided token.
func apiProtectedRecordingsAuth(c echo.Context) error {
	ctx := c.Request().Context()

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

	// Get the requested path from X-Resource-Path header
	// and the access token from the cookie.
	resourcePath := c.Request().Header.Get("X-Resource-Path")
	resourcePath = strings.TrimSuffix(resourcePath, "/")

	// Parse resourcePath to get the recording ID, luckily
	// it always ends in the recording ID.
	recordID, ok := parseRecordIDPath(resourcePath)
	if !ok {
		// This is not a video file request and can be served
		return c.NoContent(http.StatusOK)
	}

	// Get recording state
	recordingState, err := store.GetRecordingState(ctx, tx, store.Q().
		Where("recordings.record_id = ?", recordID))
	if err != nil {
		return err
	}
	if recordingState == nil {
		return echo.ErrNotFound
	}

	// Check if the recording is acutally protected
	isProtected, _ := recordingState.Recording.Metadata.GetBool(bbb.ParamProtect)
	if !isProtected {
		return c.NoContent(http.StatusOK) // Just go ahead!
	}

	// Get request cookie
	cookie, err := c.Cookie(CookieKeyProtected)
	if err != nil {
		return echo.ErrForbidden
	}
	accessToken := cookie.Value
	claims, err := auth.ParseAPIToken(
		accessToken, config.MustEnv(config.EnvJWTSecret))
	if err != nil {
		return echo.ErrForbidden
	}
	frontendID := claims.RegisteredClaims.Subject

	// Check if the subject of the token matches the frontend
	if recordingState.FrontendID != frontendID {
		return echo.ErrForbidden
	}

	return c.NoContent(http.StatusOK)
}

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
