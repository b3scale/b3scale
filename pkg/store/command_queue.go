package store

/*
 Commands are serialized operations to be executed
 by any b3scale instance.

 The store implements a command queue for processing
 commands.
*/

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const cmdQueue = "commands_queue"

// CommandHandler is a callback function for handling
// commands. The command was successful if no error was
// returned.
type CommandHandler func(*Command) (interface{}, error)

// A Command is a representation of an operation
type Command struct {
	ID  string
	Seq int

	State string

	Action string
	Params interface{}
	Result interface{}

	Deadline  time.Time
	StartedAt *time.Time
	StoppedAt *time.Time
	CreatedAt *time.Time
}

// The CommandQueue is connected to the database and
// provides methods for queuing and dequeuing commands.
type CommandQueue struct {
	conn *pgxpool.Conn
	pool *pgxpool.Pool
}

// NewCommandQueue initializes a new command queue
func NewCommandQueue(pool *pgxpool.Pool) *CommandQueue {
	// Subscribe to notifications
	return &CommandQueue{
		pool: pool,
	}
}

// Subscribe will let the queue listen for notifications
func (q *CommandQueue) Subscribe() error {
	ctx := context.Background()
	conn, err := q.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	// Subscribe to queue
	_, err = conn.Exec(ctx, "LISTEN "+cmdQueue)
	if err != nil {
		return err
	}
	q.conn = conn

	return nil
}

// Queue adds a new command to the queue
func (q *CommandQueue) Queue(cmd *Command) error {
	ctx := context.Background()
	// Our command will always expire. For now 2 minutes.
	deadline := time.Now().UTC().Add(120 * time.Second)
	// Add command to queue and notify instances
	qry := `
	  INSERT INTO commands (
	  	action,
		params,
		deadline
	  ) VALUES (
		$1, $2, $3
	  )`
	_, err := q.pool.Exec(ctx, qry, cmd.Action, cmd.Params, deadline)
	if err != nil {
		return err
	}

	// Notify subscribers
	_, err = q.pool.Exec(ctx, "NOTIFY "+cmdQueue)
	if err != nil {
		return err
	}

	return nil
}

// Receive will await a command and will block
// until a command can be processed. If the handler
// responds with an error, the error will be returned.
func (q *CommandQueue) Receive(handler CommandHandler) error {
	for {
		// We periodically check our queue. We only check instantly
		// if we got informed that there is a job waiting.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// Await command, after a timeout just try to dequeue
		if q.conn != nil {
			_, err := q.conn.Conn().WaitForNotification(ctx)
			if err != nil {
				netErr, ok := err.(net.Error)
				if ok {
					// In case we just ran into a timeout it
					// is perfectly fine to continue. Otherwise
					// we just forward the error
					if !netErr.Timeout() {
						return err
					}
				} else {
					return err
				}
			}
			dequeued, err := q.process(handler)
			if err != nil {
				return err
			}
			if dequeued {
				// We processed a command
				return nil
			}
		}
	}
}

// Process will dequeue a command and apply the
// handler function to it. If not command was dequeued 'false'
// will be returned.
func (q *CommandQueue) process(handler CommandHandler) (bool, error) {
	// We dequeue and fetch a command within a transaction.
	// During handling the command will be locked.
	qry := `
		SELECT 
			id,
			seq,
			action,
			params,
			deadline,
			created_at
		  FROM commands
		 WHERE state = 'requested'
		 ORDER BY seq ASC
		 LIMIT 1
		   FOR UPDATE SKIP LOCKED`

	// Begin transaction
	startedAt := time.Now()
	ctx := context.Background()
	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	// Select command
	cmd := &Command{}
	err = tx.QueryRow(ctx, qry).Scan(
		&cmd.ID,
		&cmd.Seq,
		&cmd.Action,
		&cmd.Params,
		&cmd.Deadline,
		&cmd.CreatedAt)
	if err != nil && err == pgx.ErrNoRows {
		return false, nil // Ok. There was just nothing to do.
	} else if err != nil {
		return false, err
	}

	// Apply command handler
	state := "success"
	result, err := handler(cmd)
	if err != nil {
		state = "error"
		result = fmt.Sprintf("%s", err)
	}

	// We are done
	stoppedAt := time.Now()

	// Write result
	qry = `
		UPDATE commands
		   SET state      = $2,
		       result     = $3,
			   started_at = $4,
			   stopped_at = $5,
		 WHERE id = $1`
	_, err = tx.Exec(ctx, qry, cmd.ID,
		state,
		result,
		startedAt,
		stoppedAt)
	if err != nil {
		return false, err
	}

	// End transaction
	err = tx.Commit(ctx)
	if err != nil {
		return false, err
	}

	return true, nil
}
