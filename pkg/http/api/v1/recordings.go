package v1

import (
	"github.com/labstack/echo/v4"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// RecordingsImportMeta will accept the contents of a
// metadata.xml from a published recording and will import
// the state.
// ! requires: `node`
func RecordingsImportMeta(c echo.Context) error {
	apiCtx := c.(*APIContext)
	ctx := ctx.Ctx()

	// Begin TX
	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := tx.Commit(ctx); err != nil {
		return err
	}
}
