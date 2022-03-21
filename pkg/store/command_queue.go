package store

/*
 Commands are serialized operations to be executed
 by any b3scale instance.

 The store implements a command queue for processing
 commands.
*/

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// CommandHandler is a callback function for handling
// commands. The command was successful if no error was
// returned.
type CommandHandler func(context.Context, *Command) (interface{}, error)

// A Command is a representation of an operation
type Command struct {
	ID  string `json:"id"`
	Seq int    `json:"seq"`

	State string `json:"state"`

	Action string      `json:"action"`
	Params interface{} `json:"params"`
	Result interface{} `json:"result"`

	Deadline  time.Time  `json:"deadline"`
	StartedAt *time.Time `json:"started_at"`
	StoppedAt *time.Time `json:"stopped_at"`
	CreatedAt time.Time  `json:"created_at"`

	tx pgx.Tx
}

// FetchParams loads the parameters and decodes them
func (cmd *Command) FetchParams(
	ctx context.Context,
	req interface{},
) error {
	qry := `SELECT params FROM commands WHERE id = $1`
	return cmd.tx.QueryRow(ctx, qry, cmd.ID).Scan(req)
}

// The CommandQueue is connected to the database and
// provides methods for queuing and dequeuing commands.
type CommandQueue struct {
}

// NewCommandQueue initializes a new command queue
func NewCommandQueue() *CommandQueue {
	return &CommandQueue{}
}

// QueueCommand adds a new command to the queue
func QueueCommand(ctx context.Context, tx pgx.Tx, cmd *Command) error {
	// Our command will always expire. For now 2 minutes.
	deadline := time.Now().UTC().Add(120 * time.Second)
	// Marshal payload
	params, err := json.Marshal(cmd.Params)
	if err != nil {
		return err
	}

	// Add command to queue and notify instances
	qry := `
	  INSERT INTO commands (
	  	action,
		params,
		deadline
	  ) VALUES (
		$1, $2, $3
	  )
	  RETURNING id`
	var cmdID string
	err = tx.QueryRow(ctx, qry, cmd.Action, params, deadline).
		Scan(&cmdID)
	if err != nil {
		return err
	}

	// Update command
	cmd.ID = cmdID
	cmd.CreatedAt = time.Now().UTC()
	return nil
}

// Receive will await a command and will block
// until a command can be processed. If the handler
// responds with an error, the error will be returned.
func (q *CommandQueue) Receive(handler CommandHandler) error {
	for {
		// Start processing in the background, so we can take
		// care of the next incomming command.
		go func(q *CommandQueue) {
			err := q.process(handler)
			if err != nil {
				log.Error().Err(err).Msg("processing job failed")
			}
		}(q)

		// Limit polling frequency to 15 Hz
		time.Sleep(66 * time.Millisecond)
	}
}

// Run the handler, but recover if an error occured.
func safeExecHandler(
	ctx context.Context,
	cmd *Command,
	handler CommandHandler,
) (res interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var conn *pgxpool.Conn
	// Get a database connection for the handler and pass
	// it through the context.
	conn, err = Acquire(ctx)
	if err != nil {
		return nil, err
	}
	ctx = ContextWithConnection(ctx, conn)
	defer conn.Release()

	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	return handler(ctx, cmd)
}

// Process will dequeue a command and apply the
// handler function to it. If not command was dequeued 'false'
// will be returned.
func (q *CommandQueue) process(handler CommandHandler) error {
	// Begin transaction with a total timelimit
	// of X seconds for the entire command. The safeExecHandler
	// will instanciate a child context with a stricter timelimit
	// of Y < X seconds for the job to complete.
	startedAt := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Second)
	defer cancel()

	tx, err := begin(ctx) // Command TX
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// We dequeue and fetch a command within a transaction.
	// During handling the command will be locked.
	qry := `
		SELECT 
			id,
			seq,
			action,
			deadline,
			created_at
		  FROM commands
		 WHERE state = 'requested'
		 ORDER BY seq ASC
		 LIMIT 1
		   FOR UPDATE SKIP LOCKED`

	// Select command
	cmd := &Command{}
	err = tx.QueryRow(ctx, qry).Scan(
		&cmd.ID,
		&cmd.Seq,
		&cmd.Action,
		&cmd.Deadline,
		&cmd.CreatedAt)
	if err != nil && err == pgx.ErrNoRows {
		return nil // Ok. There was just nothing to do.
	} else if err != nil {
		return err
	}

	cmd.tx = tx

	// Check deadline
	state := "success"
	var result interface{}
	if cmd.Deadline.Before(time.Now().UTC()) {
		// Timeout
		state = "error"
		result = "timedout"
	} else {
		// Apply command handler
		result, err = safeExecHandler(ctx, cmd, handler)
		if err != nil {
			log.Error().
				Err(err).
				Int("seq", cmd.Seq).
				Str("action", cmd.Action).
				Msg("exec command handler error")
			state = "error"
			result = fmt.Sprintf("%s", err)
		}
	}

	// We are done
	stoppedAt := time.Now()
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	// Write result
	qry = `
		UPDATE commands
		   SET state      = $2,
		       result     = $3,
			   started_at = $4,
			   stopped_at = $5

		 WHERE id = $1`
	_, err = tx.Exec(ctx, qry, cmd.ID,
		state,
		data,
		startedAt,
		stoppedAt)
	if err != nil {
		return err
	}

	// End transaction
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

// NextDeadline calculates the deadline for a
// newly requested command
func NextDeadline(dt time.Duration) time.Time {
	return time.Now().UTC().Add(dt)
}

// CountCommandsWithState retrievs the number of commands
// in the queue with a given state. e.g. requested, error,
// etc.
func CountCommandsWithState(
	ctx context.Context,
	tx pgx.Tx,
	state string,
) (int, error) {
	count := 0
	qry := `
		SELECT COUNT(1)
		  FROM (
		  	SELECT 1 FROM commands
			 WHERE state = $1
		  ORDER BY seq ASC
			   FOR SHARE SKIP LOCKED) AS Q`
	err := tx.QueryRow(ctx, qry, state).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountCommandsRequested returns the number of
// unprocessed commands in the queue.
func CountCommandsRequested(ctx context.Context, tx pgx.Tx) (int, error) {
	return CountCommandsWithState(ctx, tx, "requested")
}

// CountCommandsSuccess returns the number of
// successfully processed commands in the queue.
func CountCommandsSuccess(ctx context.Context, tx pgx.Tx) (int, error) {
	return CountCommandsWithState(ctx, tx, "success")
}

// CountCommandsError returns the number of
// successfully processed commands in the queue.
func CountCommandsError(ctx context.Context, tx pgx.Tx) (int, error) {
	return CountCommandsWithState(ctx, tx, "error")
}
