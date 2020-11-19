package main

import (
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/urfave/cli/v2"

	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// Cli is the main command line interface application
type Cli struct {
	app   *cli.App
	queue *store.CommandQueue
	pool  *pgxpool.Pool
}

// NewCli initializes the CLI application
func NewCli(
	queue *store.CommandQueue,
	pool *pgxpool.Pool,
) *Cli {
	c := &Cli{
		pool:  pool,
		queue: queue,
	}
	c.app = &cli.App{
		Usage:                "manage the b3scale cluster",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "show",
				Aliases: []string{"s"},
				Usage:   "show the cluster state",
				Subcommands: []*cli.Command{
					{
						Name:   "backends",
						Usage:  "show the cluster backends",
						Action: c.showBackends,
					},
				},
			},
		},
	}
	return c
}

// showBackends displays a list of our backends
func (c *Cli) showBackends(ctx *cli.Context) error {
	backends, err := store.GetBackendStates(pool, store.Q())
	if err != nil {
		return err
	}

	return nil
}

// Run starts the CLI
func (c *Cli) Run(args []string) error {
	return c.app.Run(args)
}
