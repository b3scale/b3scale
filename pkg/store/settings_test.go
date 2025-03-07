package store

import (
	"context"
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
)

func TestFrontendSettingsSave(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx) //nolint

	state := frontendStateFactory()

	// Params are all stringly typed.
	state.Settings.CreateDefaultParams = bbb.Params{
		"duration":         "42",
		"disabledFeatures": "chat,captions,virtualBackgrounds",
		"groups":           "[{id:'1',name:'GroupA',roster:['1235']}]",
	}
	state.Settings.Recordings = &RecordingsSettings{
		DefaultVisibility: bbb.RecordingVisibilityProtected,
	}

	if err := state.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	// Retrieve state
	state, err := GetFrontendState(ctx, tx, Q().Where("id = ?", state.ID))
	if err != nil {
		t.Fatal(err)
	}

	if state.Settings.CreateDefaultParams["duration"] != "42" {
		t.Error("unexpected settings:", state.Settings.CreateDefaultParams)
	}
	if state.Settings.Recordings.DefaultVisibility != bbb.RecordingVisibilityProtected {
		t.Error("unexpected settings:", state.Settings.Recordings)
	}
}
