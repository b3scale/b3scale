package store

import (
	"testing"

	"github.com/google/uuid"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

func TestMeetingStateInsert(t *testing.T) {
	pool := connectTest(t)
	state := InitMeetingState(pool, &MeetingState{
		ID: uuid.New().String(),
		Meeting: &bbb.Meeting{
			MeetingID:   uuid.New().String(),
			MeetingName: "meeeeeeet",
		},
	})

	id, err := state.insert()
	if err != nil {
		t.Error(err)
	}
	t.Log("New meeting state id:", id)
}
