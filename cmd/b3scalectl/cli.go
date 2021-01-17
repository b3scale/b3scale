package main

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/urfave/cli/v2"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// RetNoChange indicates the return code, that no
// change was applied.
const RetNoChange = 64

// Cli is the main command line interface application
type Cli struct {
	app        *cli.App
	queue      *store.CommandQueue
	pool       *pgxpool.Pool
	returnCode int
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
					{
						Name:   "frontends",
						Usage:  "show all frontends",
						Action: c.showFrontends,
					},
				},
			},
			{
				Name:    "set",
				Aliases: []string{"update", "u", "add", "a"},
				Usage:   "sets a backend or frontend config in the cluster",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry",
						Usage: "perform a dry run",
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:  "backend",
						Usage: "set backend params",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "tags",
								Aliases: []string{"t"},
								Usage:   "a csv list of tags for the backend",
							},
							&cli.StringFlag{
								Name:    "secret",
								Aliases: []string{"s"},
								Usage:   "the bbb secret",
							},
							&cli.StringFlag{
								Name:  "state",
								Usage: "set the admin_state of the backend node",
								Value: "init",
							},
						},
						Action: c.setBackend,
					},
					{
						Name:  "frontend",
						Usage: "set frontend params",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "secret",
								Required: true,
								Usage:    "the frontend specific bbb secret",
							},
						},
						Action: c.setFrontend,
					},
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"d", "del", "rm"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry",
						Usage: "perform a dry run",
					},
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "force delete",
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:   "backend",
						Usage:  "delete backend",
						Action: c.deleteBackend,
					},
					{
						Name:   "frontend",
						Usage:  "delete frontend",
						Action: c.deleteFrontend,
					},
				},
			},
			{
				Name:  "end",
				Usage: "force ending things on a backend",
				Subcommands: []*cli.Command{
					{
						Name:   "meetings",
						Usage:  "end all meetings on a given <host>",
						Action: c.endAllMeetings,
					},
				},
			},
			{
				Name:   "version",
				Action: c.showVersion,
			},
		},
	}
	return c
}

// setFrontend manages a forntend
func (c *Cli) setFrontend(ctx *cli.Context) error {
	dry := ctx.Bool("dry")

	// Args should be frontend key
	if ctx.NArg() < 1 {
		return fmt.Errorf("require: <frontend key>")
	}

	key := ctx.Args().Get(0)
	secret := ctx.String("secret")

	// Get or create frontend
	state, err := store.GetFrontendState(c.pool, store.Q().
		Where("key = ?", key))
	if err != nil {
		return err
	}

	if state == nil {
		// Create frontend
		if secret == "" {
			return fmt.Errorf("secret may not be empty")
		}
		state = store.InitFrontendState(c.pool, &store.FrontendState{
			Frontend: &bbb.Frontend{
				Key:    key,
				Secret: secret,
			},
		})
		if !dry {
			if err := state.Save(); err != nil {
				return err
			}
			fmt.Println("created frontend:", state.ID, state.Frontend.Key)
		} else {
			fmt.Println("skipped creating frontend")
		}
	} else {
		// Update Frontend
		changes := false
		if ctx.IsSet("secret") {
			if secret == "" {
				return fmt.Errorf("secret may not be empty")
			}
			changes = true
			state.Frontend.Secret = secret
		}

		if !changes {
			fmt.Println("no changes")
			c.returnCode = RetNoChange
		} else {
			if !dry {
				if err := state.Save(); err != nil {
					return err
				}
				fmt.Println("updated frontend")
			} else {
				fmt.Println("skipped saving changes in frontend")
			}
		}
	}

	return nil
}

// deleteFrontend removes a frontend from the cluster
func (c *Cli) deleteFrontend(ctx *cli.Context) error {
	dry := ctx.Bool("dry")

	// Args should be host
	if ctx.NArg() < 1 {
		return fmt.Errorf("require: <key>")
	}
	key := ctx.Args().Get(0)
	state, err := store.GetFrontendState(c.pool, store.Q().
		Where("key = ?", key))
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("no such frontend")
	}

	if dry {
		fmt.Println("skipping delete (dry)")
		return nil
	}

	fmt.Println("delete frontend:", state.ID)
	return state.Delete()
}

// show a list of all frontends
func (c *Cli) showFrontends(ctx *cli.Context) error {
	states, err := store.GetFrontendStates(c.pool, store.Q().
		OrderBy("frontends.key ASC"))
	if err != nil {
		return err
	}
	for _, f := range states {
		fmt.Printf("%s\t%s\t%s\n", f.ID, f.Frontend.Key, f.Frontend.Secret)
	}
	return nil
}

// setBackend manages the backends in the cluster
func (c *Cli) setBackend(ctx *cli.Context) error {
	adminState := ctx.String("state")
	dry := ctx.Bool("dry")

	// Args should be host
	if ctx.NArg() < 1 {
		return fmt.Errorf("require: <host>")
	}

	host := ctx.Args().Get(0)
	if !strings.HasPrefix(host, "http") {
		return fmt.Errorf("host should start with http(s)://")
	}
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}

	// Check if backend exists
	state, err := store.GetBackendState(c.pool, store.Q().
		Where("host = ?", host))
	if err != nil {
		return err
	}
	tags := strings.Split(ctx.String("tags"), ",")
	if state == nil {
		if !ctx.IsSet("secret") {
			return fmt.Errorf("need secret to create host")
		}
		// Create Backend
		state = store.InitBackendState(c.pool, &store.BackendState{
			Backend: &bbb.Backend{
				Host:   host,
				Secret: ctx.String("secret"),
			},
			AdminState: adminState,
			Tags:       tags,
		})
		if !dry {
			if err := state.Save(); err != nil {
				return err
			}
			fmt.Println("created backend:", state.ID, state.Backend.Host)
		} else {
			fmt.Println("skipped creating backend")
		}
	} else {
		// The state is known to use. Just make updates
		changes := false
		if ctx.IsSet("secret") {
			secret := ctx.String("secret")
			if secret == "" {
				return fmt.Errorf("secret may not be empty")
			}
			if state.Backend.Secret != secret {
				state.Backend.Secret = secret
				changes = true
			}
		}
		if ctx.IsSet("tags") {
			if !tagsEq(state.Tags, tags) {
				state.Tags = tags
				changes = true
			}
		}
		if ctx.IsSet("state") {
			if state.AdminState != adminState {
				state.AdminState = adminState
			}
			changes = true
		}
		if changes {
			if !dry {
				if err := state.Save(); err != nil {
					return err
				}
				fmt.Println("updated backend")
			} else {
				fmt.Println("skipping backend update")
			}
		} else {
			fmt.Println("no changes")
			c.returnCode = RetNoChange
		}
	}

	// Create command and enqueue
	if !dry {
		cmd := cluster.UpdateNodeState(&cluster.UpdateNodeStateRequest{
			ID: state.ID,
		})
		if err := c.queue.Queue(cmd); err != nil {
			return err
		}
	}

	return nil
}

// showBackends displays a list of our backends
func (c *Cli) showBackends(ctx *cli.Context) error {
	backends, err := store.GetBackendStates(c.pool, store.Q().
		OrderBy("backends.host ASC"))
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
		fmt.Printf("  MC/AC/R:\t %d/%d/%.02f\n",
			b.MeetingsCount,
			b.AttendeesCount,
			ratio)
		fmt.Printf("  Latency:\t%v\n", b.Latency)
		if b.NodeState == "error" && b.LastError != nil {
			fmt.Println("  LastError:", *b.LastError)
		}
		fmt.Println("")
	}

	return nil
}

// delete backend removes a backend from the store
func (c *Cli) deleteBackend(ctx *cli.Context) error {
	dry := ctx.Bool("dry")
	force := ctx.Bool("force")
	// Args should be host
	if ctx.NArg() < 1 {
		return fmt.Errorf("require: <host>")
	}
	host := ctx.Args().Get(0)
	state, err := store.GetBackendState(c.pool, store.Q().
		Where("host = ?", host))
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("no such backend")
	}

	// Check if we should hard delete. This can be either
	// the case when the deletion is forced, or when
	// the backend is in a non ready state
	hardDelete := force ||
		(state.NodeState != "ready" && state.NodeState != "init")

	if hardDelete {
		// The node is down anyhow we can issue a direct delete
		if dry {
			fmt.Println("skipping delete backend (dry run)")
			return nil
		}

		fmt.Println("deleting backend")
		return state.Delete()
	}

	// Otherwise, we mark the node for decommissioning
	state.AdminState = "decommissioned"
	if dry {
		fmt.Println("skipping decommissioning backend (dry run)")
		return nil
	}

	if err := state.Save(); err != nil {
		return err
	}

	fmt.Println("backend marked for decommissioning")
	return nil
}

// end all meetings on a backend
func (c *Cli) endAllMeetings(ctx *cli.Context) error {
	// Args should be host
	if ctx.NArg() < 1 {
		return fmt.Errorf("require: <host>")
	}
	host := ctx.Args().Get(0)
	state, err := store.GetBackendState(c.pool, store.Q().
		Where("host = ?", host))
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("no such backend")
	}

	cmd := cluster.EndAllMeetings(&cluster.EndAllMeetingsRequest{
		BackendID: state.ID,
	})
	if err := c.queue.Queue(cmd); err != nil {
		return err
	}

	return nil
}

// show the current version
func (c *Cli) showVersion(ctx *cli.Context) error {
	fmt.Printf("b3scalectl v.%s\t%s\n",
		config.Version,
		config.Build)
	return nil
}

// Run starts the CLI
func (c *Cli) Run(args []string) error {
	return c.app.Run(args)
}
