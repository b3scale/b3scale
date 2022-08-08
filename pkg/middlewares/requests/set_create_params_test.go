package requests

import (
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/store"
)

func TestUpdateCreateParams(t *testing.T) {
	fe := cluster.NewFrontend(&store.FrontendState{
		Settings: store.FrontendSettings{
			CreateDefaultParams: bbb.Params{
				"allowModsToEjectCameras": "true",
				"disabledFeatures":        "captions,polls,chat",
			},
			CreateOverrideParams: bbb.Params{
				"logo": "override-logo",
			},
		},
	})
	req := &bbb.Request{
		Params: bbb.Params{
			"userCameraCap":    "2",
			"logo":             "request-logo",
			"disabledFeatures": "captions,breakoutRooms,sharedNotes",
		},
	}

	updateCreateParams(req, fe)

	if req.Params["userCameraCap"] != "2" {
		t.Error("param should not have been touched",
			req.Params["userCameraCap"])
	}
	if req.Params["logo"] != "override-logo" {
		t.Error("param should have been overriden",
			req.Params["logo"])
	}
	if req.Params["disabledFeatures"] !=
		"captions,breakoutRooms,sharedNotes,polls,chat" {
		t.Error("disabledFeatures should have been amended",
			req.Params["disabledFeatures"])
	}
}

func TestUpdateCreateParamsClearFeatures(t *testing.T) {
	fe := cluster.NewFrontend(&store.FrontendState{
		Settings: store.FrontendSettings{
			CreateDefaultParams: bbb.Params{
				"disabledFeatures": "captions,polls,chat",
			},
			CreateOverrideParams: bbb.Params{
				"disabledFeatures": "",
			},
		},
	})
	req := &bbb.Request{
		Params: bbb.Params{
			"userCameraCap":    "2",
			"logo":             "request-logo",
			"disabledFeatures": "captions,breakoutRooms,sharedNotes",
		},
	}

	updateCreateParams(req, fe)

	if req.Params["disabledFeatures"] != "captions,polls,chat" {
		t.Error("unexpected disabled features", req.Params["disabledFeatures"])
	}
}
