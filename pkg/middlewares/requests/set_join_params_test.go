package requests

import (
	"strings"
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/store"
)

func TestUpdateJoinParams(t *testing.T) {

	fe := cluster.NewFrontend(&store.FrontendState{
		Settings: store.FrontendSettings{
			JoinDefaultParams: bbb.Params{
				"default1":         "default-param-1",
				"default2":         "default-param-2",
				"default3":         "default-param-3",
				"disabledFeatures": "captions,chat",
			},
			JoinOverrideParams: bbb.Params{
				"override1": "false",
				"override2": "override-param",
				"default3":  "",
			},
		},
	})
	req := &bbb.Request{
		Params: bbb.Params{
			"param1":           "req-param-1",
			"default1":         "req-param-2",
			"override1":        "norway",
			"default3":         "fooo",
			"disabledFeatures": "testing",
		},
	}

	updateJoinParams(req, fe)

	if req.Params["param1"] != "req-param-1" {
		t.Error("param1 should not have been touched",
			req.Params["param1"])
	}

	if req.Params["default1"] != "req-param-2" {
		t.Error("a param-2 was provided, unexpected default")
	}

	if req.Params["override1"] != "false" {
		t.Error("override1 was not applied")
	}

	if req.Params["override2"] != "override-param" {
		t.Error("override2 was not set")
	}

	// Handle disabled features array
	f := strings.Split(req.Params["disabledFeatures"], ",")
	hasFeature := map[string]bool{
		"captions": false,
		"chat":     false,
		"testing":  false,
	}
	for _, name := range f {
		hasFeature[name] = true
	}
	if hasFeature["captions"] == false || hasFeature["chat"] == false {
		t.Error("expected default feature")
	}
	if hasFeature["testing"] == false {
		t.Error("expected addition to list")
	}

}
