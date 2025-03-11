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
	"github.com/b3scale/b3scale/pkg/middlewares/requests"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/jackc/pgx/v4"
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

// Internal: Apply meeting overrides. Set the frontend key
// and associated meeting ID.
//
// Sometimes we need to import a legacy recording for a new
// frontend. In this case, we need to rewrite the meetingID
// to the new frontend.
func recordingsImportApplyOverrides(
	ctx context.Context,
	api *API,
	tx pgx.Tx,
	rec *bbb.Recording,
) error {
	// Use override_originial_meeting_id to associate the
	// recording with a different meetingID.
	meetingIDOverride := api.QueryParam("override_original_meeting_id")
	if meetingIDOverride != "" {
		rec.MeetingID = meetingIDOverride
	}
	// Use override_frontend_key to associate the recording
	// with a differnt tenant.
	frontendKeyOverride := api.QueryParam("override_frontend_key")
	if frontendKeyOverride != "" {
		log.Info().
			Str("override_frontend_key", frontendKeyOverride).
			Str("override_original_meeting_id", rec.MeetingID).
			Msg("importing recording with override")

		state, err := store.GetFrontendStateByKey(ctx, tx, frontendKeyOverride)
		if err != nil {
			return err
		}
		if state == nil {
			msg := fmt.Sprintf(
				"override_frontend_key: A frontend with key '%s' could not be found.",
				frontendKeyOverride)
			return echo.NewHTTPError(
				http.StatusNotFound,
				msg)
		}
		feID := state.ID
		feKey := state.Frontend.Key

		// Rewrite meetingID
		meetingID := (&requests.FrontendKeyMeetingID{
			FrontendKey: feKey,
			MeetingID:   rec.MeetingID,
		}).EncodeToString()
		rec.MeetingID = meetingID

		// Update meetingID in recording and register meeting
		// if not present with the frontend.
		mapping := &store.MeetingState{
			FrontendID: &feID,
			ID:         meetingID,
		}
		if err := mapping.UpdateFrontendMeetingMapping(ctx, tx); err != nil {
			return err
		}
	}

	return nil
}

// Internal: Apply frontend settings. Set the visiblity to
// published / unpublished and update gl-listed metadata
// in recording.
func recordingsImportApplyFrontendSettings(
	ctx context.Context,
	tx pgx.Tx,
	rec *bbb.Recording,
	feID string,
) error {
	// Get frontend settings
	fe, err := store.GetFrontendStateByID(ctx, tx, feID)
	if err != nil {
		return err
	}
	if fe == nil {
		return fmt.Errorf("could not get frontend by ID: %s", feID)
	}

	// Get recording visibility overrides
	rs := fe.Settings.Recordings
	if rs == nil {
		return nil // nothing to do here
	}
	vo := rs.VisibilityOverride
	if vo != nil {
		rec.SetVisibility(*vo)
	}

	return nil
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
		return err
	}

	preview := storage.MakeRecordingPreview(rec)
	// Use the same preview for all formats, for now...
	for _, f := range rec.Formats {
		f.Preview = preview
	}

	// Start store transaction
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint

	// Apply overrides, settings and defaults
	if err := recordingsImportApplyOverrides(
		ctx, api, tx, rec,
	); err != nil {
		return err
	}

	// Lookup frontendID for this recording
	frontendID, ok, err := store.LookupFrontendIDByMeetingID(
		ctx, tx, rec.MeetingID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf(
			"could not find frontendID for meetingID: %s",
			rec.MeetingID)
	}

	// Apply frontend overrides
	if err := recordingsImportApplyFrontendSettings(
		ctx, tx, rec, frontendID,
	); err != nil {
		return err
	}

	state := store.NewStateFromRecording(rec)

	// Check if recording exists, if so merge it with the new
	// recording state from the import.
	current, err := store.GetRecordingStateByID(ctx, tx, state.RecordID)
	if err != nil {
		return err
	}
	if current != nil {
		if err := state.Merge(current); err != nil {
			return err
		}
	}

	state.FrontendID = frontendID

	// Persist
	if err := state.Save(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	// Import from inbox
	if err := state.ImportFiles(); err != nil {
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
	defer tx.Rollback(ctx) //nolint

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
	defer tx.Rollback(ctx) //nolint

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

// ResourceRecordings is a restful group for managing recordings
var ResourceRecordings = &Resource{
	List: RequireScope(
		auth.ScopeAdmin,
	)(apiRecordingsList),
	Show: RequireScope(
		auth.ScopeAdmin,
	)(apiRecordingsShow),
}

// API: Recordings list endpoint
func apiRecordingsList(
	ctx context.Context,
	api *API,
) error {
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint

	fe, err := FrontendFromQueryParams(ctx, api, tx)
	if err != nil {
		return err
	}

	// Get recordings for frontend
	res, err := store.GetRecordingStates(ctx, tx, store.Q().
		Where("recordings.frontend_id = ?", fe.ID))
	if err != nil {
		return err
	}

	return api.JSON(http.StatusOK, res)
}

// API: Read a single recording
func apiRecordingsShow(
	ctx context.Context,
	api *API,
) error {
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint

	id := api.Param("id") // RecordID

	// Get recording by ID
	rec, err := store.GetRecordingState(ctx, tx, store.Q().
		Where("recordings.record_id = ?", id))
	if err != nil {
		return err
	}

	if rec == nil {
		return echo.ErrNotFound
	}

	return api.JSON(http.StatusOK, rec)
}

var ResourceRecordingsVisibility = &Resource{
	Create: RequireScope(
		auth.ScopeAdmin,
		auth.ScopeNode,
	)(apiRecordingsVisibilityUpdate),
}

// RecordingVisibilityUpdate requests changing the visibility
// of a recording.
type RecordingVisibilityUpdate struct {
	RecordID   string                  `json:"record_id" doc:"The ID of the recording. (RecordID)"`
	Visibility bbb.RecordingVisibility `json:"visibility" doc:"The new visibilty."`
}

// API: Update the recording visiblity.
func apiRecordingsVisibilityUpdate(
	ctx context.Context,
	api *API,
) error {
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint

	update := &RecordingVisibilityUpdate{}
	if err := api.Bind(update); err != nil {
		return err
	}

	recID := update.RecordID
	if recID == "" {
		return fmt.Errorf("recordID may not be empty")
	}

	// Get recording for update
	rec, err := store.GetRecordingState(ctx, tx, store.Q().
		Where("recordings.record_id = ?", recID))
	if err != nil {
		return err
	}
	if rec == nil {
		return echo.ErrBadRequest
	}

	// Update visibility
	rec.Recording.SetVisibility(update.Visibility)

	if err := rec.Save(ctx, tx); err != nil {
		return err
	}
	// Release connection before fsops, to prevent
	// exhausting the pool.
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	// Update filesystem
	if rec.Recording.Published {
		if err := rec.PublishFiles(); err != nil {
			return err
		}
	} else {
		if err := rec.UnpublishFiles(); err != nil {
			return err
		}
	}

	return api.JSON(http.StatusOK, rec)
}
