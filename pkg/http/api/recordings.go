package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// Cookies
const (
	CookieKeyProtected = "_b3scale_protected"
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

// ResourceProtectedRecordings is rest resource
// handling access to a protected recording.
var ResourceProtectedRecordings = &Resource{
	Show: apiProtectedRecordingsShow,
}

// Associate the temporary request token with
// the user session and redirect to the protected
// recording resource. The URL returned from the
// BBB operation getRecordings?... will point to this
// resource.
func apiProtectedRecordingsShow(
	ctx context.Context,
	api *API,
) error {
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	secret := config.MustEnv(config.EnvJWTSecret)
	playbackHost := config.MustEnv(config.EnvRecordingsPlaybackHost)
	playbackDomain := config.DomainOf(playbackHost)

	rawToken := api.Param("id")
	token, err := auth.ParseAPIToken(rawToken, secret)
	if err != nil {
		return err
	}

	// Get tenant ID from the auth token.
	frontendID := api.Ref
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
		return fmt.Errorf("could not find frontend by ID")
	}

	// Get the requested recording ID from the
	// request token's audience:
	recordingID := token.RegisteredClaims.Audience[0]
	if recordingID == "" {
		return fmt.Errorf("no recording ID in token")
	}
	recordingState, err := store.GetRecordingState(ctx, tx, store.Q().
		Where("recordings.id = ?", recordingID))
	if err != nil {
		return err
	}
	rec := recordingState.Recording
	rec.SetPlaybackHost(playbackHost)

	// Create access token and store it in the session
	accessToken, err := auth.NewAuthClaims(frontendID).
		WithScopes(auth.ScopeRecordings).
		WithLifetime(8 * time.Hour).
		Sign(secret)
	if err != nil {
		return err
	}
	log.Info().Str("accessToken", accessToken).Msg("created access token")

	// Set cookie for top-level domain
	api.SetCookie(&http.Cookie{
		Name:   CookieKeyProtected,
		Value:  accessToken,
		Path:   "/",
		Domain: playbackDomain,
	})

	// Redirect to recording URL

	return nil
}

// ResourceProtectedAuth handles the NGINX auth_request.
var ResourceProtectedAuth = &Resource{
	List: apiProtectedAuthenticate,
}

// Authenticate the request using the provided token.
func apiProtectedAuthenticate(
	ctx context.Context,
	api *API,
) error {
	// Get the original URL from the X-Original-URL header
	// and the access token from the _b3scale_protected cookie.
	originalURL := api.Request().Header.Get("X-Original-URL")
	log.Info().Str("originalURL", originalURL).Msg("original URL")
	cookie, err := api.Cookie(CookieKeyProtected)
	if err != nil {
		return echo.ErrForbidden
	}
	accessToken := cookie.Value
	log.Info().Str("accessToken", accessToken).Msg("access token")
	claims, err := auth.ParseAPIToken(
		accessToken, config.MustEnv(config.EnvJWTSecret))
	if err != nil {
		return echo.ErrForbidden
	}
	log.Info().Interface("claims", claims).Msg("claims")

	// TODO: Check if claim contains the required scopes

	// Use context from api to return a response
	// to the NGINX auth_request.

	return nil
}
