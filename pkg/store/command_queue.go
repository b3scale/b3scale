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

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const cmdQueue = "commands_queue"

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

	pool *pgxpool.Pool
}

// The CommandQueue is connected to the database and
// provides methods for queuing and dequeuing commands.
type CommandQueue struct {
	pool *pgxpool.Pool
}

// NewCommandQueue initializes a new command queue
func NewCommandQueue(conn *pgxpool.Pool) *CommandQueue {
	// Subscribe to notifications
	return &CommandQueue{
		pool: pool,
	}
}

// Subscribe will let the queue listen for notifications
func (q *CommandQueue) Subscribe() error {

}

// Queue adds a new command to the queue
func (q *CommandQueue) Queue(cmd *Command) error {
	return nil
}

// CommandHandler is a callback function for handling
// commands. The command was successful if no error was
// returned.
type CommandHandler func(*Command) error

// Receive will await a command and will block
// until a command can be processed. If the handler
// responds with an error, the error will be returned.
func Receive(handler CommandHandler) error {
	for {
		// Await command, after a timeout just try to dequeue

	}

	return nil
}
