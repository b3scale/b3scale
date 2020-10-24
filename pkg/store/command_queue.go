package store

/*
 Commands are serialized operations to be executed
 by any b3scale instance.

 The store implements a command queue for processing
 commands.
*/

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const cmdQueue = "commands_queue"

// CommandHandler is a callback function for handling
// commands. The command was successful if no error was
// returned.
type CommandHandler func(*Command) error

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
	return nil
}

// Receive will await a command and will block
// until a command can be processed. If the handler
// responds with an error, the error will be returned.
func (q *CommandQueue) Receive(handler CommandHandler) error {
	for {
		ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
		// Await command, after a timeout just try to dequeue
		if q.conn != nil {
			_, err := q.conn.Conn().WaitForNotification(ctx)
			if err != nil {
				return err
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

	return nil
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
		   FOR UPDATE SKIP LOCKED
		 WHERE state = 'requested'
		 LIMIT 1`

	// Begin transaction
	ctx := context.Background()
	tx, err := q.pool.Begin(ctx)
	if err != nil {
		return false, err
	}

	// Select command
	cmd := &Command{}
	err = tx.QueryRow(ctx, qry).Scan(
		&cmd.ID,
		&cmd.Seq,
		&cmd.Action,
		&cmd.Params,
		&cmd.Deadline,
		&cmd.CreatedAt)
	if err != nil {
		err1 := tx.Rollback(ctx)
		if err1 != nil {
			return false, err1
		}
		return false, err
	}

	return false, nil
}
