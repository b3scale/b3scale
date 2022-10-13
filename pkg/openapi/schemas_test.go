package openapi

import (
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/store"
)

func TestPropertiesFromObject(t *testing.T) {
	props := PropertiesFromObject(store.FrontendState{})
	// The test will fail if this panics.
	t.Log(props)

	props = PropertiesFromObject(store.FrontendSettings{})
	t.Log(props)

	props = PropertiesFromObject(bbb.Frontend{})
	t.Log(props)
}
