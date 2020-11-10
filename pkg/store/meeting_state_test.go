package store

import (
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

func meetingStateFactory(pool *pgxpool.Pool, init *MeetingState) (*MeetingState, error) {
	// We start with a blank meeting
	if init == nil {
		init = &MeetingState{
			ID: uuid.New().String(),
		}
	}
	// A meeting can not exist without a backend and
	// frontend.
	if init.Frontend == nil {
		init.Frontend = frontendStateFactory(pool)
		if err := init.Frontend.Save(); err != nil {
			return nil, err
		}
	}
	if init.Backend == nil {
		init.Backend = backendStateFactory(pool)
		if err := init.Backend.Save(); err != nil {
			return nil, err
		}
	}

	if init.Meeting == nil {
		init.Meeting = &bbb.Meeting{
			MeetingID:   uuid.New().String(),
			MeetingName: "MyMeetingName-" + uuid.New().String(),
		}
	}

	return InitMeetingState(pool, init), nil
}

func TestMeetingStateSave(t *testing.T) {
	pool := connectTest(t)

	state, err := meetingStateFactory(pool, nil)
	if err != nil {
		t.Error(err)
		return
	}
	err = state.Save()
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("New meeting state id:", state.ID)
}

func TestMeetingStateSaveUpdate(t *testing.T) {
	pool := connectTest(t)
	state, err := meetingStateFactory(pool, nil)
	if err != nil {
		t.Error(err)
		return
	}
	if err := state.Save(); err != nil {
		t.Error(err)
		return
	}

	state.Meeting = &bbb.Meeting{
		MeetingName: "bar",
	}
	if err := state.Save(); err != nil {
		t.Error(err)
		return
	}
}
