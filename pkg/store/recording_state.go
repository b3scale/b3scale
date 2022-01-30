package store

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// RecordingState holds a recording and its relation to a
// meeting.
type RecordingState struct {
	RecordID string

	Recording *bbb.Recording

	MeetingID         string
	InternalMeetingID string

	BackendID string

	CreatedAt time.Time
	UpdatedAt time.Time
	SyncedAt  time.Time
}

// InitRecordingState initializes the state with
// default values where required
func InitRecordingState(init *RecordingState) *RecordingState {
	return init
}

// GetRecordingStates retrieves a list of recordings
// filtered by criteria in the select builder
func GetRecordingStates(
	ctx context.Context,
	tx pgx.Tx,
	q sq.SelectBuilder,
) ([]*RecordingState, error) {
	qry, params, _ := q.Columns(
		"record_id",
		"meeting_id",
		"internal_meeting_id",
		"backend_id",
		"state",
	).From("recordings").ToSql()

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
			&state.BackendID,
			&state.Recording,
		)
		recordings = append(recordings, state)
	}
	return recordings, nil
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
			backend_id,
			state,
			updated_at,
			synced_at,
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT ON CONSTRAINT recordings_pkey DO UPDATE
		  SET meeting_id          = EXCLUDED.meeting_id,
		      internal_meeting_id = EXCLUDED.internal_meeting_id,
			  backend_id          = EXCLUDED.backend_id,
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
		s.BackendID,
		s.Recording,
		s.UpdatedAt,
		s.SyncedAt,
	)
	return err
}

// Delete will remove a recording from the database.
// This cascades to associated text tracks.
func (s *RecordingState) Delete(ctx context.Context, tx pgx.Tx) error {
	qry = `
		DELETE FROM recordings WHERE record_id = $1
	`

	_, err := tx.Exec(ctx, qry, s.RecordID)
	return err
}
