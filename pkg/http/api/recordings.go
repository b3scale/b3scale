package api

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
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
		ScopeAdmin,
		ScopeNode,
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

// RecordingsGetProtected checks the issued request token,
// creates a new auth token and writes it as a cookie. Then the
// user is redirected to the recording URL.
var ResourceProtectedRecordings = &Resource{
	Show: apiProtectedRecordingsShow,
}

func apiProtectedRecordingsShow(
	ctx context.Context,
	api *API,
) error {
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get the request token as path parameter and
	// decode the JWT.
	token, _ := api.ParamID()
	log.Info().Str("token", token).Msg("protected recording token")

	return nil
}
