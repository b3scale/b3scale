package requests

import (
	"context"
	"os"
	"testing"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/store"
)

func TestRewriteRecordingReadyURL(t *testing.T) {
	// Configure Environment
	os.Setenv(config.EnvAPIURL, "https://b3ds.example.com/api/")
	os.Setenv(config.EnvJWTSecret, "secret42")

	frontend := cluster.NewFrontend(&store.FrontendState{
		ID: "frontend42",
	})

	ctx := cluster.ContextWithFrontend(context.Background(), frontend)
	req := &bbb.Request{
		Params: bbb.Params{
			bbb.MetaParamRecordingReadyURL: "https://example.com/recording_ready",
		},
	}

	rewriteRecordingReadyURL(ctx, req)

	t.Log(req.Params)
}
