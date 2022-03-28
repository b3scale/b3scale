package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"syscall"

	"github.com/urfave/cli/v2"
	"golang.org/x/term"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/config"
	v1 "gitlab.com/infra.run/public/b3scale/pkg/http/api/v1"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// RetNoChange indicates the return code, that no
// change was applied.
const RetNoChange = 64

// Frontend retrieval helper
func getFrontendByKey(
	ctx context.Context, c v1.Client, key string,
) (*store.FrontendState, error) {
	frontends, err := c.FrontendsList(ctx, url.Values{
		"key": []string{key},
	})
	if err != nil {
		return nil, err
	}
	if len(frontends) > 0 {
		return frontends[0], nil
	}
	return nil, nil
}

// Backend retrieval helper
func getBackendByHost(
	ctx context.Context, c v1.Client, key string,
) (*store.BackendState, error) {
	backends, err := c.BackendsList(ctx, url.Values{
		"host": []string{key},
	})
	if err != nil {
		return nil, err
	}
	if len(backends) > 0 {
		return backends[0], nil
	}
	return nil, nil
}

// Cli is the main command line interface application
type Cli struct {
	app        *cli.App
	returnCode int
}

// NewCli initializes the CLI application
func NewCli() *Cli {
	c := &Cli{}
	c.app = &cli.App{
		Usage:                "manage the b3scale cluster",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "api",
				Aliases: []string{"b"},
				Value:   "http://" + config.EnvListenHTTPDefault,
			},
		},
		Action: c.showStatus,
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
						Name:   "backend",
						Usage:  "show a specific cluster backend",
						Action: c.showBackend,
					},
					{
						Name:   "frontends",
						Usage:  "show all frontends",
						Action: c.showFrontends,
					},
					{
						Name:   "frontend",
						Usage:  "show frontend settings",
						Action: c.showFrontend,
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
								Name:    "secret",
								Aliases: []string{"s"},
								Usage:   "the bbb secret",
							},
							&cli.StringFlag{
								Name:    "opts",
								Aliases: []string{"j"},
								Usage:   "a generic settings property (as json)",
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
								Name:    "opts",
								Aliases: []string{"j"},
								Usage:   "a generic settings property (as json)",
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
			{
				Name: "auth",
				Subcommands: []*cli.Command{
					{
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "sub",
								Usage:    "userID or other identifier",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "scopes",
								Usage:    "a comma separated list of scopes",
								Required: true,
							},
						},
						Name:   "create_access_token",
						Usage:  "Create an access token for interacting with the API",
						Action: c.createAccessToken,
					},
				},
			},
		},
	}
	return c
}

// Auth: create access token. Scopes can be passed through
// options. A "sub" (user id) is required.
func (c *Cli) createAccessToken(ctx *cli.Context) error {
	sub := ctx.String("sub")
	scopes := ctx.String("scopes")
	scopes = strings.Join(strings.Split(scopes, ","), " ")

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "** Creating access token **")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "     Sub:", sub)
	fmt.Fprintln(os.Stderr, "  Scopes:", scopes)
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Please paste your shared secret.")
	fmt.Fprintf(os.Stderr, "Secret: ")
	secret, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "") // add missing newline

	token, err := v1.SignAccessToken(sub, scopes, secret)
	if err != nil {
		return err
	}

	fmt.Println(token)

	return nil
}

func (c *Cli) showStatus(ctx *cli.Context) error {
	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	status, err := client.Status(ctx.Context)
	if err != nil {
		return err
	}

	apiHost := ctx.String("api")

	// Print server info
	fmt.Println("b3scale @", apiHost)
	fmt.Println("")
	fmt.Println("server version:", status.Version, "\tbuild:", status.Build)
	fmt.Println("   api version:", status.API)
	fmt.Println("")

	return nil
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

	fmt.Println("getting frontend:", key)

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}

	// Get or create frontend
	state, err := getFrontendByKey(ctx.Context, client, key)
	if err != nil {
		return err
	}

	if state == nil {
		fmt.Println("creating frontend")
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
		if ctx.IsSet("opts") {
			if err := json.Unmarshal(
				[]byte(ctx.String("opts")), &state.Settings); err != nil {
				return err
			}
		}
		if !dry {
			state, err = client.FrontendCreate(ctx.Context, state)
			if err != nil {
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
				return fmt.Errorf("secret may not be empty for update")
			}
			changes = true
			state.Frontend.Secret = secret
		}

		if ctx.IsSet("opts") {
			// Frontend settings
			if err := json.Unmarshal(
				[]byte(ctx.String("opts")), &state.Settings); err != nil {
				return err
			}
			changes = true
		}

		if !changes {
			fmt.Println("no changes")
			c.returnCode = RetNoChange
		} else {
			if !dry {
				_, err := client.FrontendUpdate(ctx.Context, state)
				if err != nil {
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
	key := ctx.Args().Get(0)
	if key == "" {
		return fmt.Errorf("need frontend key for delete")
	}

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	state, err := getFrontendByKey(ctx.Context, client, key)
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
	_, err = client.FrontendDelete(ctx.Context, state)
	return err
}

// showFrontend displays information about a frontend
func (c *Cli) showFrontend(ctx *cli.Context) error {
	key := ctx.Args().Get(0)
	if key == "" {
		return fmt.Errorf("need frontend key for showing info")
	}

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}

	state, err := getFrontendByKey(ctx.Context, client, key)
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("no such frontend")
	}

	fmt.Println("Frontend:", state.Frontend.Key)
	fmt.Println("Settings:")
	s, _ := json.MarshalIndent(state.Settings, "   ", " ")
	fmt.Println(string(s))
	return nil
}

// show a list of all frontends
func (c *Cli) showFrontends(ctx *cli.Context) error {
	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	frontends, err := client.FrontendsList(ctx.Context, nil)
	if err != nil {
		return err
	}

	for _, f := range frontends {
		settings, _ := json.Marshal(f.Settings)
		fmt.Printf("%s\t%s\t%s\t%s\n", f.ID, f.Frontend.Key, f.Frontend.Secret, settings)
	}
	return nil
}

// setBackend manages the backends in the cluster
func (c *Cli) setBackend(ctx *cli.Context) error {
	adminState := ctx.String("state")
	dry := ctx.Bool("dry")

	// Args should be host
	host := ctx.Args().Get(0)
	if !strings.HasPrefix(host, "http") {
		return fmt.Errorf("host should start with http(s)://")
	}
	if !strings.HasSuffix(host, "/") {
		host += "/"
	}

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}

	// Check if backend exists
	state, err := getBackendByHost(ctx.Context, client, host)
	if err != nil {
		return err
	}
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
		})
		if ctx.IsSet("opts") {
			if err := json.Unmarshal(
				[]byte(ctx.String("opts")), &state.Settings); err != nil {
				return err
			}
		}
		if !dry {
			state, err = client.BackendCreate(ctx.Context, state)
			if err != nil {
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
		if ctx.IsSet("state") {
			if state.AdminState != adminState {
				state.AdminState = adminState
			}
			changes = true
		}
		if ctx.IsSet("opts") {
			if err := json.Unmarshal(
				[]byte(ctx.String("opts")), &state.Settings); err != nil {
				return err
			}
			changes = true
		}
		if changes {
			if !dry {
				state, err = client.BackendUpdate(ctx.Context, state)
				if err != nil {
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
	return nil
}

// showBackends displays information about our backend
func (c *Cli) showBackend(ctx *cli.Context) error {
	host := ctx.Args().Get(0)
	if host == "" {
		return fmt.Errorf("need host for showing backend info")
	}

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	// Check if backend exists
	backend, err := getBackendByHost(ctx.Context, client, host)
	if err != nil {
		return err
	}
	if backend == nil {
		return fmt.Errorf("backend not found")
	}

	fmt.Println("Backend:", backend.Backend.Host)
	fmt.Println("Settings:")
	s, _ := json.MarshalIndent(backend.Settings, "  ", "  ")
	fmt.Println(string(s))

	return nil
}

// showBackends displays a list of our backends
func (c *Cli) showBackends(ctx *cli.Context) error {
	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	// Check if backend exists
	backends, err := client.BackendsList(ctx.Context, nil)
	if err != nil {
		return err
	}
	for _, b := range backends {
		ratio := 0.0
		if b.MeetingsCount > 0 {
			ratio = float64(b.AttendeesCount) / float64(b.MeetingsCount)
		}
		settings, _ := json.Marshal(b.Settings)
		fmt.Printf("%s\n  Host:\t %s\n", b.ID, b.Backend.Host)
		fmt.Printf("  Settings:\t %v\n", string(settings))
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
	return c.setBackendAdminState(ctx, host, dry, "ready")
}

func (c *Cli) disableBackend(ctx *cli.Context) error {
	dry := ctx.Bool("dry")
	// Args should be host
	if ctx.NArg() < 1 {
		return fmt.Errorf("require: <host>")
	}
	host := ctx.Args().Get(0)
	return c.setBackendAdminState(ctx, host, dry, "stopped")
}

func (c *Cli) setBackendAdminState(
	ctx *cli.Context,
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

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}

	state, err := getBackendByHost(ctx.Context, client, host)
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
			state, err = client.BackendUpdate(ctx.Context, state)
			if err != nil {
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

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	host := ctx.Args().Get(0)
	state, err := getBackendByHost(ctx.Context, client, host)
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
		state, err = client.BackendDelete(
			ctx.Context, state, url.Values{
				"force": []string{"true"},
			})
		if err != nil {
			return err
		}
	} else {
		// Otherwise, we mark the node for decommissioning
		state.AdminState = "decommissioned"
		if dry {
			fmt.Println("skipping decommissioning backend (dry run)")
			return nil
		}
		state, err = client.BackendUpdate(ctx.Context, state)
		if err != nil {
			return err
		}
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
	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	host := ctx.Args().Get(0)
	state, err := getBackendByHost(ctx.Context, client, host)
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("no such backend")
	}

	cmd, err := client.BackendMeetingsEnd(ctx.Context, state.ID)
	if err != nil {
		return err
	}
	fmt.Println(cmd)

	return nil
}

// show the current version
func (c *Cli) showVersion(ctx *cli.Context) error {
	fmt.Printf("b3scalectl v.%s\t%s\n",
		config.Version,
		config.Build)
	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	status, err := client.Status(ctx.Context)
	if err != nil {
		return err
	}
	fmt.Println("API status:")
	fmt.Println(status)
	return nil
}

// Run starts the CLI
func (c *Cli) Run(ctx context.Context, args []string) error {
	return c.app.RunContext(ctx, args)
}
