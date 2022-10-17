package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/api"
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
				Name:   "export-openapi-schema",
				Action: c.exportOpenAPISchema,
			},
			{
				Name: "auth",
				Subcommands: []*cli.Command{
					{
						Name:  "create_access_token",
						Usage: "Create an access token for interacting with the API",
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
							&cli.StringFlag{
								Name:  "secret",
								Usage: "shared secret, if not from env: B3SCALE_API_JWT_SECRET, if not present read from STDIN",
							},
						},
						Action: c.createAccessToken,
					},
					{
						Name:  "authorize_node_agent",
						Usage: "Create an access token for API access for a node agent",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "ref",
								Usage: "Agent reference for example backend-01, will be generated if not present",
							},
							&cli.StringFlag{
								Name:  "secret",
								Usage: "shared secret, if not from env B3SCALE_API_JWT_SECRET, if not present read from STDIN",
							},
						},
						Action: c.createNodeAccessToken,
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
	var err error

	sub := ctx.String("sub")
	scopes := ctx.String("scopes")
	scopes = strings.Join(strings.Split(scopes, ","), " ")

	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "** Creating access token **")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "     Sub:", sub)
	fmt.Fprintln(os.Stderr, "  Scopes:", scopes)
	fmt.Fprintln(os.Stderr, "")

	secret, err := readSecretOrEnv(ctx)
	if err != nil {
		return err
	}

	token, err := api.SignAccessToken(sub, scopes, secret)
	if err != nil {
		return err
	}

	fmt.Println(token)

	return nil
}

// Auth: create node access token.
func (c *Cli) createNodeAccessToken(ctx *cli.Context) error {
	var err error

	ref := ctx.String("ref")
	if ref == "" {
		ref = api.GenerateRef(3)
	}

	secret, err := readSecretOrEnv(ctx)
	if err != nil {
		return err
	}

	scopes := api.ScopeNode
	token, err := api.SignAccessToken(ref, scopes, secret)
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
	backend, err := getBackendByHost(ctx.Context, client, host)
	if err != nil {
		return err
	}
	if backend == nil {
		return fmt.Errorf("no such backend")
	}

	cmd, err := client.BackendMeetingsEnd(ctx.Context, backend.ID)
	if err != nil {
		return err
	}
	fmt.Println("Dispatch:", cmd.Action, cmd.Params)

	// Poll state changes
	state := cmd.State
	for {
		update, err := client.CommandRetrieve(ctx.Context, cmd.ID)
		if err != nil {
			return err
		}
		if update.State != state {
			fmt.Println("State:", update.State)
		}
		if update.State == "success" || update.State == "error" {
			fmt.Println("Result:", update.Result)
			break
		}

		state = update.State
		time.Sleep(500 * time.Millisecond)
	}
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

// Export the current openapi schema
func (c *Cli) exportOpenAPISchema(ctx *cli.Context) error {
	spec := api.NewAPISpec()
	doc, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(doc))
	return nil
}

// Run starts the CLI
func (c *Cli) Run(ctx context.Context, args []string) error {
	return c.app.RunContext(ctx, args)
}
