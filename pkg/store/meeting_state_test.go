package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

func meetingStateFactory(
	ctx context.Context,
	tx pgx.Tx,
	init *MeetingState,
) (*MeetingState, error) {
	// We start with a blank meeting
	if init == nil {
		init = &MeetingState{
			ID:         uuid.New().String(),
			InternalID: uuid.New().String(),
		}
	}
	// A meeting should not exist without a backend and
	// frontend.
	if init.frontend == nil {
		init.frontend = frontendStateFactory()
		if err := init.frontend.Save(ctx, tx); err != nil {
			return nil, err
		}
		init.FrontendID = &init.frontend.ID
	}
	if init.backend == nil {
		init.backend = backendStateFactory()
		if err := init.backend.Save(ctx, tx); err != nil {
			return nil, err
		}
		init.BackendID = &init.backend.ID
	}

	// Prepare meeting state
	if init.Meeting == nil {
		init.Meeting = &bbb.Meeting{
			MeetingName: "MyMeetingName-" + uuid.New().String(),
		}
	}
	init.Meeting.MeetingID = init.ID
	init.Meeting.InternalMeetingID = init.InternalID

	return InitMeetingState(init), nil
}

func TestGetMeetingStates(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	m1, err := meetingStateFactory(ctx, tx, &MeetingState{
		ID:         uuid.New().String(),
		InternalID: uuid.New().String(),
		Meeting: &bbb.Meeting{
			Running: true,
		}})
	t.Log(m1)
	if err != nil {
		t.Fatal(err)
	}
	err = m1.Save(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}

	// Get running meetings
	states, err := GetMeetingStates(ctx, tx, Q().
		Where("id = ?", m1.ID).
		Where("state->'Running' = ?", true))
	if err != nil {
		t.Error(err)
	}
	if len(states) != 1 {
		t.Error("expected meeting to be in result set")
	}
}

func TestMeetingStateSave(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	state, err := meetingStateFactory(ctx, tx, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("meeting state after factory:", state)
	t.Log(
		"backend:", state.backend,
		"backendID:", state.BackendID,
		"frontendID:", state.FrontendID)

	err = state.Save(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("New meeting state id:", state.ID)
}

func TestMeetingStateSaveUpdate(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	state, err := meetingStateFactory(ctx, tx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	state.Meeting = &bbb.Meeting{
		MeetingName:       "bar",
		InternalMeetingID: uuid.New().String(),
	}
	if err := state.Save(ctx, tx); err != nil {
		t.Error(err)
		return
	}
}

func TestMeetingStateQueryUpdate(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	state, err := meetingStateFactory(ctx, tx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	state, err = GetMeetingState(ctx, tx, Q().
		Where("id = ?", state.ID))
	if err != nil {
		t.Fatal(err)
	}

	state.Meeting = &bbb.Meeting{
		MeetingName:       "bar",
		InternalMeetingID: uuid.New().String(),
	}
	if err := state.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}
}

func TestMeetingStateIsStale(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	state, err := meetingStateFactory(ctx, tx, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("before Save()")
	if err := state.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}
	t.Log("after Save()")

	state.SyncedAt = time.Now().UTC()
	err = state.Save(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if state.IsStale(time.Minute) {
		t.Error("state should be fresh")
	}

	state.SyncedAt = time.Now().UTC().Add(-10 * time.Minute)
	err = state.Save(ctx, tx)
	if err != nil {
		t.Error(err)
	}

	if !state.IsStale(1 * time.Minute) {
		t.Error("state should be stale")
	}
}

func TestDeleteMeetingStateByInternalID(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	state, err := meetingStateFactory(ctx, tx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.Save(ctx, tx); err != nil {
		t.Error(err)
	}

	// Now delete the meeting state
	if err := DeleteMeetingStateByInternalID(ctx, tx, state.InternalID); err != nil {
		t.Error(err)
	}
}

func TestDeleteMeetingStateByID(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	state, err := meetingStateFactory(ctx, tx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := state.Save(ctx, tx); err != nil {
		t.Error(err)
	}

	// Now delete the meeting state
	if err := DeleteMeetingStateByID(ctx, tx, state.ID); err != nil {
		t.Error(err)
	}
}

func TestDeleteOrphanMeetings(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	m1, err := meetingStateFactory(ctx, tx, nil)
	if err != nil {
		t.Fatal(err)
	}
	backend := m1.backend // this is lost because of the refresh at save...
	t.Log(backend.ID)

	if err := m1.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}
	m2, err := meetingStateFactory(ctx, tx, &MeetingState{
		ID:         uuid.New().String(),
		InternalID: uuid.New().String(),
		backend:    backend,
		BackendID:  &backend.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := m2.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}
	m3, err := meetingStateFactory(ctx, tx, &MeetingState{
		ID:         uuid.New().String(),
		InternalID: uuid.New().String(),
		backend:    backend,
		BackendID:  &backend.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := m2.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	// Create an unrelated meeting at a different backend
	mUnrel, err := meetingStateFactory(ctx, tx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := mUnrel.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	// Delete meeting
	keep := []string{m1.InternalID, m3.InternalID}
	count, err := DeleteOrphanMeetings(ctx, tx, *m1.BackendID, keep)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("deleted", count, "orphans")
	if count != 1 {
		t.Error("expected 1 orphan")
	}

	// The unrelated meeting should still be present
	m, err := GetMeetingState(ctx, tx, Q().
		Where("meetings.id = ?", mUnrel.ID))
	if err != nil {
		t.Error(err)
	}
	if m == nil {
		t.Error("unrelated meeting was deleted")
	}

}

func TestMeetingStateUpsert(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	m0, err := meetingStateFactory(ctx, tx, nil) // unrelated
	if err != nil {
		t.Fatal(err)
	}
	if err := m0.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	m1, err := meetingStateFactory(ctx, tx, nil)
	if err != nil {
		t.Fatal(err)
	}

	id, err := m1.Upsert(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	if !m1.UpdatedAt.IsZero() {
		t.Error("unexpected updated_at:", m1.UpdatedAt)
	}
	if m1.Meeting.Running {
		t.Error("unexpected running state in meeting:", m1.Meeting)
	}

	now := time.Now().UTC()

	// Make update
	m1.Meeting.Running = true
	m1.UpdatedAt = now
	id, err = m1.Upsert(ctx, tx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(id)
	t.Log(m1.UpdatedAt)

	// Read state
	m2, err := GetMeetingState(ctx, tx, Q().Where("meetings.id = ?", id))
	if err != nil {
		t.Fatal(err)
	}
	if m2.Meeting.Running != true {
		t.Error("meeting should have an updated state")
	}

	// m0 should not be affected
	if err := m0.Refresh(ctx, tx); err != nil {
		t.Fatal(err)
	}
	if m0.Meeting.Running {
		t.Error("unexpected meeting running:", m0.Meeting)
	}
}

func TestMeetingStateUpdateFrontendMeetingMapping(t *testing.T) {
	ctx := context.Background()
	tx := beginTest(ctx, t)
	defer tx.Rollback(ctx)

	frontend := frontendStateFactory()
	if err := frontend.Save(ctx, tx); err != nil {
		t.Fatal(err)
	}

	m := &MeetingState{
		ID:         "meeting23421",
		FrontendID: &frontend.ID,
	}
	feID, ok, err := LookupFrontendIDByMeetingID(ctx, tx, m.ID)
	if err != nil {
		t.Fatal(err)
	}

	if ok {
		t.Error("did not expect meeting id associated with frontend")
	}

	if err := m.updateFrontendMeetingMapping(ctx, tx); err != nil {
		t.Error(err)
	}

	feID, ok, err = LookupFrontendIDByMeetingID(ctx, tx, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected found frontend id")
	}
	if feID != frontend.ID {
		t.Error("unexpected frontend id", feID)
	}

	// Should be idempotent and safe to call multiple times
	if err := m.updateFrontendMeetingMapping(ctx, tx); err != nil {
		t.Error(err)
	}
}
