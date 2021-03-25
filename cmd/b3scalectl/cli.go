package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
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
	returnCode int
}

// NewCli initializes the CLI application
func NewCli(
	queue *store.CommandQueue,
) *Cli {
	c := &Cli{}
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
								Name:    "prop",
								Aliases: []string{"p"},
								Usage:   "a generic settings property",
							},
						},
						Action: c.setBackend,
					},
					{
						Name:  "frontend",
						Usage: "set frontend params",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "secret",
								Usage: "the frontend specific bbb secret",
							},
							&cli.StringFlag{
								Name:    "prop",
								Aliases: []string{"p"},
								Usage:   "a generic settings property",
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
				Name:    "enable",
				Aliases: []string{"en", "start"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry",
						Usage: "perform a dry run",
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:   "backend",
						Usage:  "enable backend",
						Action: c.enableBackend,
					},
				},
			},
			{
				Name:    "disable",
				Aliases: []string{"dis", "stop"},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry",
						Usage: "perform a dry run",
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:   "backend",
						Usage:  "disable backend",
						Action: c.disableBackend,
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

	// Begin TX
	tx, err := store.ConnectionFromContext(ctx.Context).Begin(ctx.Context)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
	}
	defer tx.Rollback(ctx.Context)

	// Get or create frontend
	state, err := store.GetFrontendState(ctx.Context, tx, store.Q().
		Where("key = ?", key))
	if err != nil {
		return err
	}

	if state == nil {
		// Create frontend
		if secret == "" {
			return fmt.Errorf("secret may not be empty")
		}
		state = store.InitFrontendState(&store.FrontendState{
			Frontend: &bbb.Frontend{
				Key:    key,
				Secret: secret,
			},
		})
		if ctx.IsSet("prop") {
			propKey, propValue := parseSetProp(ctx.String("prop"))
			state.Settings.Set(propKey, propValue)
		}
		if !dry {
			if err := state.Save(ctx.Context, tx); err != nil {
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

		if ctx.IsSet("prop") {
			propKey, propValue := parseSetProp(ctx.String("prop"))
			if state.Settings.Get(propKey, nil) != propValue {
				changes = true
			}
			state.Settings.Set(propKey, propValue)
		}

		if !changes {
			fmt.Println("no changes")
			c.returnCode = RetNoChange
		} else {
			if !dry {
				if err := state.Save(ctx.Context, tx); err != nil {
					return err
				}
				fmt.Println("updated frontend")
			} else {
				fmt.Println("skipped saving changes in frontend")
			}
		}
	}

	// Commit changes
	return tx.Commit(ctx.Context)
}

// deleteFrontend removes a frontend from the cluster
func (c *Cli) deleteFrontend(ctx *cli.Context) error {
	dry := ctx.Bool("dry")

	// Begin TX
	tx, err := store.ConnectionFromContext(ctx.Context).Begin(ctx.Context)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
	}
	defer tx.Rollback(ctx.Context)

	// Args should be host
	if ctx.NArg() < 1 {
		return fmt.Errorf("require: <key>")
	}
	key := ctx.Args().Get(0)
	state, err := store.GetFrontendState(ctx.Context, tx, store.Q().
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
	if err := state.Delete(ctx.Context, tx); err != nil {
		return err
	}

	return tx.Commit(ctx.Context)
}

// show a list of all frontends
func (c *Cli) showFrontends(ctx *cli.Context) error {
	// Begin TX
	tx, err := store.ConnectionFromContext(ctx.Context).Begin(ctx.Context)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
	}
	defer tx.Rollback(ctx.Context)
	states, err := store.GetFrontendStates(ctx.Context, tx, store.Q().
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
	// Begin TX
	tx, err := store.ConnectionFromContext(ctx.Context).Begin(ctx.Context)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
	}
	defer tx.Rollback(ctx.Context)

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
	state, err := store.GetBackendState(ctx.Context, tx, store.Q().
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
		state = store.InitBackendState(&store.BackendState{
			Backend: &bbb.Backend{
				Host:   host,
				Secret: ctx.String("secret"),
			},
			AdminState: adminState,
			Tags:       tags,
		})
		if ctx.IsSet("prop") {
			propKey, propValue := parseSetProp(ctx.String("prop"))
			state.Settings.Set(propKey, propValue)
		}
		if !dry {
			if err := state.Save(ctx.Context, tx); err != nil {
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
		if ctx.IsSet("prop") {
			propKey, propValue := parseSetProp(ctx.String("prop"))
			if state.Settings.Get(propKey, nil) != propValue {
				changes = true
			}
			state.Settings.Set(propKey, propValue)
		}
		if changes {
			if !dry {
				if err := state.Save(ctx.Context, tx); err != nil {
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
		if err := store.QueueCommand(ctx.Context, tx, cmd); err != nil {
			return err
		}
	}

	return tx.Commit(ctx.Context)
}

// showBackends displays a list of our backends
func (c *Cli) showBackends(ctx *cli.Context) error {
	// Begin TX
	tx, err := store.ConnectionFromContext(ctx.Context).Begin(ctx.Context)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
	}
	defer tx.Rollback(ctx.Context)

	backends, err := store.GetBackendStates(ctx.Context, tx, store.Q().
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
		fmt.Printf("  LoadFactor:\t %v\n", b.LoadFactor)
		fmt.Printf("  Latency:\t %v\n", b.Latency)
		if b.NodeState == "error" && b.LastError != nil {
			fmt.Println("  LastError:", *b.LastError)
		}
		fmt.Println("")
	}

	return nil
}

// enable a backend means setting the admin state
// to ready
func (c *Cli) enableBackend(ctx *cli.Context) error {
	dry := ctx.Bool("dry")
	// Args should be host
	if ctx.NArg() < 1 {
		return fmt.Errorf("require: <host>")
	}
	host := ctx.Args().Get(0)
	return c.setBackendAdminState(ctx.Context, host, dry, "ready")
}

func (c *Cli) disableBackend(ctx *cli.Context) error {
	dry := ctx.Bool("dry")
	// Args should be host
	if ctx.NArg() < 1 {
		return fmt.Errorf("require: <host>")
	}
	host := ctx.Args().Get(0)
	return c.setBackendAdminState(ctx.Context, host, dry, "stopped")
}

func (c *Cli) setBackendAdminState(
	ctx context.Context,
	host string,
	dry bool,
	adminState string,
) error {
	if !strings.HasPrefix(host, "http") {
		return fmt.Errorf("host should start with http(s)://")
	}
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}

	// Begin TX
	tx, err := store.ConnectionFromContext(ctx).Begin(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
	}
	defer tx.Rollback(ctx)

	// Check if backend exists
	state, err := store.GetBackendState(ctx, tx, store.Q().
		Where("host = ?", host))
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("backend not found")
	}

	// The state is known to use. Just make updates
	changes := false
	if state.AdminState != adminState {
		state.AdminState = adminState
		changes = true
	}
	if changes {
		if !dry {
			if err := state.Save(ctx, tx); err != nil {
				return err
			}
			fmt.Println("updated backend admin state")
		} else {
			fmt.Println("skipping backend admin state update")
		}
	} else {
		fmt.Println("no changes")
		c.returnCode = RetNoChange
	}

	// Create command and enqueue
	if !dry {
		cmd := cluster.UpdateNodeState(&cluster.UpdateNodeStateRequest{
			ID: state.ID,
		})
		if err := store.QueueCommand(ctx, tx, cmd); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// delete backend removes a backend from the store
func (c *Cli) deleteBackend(ctx *cli.Context) error {
	dry := ctx.Bool("dry")
	force := ctx.Bool("force")
	// Args should be host
	if ctx.NArg() < 1 {
		return fmt.Errorf("require: <host>")
	}

	// Begin TX
	tx, err := store.ConnectionFromContext(ctx.Context).Begin(ctx.Context)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
	}
	defer tx.Rollback(ctx.Context)

	host := ctx.Args().Get(0)
	state, err := store.GetBackendState(ctx.Context, tx, store.Q().
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
		if err := state.Delete(ctx.Context, tx); err != nil {
			return err
		}
	} else {
		// Otherwise, we mark the node for decommissioning
		state.AdminState = "decommissioned"
		if dry {
			fmt.Println("skipping decommissioning backend (dry run)")
			return nil
		}

		if err := state.Save(ctx.Context, tx); err != nil {
			return err
		}
	}

	fmt.Println("backend marked for decommissioning")
	return tx.Commit(ctx.Context)
}

// end all meetings on a backend
func (c *Cli) endAllMeetings(ctx *cli.Context) error {
	// Args should be host
	if ctx.NArg() < 1 {
		return fmt.Errorf("require: <host>")
	}
	host := ctx.Args().Get(0)

	// Begin TX
	tx, err := store.ConnectionFromContext(ctx.Context).Begin(ctx.Context)
	if err != nil {
		log.Fatal().Err(err).Msg("could not start transaction")
	}
	defer tx.Rollback(ctx.Context)

	state, err := store.GetBackendState(ctx.Context, tx, store.Q().
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
	if err := store.QueueCommand(ctx.Context, tx, cmd); err != nil {
		return err
	}

	return tx.Commit(ctx.Context)
}

// show the current version
func (c *Cli) showVersion(ctx *cli.Context) error {
	fmt.Printf("b3scalectl v.%s\t%s\n",
		config.Version,
		config.Build)
	return nil
}

// Run starts the CLI
func (c *Cli) Run(ctx context.Context, args []string) error {
	conn, err := store.Acquire(ctx)
	if err != nil {
		return err
	}
	ctx = store.ContextWithConnection(ctx, conn)

	return c.app.RunContext(ctx, args)
}

// Helper: parseSetProp will decode a property
// set request of the form my.key=value.
// The value is decoded into
func parseSetProp(prop string) (string, interface{}) {
	t := strings.Split(prop, "=")
	if len(t) != 2 {
		panic("syntax error in prop: must be of format '<key> = <value>'")
	}
	key := strings.TrimSpace(t[0])
	value := strings.TrimSpace(t[1])
	if value == "" {
		return key, nil
	}
	return key, value
}
