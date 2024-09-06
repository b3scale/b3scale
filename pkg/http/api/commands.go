package api

import (
	"context"
	"net/http"

	"github.com/b3scale/b3scale/pkg/cluster"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/b3scale/b3scale/pkg/store"
)

// ResourceCommands bundles read and create operations
// for manipulating the command queue.
var ResourceCommands = &Resource{
	List: RequireScope(
		auth.ScopeAdmin,
	)(apiCommandList),

	Show: RequireScope(
		auth.ScopeAdmin,
	)(apiCommandShow),

	Create: RequireScope(
		auth.ScopeAdmin,
	)(apiCommandCreate),
}

// ErrCommandNotAllowed is a validation error
var ErrCommandNotAllowed = store.ValidationError{
	"action": []string{"this action is not allowed"},
}

// validateCommand checks if the command is ok
func validateCommand(cmd *store.Command) error {
	// Only allow DeleteBackend for now
	if cmd.Action != cluster.CmdEndAllMeetings {
		return ErrCommandNotAllowed
	}
	return nil
}

// apiCommandList returns the command queue
func apiCommandList(ctx context.Context, api *API) error {
	// Begin TX
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	commands, err := store.GetCommands(ctx, tx, store.Q())
	if err != nil {
		return err
	}
	return api.JSON(http.StatusOK, commands)
}

// apiCommandShow returns a single command by ID
func apiCommandShow(ctx context.Context, api *API) error {
	// Begin TX
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Fetch command
	id := api.Param("id")
	qry := store.Q().Where("id = ?", id)
	commands, err := store.GetCommand(ctx, tx, qry)
	if err != nil {
		return err
	}

	return api.JSON(http.StatusOK, commands)
}

// apiCommandCreate adds a new well known command to the queue
func apiCommandCreate(ctx context.Context, api *API) error {
	// Begin TX
	tx, err := api.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Parse command and insert into queue
	cmd := &store.Command{}
	if err := api.Bind(cmd); err != nil {
		return err
	}
	if err := validateCommand(cmd); err != nil {
		return err
	}
	if err := store.QueueCommand(ctx, tx, cmd); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	// Ok
	return api.JSON(http.StatusAccepted, cmd)
}
