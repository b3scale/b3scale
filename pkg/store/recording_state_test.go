package store

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
)

// todo refactor into testdata
func readTestResponse(name string) []byte {
	filename := path.Join(
		"../../testdata/responses/",
		name)
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	return data
}

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
	frontend := frontendStateFactory()
	if err := frontend.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	state := &RecordingState{
		RecordID:          "record42",
		MeetingID:         m.ID,
		InternalMeetingID: m.InternalID,
		FrontendID:        frontend.ID,
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

	frontend := frontendStateFactory()
	if err := frontend.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	state := &RecordingState{
		RecordID:          "record42",
		MeetingID:         "meeting23",
		InternalMeetingID: "meeting23INT",
		FrontendID:        frontend.ID,
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

func TestRecordingSetFrontendID(t *testing.T) {
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

	frontend := frontendStateFactory()
	if err := frontend.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	state := &RecordingState{
		RecordID:          "record42",
		MeetingID:         m.ID,
		FrontendID:        frontend.ID,
		InternalMeetingID: m.InternalID,
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

	frontend = frontendStateFactory()
	if err := frontend.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	if err := state.SetFrontendID(ctx, tx, frontend.ID); err != nil {
		t.Fatal(err)
	}

	res, err := GetRecordingStates(ctx, tx, Q().Where(
		"frontend_id = ?", frontend.ID))
	if err != nil {
		t.Fatal(err)
	}

	if res[0].FrontendID != frontend.ID {
		t.Error("unexpected name:", res[0])
	}
}

func TestMerge(t *testing.T) {
	data := readTestResponse("../recordings/metadata.xml")
	meta, err := bbb.UnmarshalRecordingMetadata(data)
	if err != nil {
		t.Fatal(err)
	}
	rec := meta.ToRecording()
	state := NewStateFromRecording(rec)

	data = readTestResponse("../recordings/metadata.m4v.xml")
	meta, err = bbb.UnmarshalRecordingMetadata(data)
	if err != nil {
		t.Fatal(err)
	}
	rec = meta.ToRecording()
	state2 := NewStateFromRecording(rec)
	state2.Merge(state)

	if len(state2.Recording.Formats) != 2 {
		t.Error("expected 2 formats")
	}

	if state2.MeetingID == "" {
		t.Error("unexpected empty meeting id")
	}
	if state2.MeetingID != state.MeetingID {
		t.Error("unexpected meeting id:", state2.MeetingID, state.MeetingID)
	}
	if state2.InternalMeetingID == "" {
		t.Error("unexpected empty internal meeting id")
	}
	if state2.InternalMeetingID != state.InternalMeetingID {
		t.Error("unexpected internal meeting id")
	}
	if state2.FrontendID != state.FrontendID {
		t.Error("unexpected frontend id")
	}

	t.Log(state2)
}
