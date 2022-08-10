package api

import (
	"context"
	"fmt"
	"io/ioutil"
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
	body, err := ioutil.ReadAll(api.Request().Body)
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
	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	state := store.StateFromRecording(rec)

	// Check if recording exists, to prevent overriding
	// metadatachanges from the user.
	present, err := state.Exists(ctx, tx)
	if err != nil {
		return err
	}
	if present {
		return api.JSON(http.StatusOK, rec)
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
