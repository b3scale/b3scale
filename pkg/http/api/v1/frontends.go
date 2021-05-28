package v1

// FrontendsList will list all frontends known
// to the cluster or within the user scope.
func FrontendsList(c echo.Context) error {
	ctx := c.(*APIContext)
	ref := ctx.FilterSubjectRef()
	reqCtx := ctx.Ctx()

	q := store.Q()
	if ref != nil {
		q.Where("subject_ref = ?", *ref)
	}
	tx, err := store.ConnectionFromContext(ctx.Ctx()).Begin(reqCtx)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
	}
	defer tx.Rollback(reqCtx)
	frontends, err := store.GetFrontendStates(reqCtx, tx, q)

	c.JSON(http.StatusOK, frontends)

	return nil
}

// FrontendCreate will add a new frontend to the cluster.
func FrontendCreate(c echo.Context) error {
	return nil
}

// FrontendRetrieve will retrieve a single frontend
// identified by ID.
func FrontendRetrieve(c echo.Context) error {
	return nil
}

// FrontendDestroy will remove a frontend from the cluster.
// The frontend is identified by ID.
func FrontendDestroy(c echo.Context) error {
	return nil
}

// FrontendUpdate will update the frontend with values
// provided by the request. Only keys provided will
// be updated.
func FrontendUpdate(c echo.Context) error {
	return nil
}
