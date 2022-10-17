package openapi

import (
	"encoding/json"
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/store"
)

func jsonProps(p Properties) string {
	j, _ := json.MarshalIndent(p, "", "  ")
	return string(j)
}

func TestPropertiesFromStruct(t *testing.T) {
	// The test will fail if this panics.
	props := PropertiesFrom(store.FrontendState{})
	t.Log(jsonProps(props))

	props = PropertiesFrom(store.FrontendSettings{})
	t.Log(jsonProps(props))

	props = PropertiesFrom(bbb.Frontend{})
	t.Log(jsonProps(props))

	props = PropertiesFrom(store.DefaultPresentationSettings{})
	t.Log(jsonProps(props))
}

func TestPropsFromStructMeetingstate(t *testing.T) {
	props := PropertiesFrom(store.MeetingState{})
	t.Log(jsonProps(props))
}

func TestPropsFromStructMeeting(t *testing.T) {
	props := PropertiesFrom(bbb.Meeting{})
	t.Log(jsonProps(props))
}

func TestPropsFromStructCommand(t *testing.T) {
	props := PropertiesFrom(store.Command{})
	t.Log(jsonProps(props))
}
