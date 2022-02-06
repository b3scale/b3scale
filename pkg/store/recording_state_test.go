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

func TestSetGetTextTracks(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	b := backendStateFactory()
	if err := b.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	state := &RecordingState{
		RecordID:          "record42",
		MeetingID:         "meeting23",
		InternalMeetingID: "meeting23INT",
		BackendID:         b.ID,
		Recording: &bbb.Recording{
			RecordID:          "record42",
			MeetingID:         "meeting23",
			InternalMeetingID: "meeting23INT",
			Name:              "recording42",
		},
	}
	if err := state.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	// Blank state should not have recordings
	tracks, err := GetRecordingTextTracks(ctx, tx, "record42")
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 0 {
		t.Error("expected 0 tracks")
	}

	if err := state.SetTextTracks(ctx, tx, []*bbb.TextTrack{
		{Href: "https://foo.bar"},
		{Href: "https://test123"},
	}); err != nil {
		t.Fatal(err)
	}

	tracks, err = GetRecordingTextTracks(ctx, tx, "record42")
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 2 {
		t.Error("expected 2 tracks")
	}
	if tracks[0].Href != "https://foo.bar" {
		t.Error("unexpected text track", tracks[0])
	}
}

func TestRecordingSetFrontend(t *testing.T) {
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

	frontend := frontendStateFactory()
	if err := frontend.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	if err := state.SetFrontend(ctx, tx, frontend); err != nil {
		t.Fatal(err)
	}

	res, err := GetRecordingStates(ctx, tx, Q().Where(
		"frontend_id = ?", frontend.ID))
	if err != nil {
		t.Fatal(err)
	}

	if *res[0].FrontendID != frontend.ID {
		t.Error("unexpected name:", res[0])
	}
}
