package v1

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/store"
)

// Errors
var (
	ErrRequestBodyRequired = echo.NewHTTPError(
		http.StatusBadRequest,
		"the request must contain content of a metadata.xml")
)

// RecordingsImportMeta will accept the contents of a
// metadata.xml from a published recording and will import
// the state.
// ! requires: `node`
func RecordingsImportMeta(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse request body, which should be the content of a
	// metadata.xml
	if c.Request().Body == nil { // Read
		return ErrRequestBodyRequired
	}

	body, err := ioutil.ReadAll(c.Request().Body)
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
		return c.JSON(http.StatusOK, rec)
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

	return c.JSON(http.StatusOK, rec)
}
