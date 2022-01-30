package store

import (
	"context"
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// Tests for recording states

func TestInsertAndUpdateRecording(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	b := backendStateFactory()
	if err := b.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	m, err := meetingStateFactory(ctx, tx, nil)
	if err != nil {
		t.Fatal(err)
	}

	state := &RecordingState{
		RecordID:          "record42",
		MeetingID:         m.ID,
		InternalMeetingID: m.InternalID,
		BackendID:         b.ID,
		Recording: &bbb.Recording{
			RecordID:          "record42",
			MeetingID:         m.ID,
			InternalMeetingID: m.InternalID,
			Name:              "recording42",
		},
	}

	if err := state.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	// now update
	state.Recording.Name = "newname"
	if err := state.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	// And fetch from DB
	res, err := GetRecordingStates(ctx, tx, Q().Where(
		"record_id = ?", "record42"))
	if err != nil {
		t.Fatal(err)
	}

	if res[0].Recording.Name != "newname" {
		t.Error("unexpected name:", res[0].Recording.Name)
	}
}
