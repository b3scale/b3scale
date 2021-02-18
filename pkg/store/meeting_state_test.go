package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

func meetingStateFactory(ctx context.Context, init *MeetingState) (*MeetingState, error) {
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
		if err := init.frontend.Save(ctx); err != nil {
			return nil, err
		}
		init.FrontendID = &init.frontend.ID
	}
	if init.backend == nil {
		init.backend = backendStateFactory()
		if err := init.backend.Save(ctx); err != nil {
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
	ctx, end := beginTest(t)
	defer end()

	m1, err := meetingStateFactory(ctx, &MeetingState{
		ID:         uuid.New().String(),
		InternalID: uuid.New().String(),
		Meeting: &bbb.Meeting{
			Running: true,
		}})
	t.Log(m1)
	if err != nil {
		t.Error(err)
		return
	}
	err = m1.Save(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	// Get running meetings
	states, err := GetMeetingStates(ctx, Q().
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
	ctx, end := beginTest(t)
	defer end()

	state, err := meetingStateFactory(ctx, nil)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("meeting state after factory:", state)
	t.Log(
		"backend:", state.backend,
		"backendID:", state.BackendID,
		"frontendID:", state.FrontendID)

	err = state.Save(ctx)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("New meeting state id:", state.ID)
}

func TestMeetingStateSaveUpdate(t *testing.T) {
	ctx, end := beginTest(t)
	defer end()
	state, err := meetingStateFactory(ctx, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if err := state.Save(ctx); err != nil {
		t.Error(err)
		return
	}

	state.Meeting = &bbb.Meeting{
		MeetingName:       "bar",
		InternalMeetingID: uuid.New().String(),
	}
	if err := state.Save(ctx); err != nil {
		t.Error(err)
		return
	}
}

func TestMeetingStateUpdateIfExists(t *testing.T) {
	ctx, end := beginTest(t)
	defer end()
	state, err := meetingStateFactory(ctx, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if err := state.Save(ctx); err != nil {
		t.Error(err)
		return
	}

	m := &bbb.Meeting{
		MeetingName:       "new-meeting-name",
		InternalMeetingID: state.InternalID,
		MeetingID:         state.ID,
	}

	c, err := UpdateMeetingStateIfExists(ctx, m)
	if err != nil {
		t.Error(err)
		return
	}

	if c != 1 {
		t.Error("there should have been 1 update")
	}

	state, err = GetMeetingState(ctx, Q().
		Where("id = ?", state.ID))
	if err != nil {
		t.Error(err)
		return
	}

	if state.Meeting.MeetingName != "new-meeting-name" {
		t.Error("unexpected meeting name:", state.Meeting.MeetingName)
	}

}

func TestMeetingStateQueryUpdate(t *testing.T) {
	ctx, end := beginTest(t)
	defer end()
	state, err := meetingStateFactory(ctx, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if err := state.Save(ctx); err != nil {
		t.Error(err)
		return
	}

	state, err = GetMeetingState(ctx, Q().
		Where("id = ?", state.ID))
	if err != nil {
		t.Error(err)
		return
	}

	state.Meeting = &bbb.Meeting{
		MeetingName:       "bar",
		InternalMeetingID: uuid.New().String(),
	}
	if err := state.Save(ctx); err != nil {
		t.Error(err)
		return
	}
}

func TestMeetingStateIsStale(t *testing.T) {
	ctx, end := beginTest(t)
	defer end()
	state, err := meetingStateFactory(ctx, nil)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("before SavE()")
	if err := state.Save(ctx); err != nil {
		t.Error(err)
		return
	}
	t.Log("after SavE()")

	state.SyncedAt = time.Now().UTC()
	err = state.Save(ctx)
	if err != nil {
		t.Error(err)
	}

	if state.IsStale() {
		t.Error("state should be fresh")
	}

	state.SyncedAt = time.Now().UTC().Add(-10 * time.Minute)
	err = state.Save(ctx)
	if err != nil {
		t.Error(err)
	}

	if !state.IsStale() {
		t.Error("state should be stale")
	}
}

func TestDeleteMeetingStateByInternalID(t *testing.T) {
	ctx, end := beginTest(t)
	defer end()
	state, err := meetingStateFactory(ctx, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if err := state.Save(ctx); err != nil {
		t.Error(err)
	}

	// Now delete the meeting state
	if err := DeleteMeetingStateByInternalID(ctx, state.InternalID); err != nil {
		t.Error(err)
	}
}

func TestDeleteMeetingStateByID(t *testing.T) {
	ctx, end := beginTest(t)
	defer end()
	state, err := meetingStateFactory(ctx, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if err := state.Save(ctx); err != nil {
		t.Error(err)
	}

	// Now delete the meeting state
	if err := DeleteMeetingStateByID(ctx, state.ID); err != nil {
		t.Error(err)
	}
}

func TestDeleteOrphanMeetings(t *testing.T) {
	ctx, end := beginTest(t)
	defer end()
	m1, err := meetingStateFactory(ctx, nil)
	if err != nil {
		t.Error(err)
		return
	}
	backend := m1.backend // this is lost because of the refresh at save...
	t.Log(backend.ID)

	if err := m1.Save(ctx); err != nil {
		t.Error(err)
		return
	}
	m2, err := meetingStateFactory(ctx, &MeetingState{
		ID:         uuid.New().String(),
		InternalID: uuid.New().String(),
		backend:    backend,
		BackendID:  &backend.ID,
	})
	if err != nil {
		t.Error(err)
		return
	}
	if err := m2.Save(ctx); err != nil {
		t.Error(err)
		return
	}
	m3, err := meetingStateFactory(ctx, &MeetingState{
		ID:         uuid.New().String(),
		InternalID: uuid.New().String(),
		backend:    backend,
		BackendID:  &backend.ID,
	})
	if err != nil {
		t.Error(err)
		return
	}
	if err := m2.Save(ctx); err != nil {
		t.Error(err)
		return
	}

	// Create an unrelated meeting at a different backend
	mUnrel, err := meetingStateFactory(ctx, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if err := mUnrel.Save(ctx); err != nil {
		t.Error(err)
		return
	}

	// Delete meeting
	keep := []string{m1.InternalID, m3.InternalID}
	count, err := DeleteOrphanMeetings(ctx, *m1.BackendID, keep)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("deleted", count, "orphans")
	if count != 1 {
		t.Error("expected 1 orphan")
	}

	// The unrelated meeting should still be present
	m, err := GetMeetingState(ctx, Q().
		Where("meetings.id = ?", mUnrel.ID))
	if err != nil {
		t.Error(err)
	}
	if m == nil {
		t.Error("unrelated meeting was deleted")
	}

}
