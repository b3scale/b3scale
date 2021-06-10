package v1

import (
	"net/http"
	"testing"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

func createTestFrontend(ctx *APIContext) (*store.FrontendState, error) {
	cctx := ctx.Ctx()
	tx, err := store.ConnectionFromContext(cctx).Begin(cctx)
	if err != nil {
		return nil, err
	}

	f := store.InitFrontendState(&store.FrontendState{
		Frontend: &bbb.Frontend{
			Key:    "testkey",
			Secret: "testsecret",
		},
		Settings: store.FrontendSettings{
			RequiredTags: []string{"tag1"},
		},
	})

	if err := f.Save(cctx, tx); err != nil {
		return nil, err
	}

	return f, tx.Commit(cctx)
}

func clearFrontends() error {
	ctx, _ := MakeTestContext(nil)
	defer ctx.Release()

	reqCtx := ctx.Ctx()
	tx, err := store.ConnectionFromContext(reqCtx).Begin(reqCtx)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(reqCtx, "DELETE FROM frontends"); err != nil {
		return err
	}
	if err := tx.Commit(reqCtx); err != nil {
		return err
	}
	return nil
}

func TestFrontendsList(t *testing.T) {
	ctx, rec := MakeTestContext(nil)
	ctx = AuthorizeTestContext(ctx, "user42", []string{ScopeAdmin})
	if err := FrontendsList(ctx); err != nil {
		t.Fatal(err)
	}
	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", res.StatusCode)
	}
}
