package store

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"

	"github.com/b3scale/b3scale/pkg/bbb"
)

// RecordingState holds a recording and its relation to a
// meeting.
type RecordingState struct {
	RecordID string `json:"record_id" doc:"ID of the recording."`

	Recording *bbb.Recording `json:"recording" api:"RecordingData" doc:"The bbb recording data"`

	MeetingID         string `json:"meeting_id" doc:"The id of the related meeting."`
	InternalMeetingID string `json:"internal_meeting_id" doc:"The internal meeting id."`

	FrontendID string `json:"frontend_id" doc:"The id of the associated frontend."`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	SyncedAt  time.Time `json:"synced_at"`
}

// InitRecordingState initializes the state with
// default values where required
func InitRecordingState(init *RecordingState) *RecordingState {
	return init
}

// NewStateFromRecording will initialize a recording state
// with a recording.
func NewStateFromRecording(
	recording *bbb.Recording,
) *RecordingState {
	return InitRecordingState(&RecordingState{
		RecordID:          recording.RecordID,
		MeetingID:         recording.MeetingID,
		InternalMeetingID: recording.InternalMeetingID,

		Recording: recording,

		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		SyncedAt:  time.Now().UTC(),
	})
}

// GetRecordingStates retrieves a list of recordings
// filtered by criteria in the select builder
func GetRecordingStates(
	ctx context.Context,
	tx pgx.Tx,
	q sq.SelectBuilder,
) ([]*RecordingState, error) {
	qry, params, _ := q.Columns(
		"recordings.record_id",
		"recordings.meeting_id",
		"recordings.internal_meeting_id",
		"recordings.frontend_id",
		"recordings.state",
	).From("recordings").ToSql()

	log.Debug().Str("sql", qry).Msg("GetRecordingStates query")

	rows, err := tx.Query(ctx, qry, params...)
	if err != nil {
		return nil, err
	}

	// Load result
	cmd := rows.CommandTag()
	recordings := make([]*RecordingState, 0, cmd.RowsAffected())

	for rows.Next() {
		state := InitRecordingState(&RecordingState{})
		err := rows.Scan(
			&state.RecordID,
			&state.MeetingID,
			&state.InternalMeetingID,
			&state.FrontendID,
			&state.Recording,
		)
		if err != nil {
			return nil, err
		}
		recordings = append(recordings, state)
	}
	return recordings, nil
}

// GetRecordingState retrieves a single state from the
// database.
func GetRecordingState(
	ctx context.Context,
	tx pgx.Tx,
	q sq.SelectBuilder,
) (*RecordingState, error) {
	states, err := GetRecordingStates(ctx, tx, q)
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, nil // Should be an error
	}
	return states[0], nil
}

// QueryRecordingsByFrontendKey creates a query
// selecting recordings by frontend key. The query can
// be extended e.g. for filtering by recordingID.
func QueryRecordingsByFrontendKey(frontendKey string) sq.SelectBuilder {
	return Q().
		Join("frontends ON frontends.id = recordings.frontend_id").
		Where("recordings.frontend_id IS NOT NULL").
		Where("frontends.key = ?", frontendKey)
}

// GetRecordingStateByID retrievs a single state
// identified by recordID.
func GetRecordingStateByID(
	ctx context.Context,
	tx pgx.Tx,
	recordID string,
) (*RecordingState, error) {
	q := Q().Where("record_id = ?", recordID)
	return GetRecordingState(ctx, tx, q)
}

// Exists checks if the recording is already present in
// in the store.
func (s *RecordingState) Exists(ctx context.Context, tx pgx.Tx) (bool, error) {
	r, err := GetRecordingStateByID(ctx, tx, s.RecordID)
	if err != nil {
		return false, err
	}
	if r == nil {
		return false, nil
	}
	return true, nil
}

// Merge will combine the state and also fill in missing fields
func (s *RecordingState) Merge(other *RecordingState) error {
	err := s.Recording.Merge(other.Recording)
	if err != nil {
		return err
	}
	if other.MeetingID != "" {
		s.MeetingID = other.MeetingID
	}
	if other.InternalMeetingID != "" {
		s.InternalMeetingID = other.InternalMeetingID
	}
	if other.FrontendID != "" {
		s.FrontendID = other.FrontendID
	}
	if !other.CreatedAt.IsZero() {
		s.CreatedAt = other.CreatedAt
	}
	if !other.UpdatedAt.IsZero() {
		s.UpdatedAt = other.UpdatedAt
	}
	return nil
}

// Save the recording state
func (s *RecordingState) Save(
	ctx context.Context,
	tx pgx.Tx,
) error {
	qry := `
		INSERT INTO recordings (
			record_id,
			meeting_id,
			internal_meeting_id,
			frontend_id,
			state,
			updated_at,
			synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		  ON CONFLICT ON CONSTRAINT recordings_pkey DO UPDATE
		  SET meeting_id          = EXCLUDED.meeting_id,
		      internal_meeting_id = EXCLUDED.internal_meeting_id,
			  state               = EXCLUDED.state,
			  updated_at          = EXCLUDED.updated_at,
			  synced_at           = EXCLUDED.synced_at
	`
	s.UpdatedAt = time.Now().UTC()

	// Upsert recording
	_, err := tx.Exec(ctx, qry,
		s.RecordID,
		s.MeetingID,
		s.InternalMeetingID,
		s.FrontendID,
		s.Recording,
		s.UpdatedAt,
		s.SyncedAt,
	)
	return err
}

// SetFrontendID will set the frontend_id attribute of
// the recording state.
func (s *RecordingState) SetFrontendID(
	ctx context.Context,
	tx pgx.Tx,
	frontendID string,
) error {
	qry := `UPDATE recordings
		       SET frontend_id = $2,
			       updated_at  = $3
			 WHERE record_id   = $1
	`
	_, err := tx.Exec(ctx, qry, s.RecordID, frontendID, time.Now().UTC())
	return err
}

// SetTextTracks will persist associated text tracks
// without touching the rest. The recording has to be
// present in the database.
func (s *RecordingState) SetTextTracks(
	ctx context.Context,
	tx pgx.Tx,
	tracks []*bbb.TextTrack,
) error {
	return SetRecordingTextTracks(ctx, tx, s.RecordID, tracks)
}

// DeleteRecordingByID will delete a recording
// identified by its id.
func DeleteRecordingByID(ctx context.Context, tx pgx.Tx, recordID string) error {
	qry := `
		DELETE FROM recordings WHERE record_id = $1
	`
	_, err := tx.Exec(ctx, qry, recordID)
	return err
}

// Delete will remove a recording from the database.
// This cascades to associated text tracks.
func (s *RecordingState) Delete(ctx context.Context, tx pgx.Tx) error {
	return DeleteRecordingByID(ctx, tx, s.RecordID)
}

// DeleteFiles will remove the recording from the
// filesystem.
func (s *RecordingState) DeleteFiles() error {
	storage, err := NewRecordingsStorageFromEnv()
	if err != nil {
		return err
	}
	return storage.DeleteRecording(s)
}

// PublishFiles will move the recording from the unpublished
// directory to the published directory.
func (s *RecordingState) PublishFiles() error {
	storage, err := NewRecordingsStorageFromEnv()
	if err != nil {
		return err
	}
	return storage.PublishRecording(s)
}

// UnpublishFiles will move the recording from the published
// directory to the unpublished directory.
func (s *RecordingState) UnpublishFiles() error {
	storage, err := NewRecordingsStorageFromEnv()
	if err != nil {
		return err
	}
	return storage.UnpublishRecording(s)
}

// ImportFiles will move incoming files from the
// inbox to the published or unpublished folder.
func (s *RecordingState) ImportFiles() error {
	storage, err := NewRecordingsStorageFromEnv()
	if err != nil {
		return err
	}
	return storage.ImportRecording(s)
}

// GetRecordingTextTracks retrieves the text tracks from
// a recording.
func GetRecordingTextTracks(
	ctx context.Context,
	tx pgx.Tx,
	recordID string,
) ([]*bbb.TextTrack, error) {
	// TODO: maybe just forward this to the
	// backend.
	qry := `
		SELECT text_track_states
		  FROM recordings
		 WHERE record_id = $1
	`
	tracks := []*bbb.TextTrack{}
	err := tx.QueryRow(ctx, qry, recordID).Scan(&tracks)
	return tracks, err
}

// SetRecordingTextTracks updates the text tracks attribute
// of a recording.
func SetRecordingTextTracks(
	ctx context.Context,
	tx pgx.Tx,
	recordID string,
	tracks []*bbb.TextTrack,
) error {
	qry := `
    UPDATE recordings
	     SET text_track_states = $2,
	         updated_at        = $3
     WHERE record_id         = $1
	`
	_, err := tx.Exec(
		ctx, qry, recordID, tracks, time.Now().UTC())
	return err
}
