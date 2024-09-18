package requests

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/store"
)

func TestRewriteMetaCallbackURLs(t *testing.T) {
	// Configure Environment
	os.Setenv(config.EnvAPIURL, "https://b3s.example.com")
	os.Setenv(config.EnvJWTSecret, "secret42")

	frontend := cluster.NewFrontend(&store.FrontendState{
		ID: "frontend42",
	})

	ctx := cluster.ContextWithFrontend(context.Background(), frontend)
	req := &bbb.Request{
		Params: bbb.Params{
			bbb.MetaParamMeetingEndCallbackURL: "originalURL1",
			bbb.MetaParamRecordingReadyURL:     "originalURL2",
			bbb.ParamMeetingEndedURL:           "originalURL3",
		},
	}

	rewriteCallbacks(ctx, req)

	t.Log(req.Params)

	// Check params:
	newURL, ok := req.Params[bbb.MetaParamRecordingReadyURL]
	if !ok {
		t.Error("expected recording ready url")
	}
	t.Log(newURL)

	if !strings.HasPrefix(
		newURL,
		"https://b3s.example.com/api/v1/callbacks/proxy/") {
		t.Error("unexpected url:", newURL)
	}
	t.Log(newURL)

	newURL, _ = req.Params[bbb.MetaParamMeetingEndCallbackURL]
	if !strings.HasPrefix(
		newURL,
		"https://b3s.example.com/api/v1/callbacks/proxy/") {
		t.Error("unexpected url:", newURL)
	}
	t.Log(newURL)

	newURL, _ = req.Params[bbb.ParamMeetingEndedURL]
	if !strings.HasPrefix(
		newURL,
		"https://b3s.example.com/api/v1/callbacks/proxy/") {
		t.Error("unexpected url:", newURL)
	}
	t.Log(newURL)
}
