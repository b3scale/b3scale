package main

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/urfave/cli/v2"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
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
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "adds a node or frontend to the cluster",
				Subcommands: []*cli.Command{
					{
						Name:  "backend",
						Usage: "add a new backend",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "tags",
								Aliases: []string{"t"},
								Usage:   "a csv list of tags for the backend",
							},
						},
						Action: c.addBackend,
					},
				},
			},
		},
	}
	return c
}

// addBackend adds a backend to the cluster
func (c *Cli) addBackend(ctx *cli.Context) error {
	// Args should be host and secret
	if ctx.NArg() < 2 {
		return fmt.Errorf("require: <host> <secret>")
	}

	host := ctx.Args().Get(0)
	secret := ctx.Args().Get(1)

	if !strings.HasPrefix(host, "http") {
		return fmt.Errorf("host should start with http(s)://")
	}
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}

	fmt.Println("adding backend:", host)

	tags := strings.Split(ctx.String("tags"), ",")

	// Create command and enqueue
	cmd := cluster.AddBackend(&cluster.AddBackendRequest{
		Backend: &bbb.Backend{
			Host:   host,
			Secret: secret,
		},
		Tags: tags,
	})

	return c.queue.Queue(cmd)
}

// showBackends displays a list of our backends
func (c *Cli) showBackends(ctx *cli.Context) error {
	backends, err := store.GetBackendStates(c.pool, store.Q())
	if err != nil {
		return err
	}
	if len(backends) == 0 {
		fmt.Println("no backends found.")
		return nil
	}

	for _, b := range backends {
		ratio := 0.0
		if b.MeetingsCount > 0 {
			ratio = float64(b.AttendeesCount) / float64(b.MeetingsCount)
		}

		fmt.Printf("%s\n  Host:\t %s\n", b.ID, b.Backend.Host)
		fmt.Printf("  Tags:\t %v\n", b.Tags)
		fmt.Printf("  NodeState:\t %s\t", b.NodeState)
		fmt.Printf("  AdminState:\t %s\n", b.AdminState)
		fmt.Printf("  AC/MC/R:\t %d/%d/%.02f\n",
			b.AttendeesCount,
			b.MeetingsCount,
			ratio)
		fmt.Printf("  Latency:\t%v\n", b.Latency)
		if b.NodeState == "error" {
			fmt.Println("  LastError:", b.LastError)
		}
		fmt.Println("")
	}

	return nil
}

// Run starts the CLI
func (c *Cli) Run(args []string) error {
	return c.app.Run(args)
}
