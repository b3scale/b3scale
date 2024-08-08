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
	"github.com/b3scale/b3scale/pkg/http/api/client"
	"github.com/b3scale/b3scale/pkg/http/auth"
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
				EnvVars: []string{config.EnvAPIURL},
			},
		},
		Action: c.showStatusHelp,
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
						Name:         "backend",
						Usage:        "show a specific cluster backend",
						Action:       c.showBackend,
						BashComplete: c.completeBackend,
					},
					{
						Name:   "frontends",
						Usage:  "show all frontends",
						Action: c.showFrontends,
					},
					{
						Name:         "frontend",
						Usage:        "show frontend settings",
						Action:       c.showFrontend,
						BashComplete: c.completeFrontend,
					},
					{
						Name:   "meetings",
						Usage:  "show running meetings",
						Action: c.showMeetings,
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
				Usage:   "deletes a backend or frontend config in the cluster",
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
						Name:         "backend",
						Usage:        "delete backend",
						Action:       c.deleteBackend,
						BashComplete: c.completeBackend,
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
				Usage:   "enables a backend in the cluster",
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
				Usage:   "disables a backend in the cluster",
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
				Name:  "completions",
				Usage: "shell completion for b3scalectl",
				Subcommands: []*cli.Command{
					{
						Name:   "bash",
						Usage:  "completions for BASH",
						Action: c.printBashCompletion,
					}, {
						Name:   "zsh",
						Usage:  "completions for ZSH",
						Action: c.printZshCompletion,
					},
				},
			},
			{
				Name:   "version",
				Usage:  "show version information",
				Action: c.showVersion,
			},
			{
				Name:   "export-openapi-schema",
				Usage:  "exports as OpenAPI Schema for the b3scale API",
				Action: c.exportOpenAPISchema,
			},
			{
				Name:    "create-meeting",
				Aliases: []string{"cm"},
				Usage:   "create a meeting for a frontend",
				Action:  c.createMeeting,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Required: true,
						Name:     "frontend",
						Usage:    "the frontend to create the meeting for",
					},
					&cli.StringFlag{
						Required: false,
						Name:     "name",
						Usage:    "name of the room. otherwise a random name will be generated.",
					},

					&cli.StringSliceFlag{
						Name:  "param",
						Usage: "key=value pairs to be passed to the meeting create request.",
					},
				},
			},
			{
				Name:  "db",
				Usage: "control database operations on the server",
				Subcommands: []*cli.Command{
					{
						Name:   "migrate",
						Usage:  "Apply all pending migrations to the database",
						Action: c.applyMigrations,
					},
				},
			},
			{
				Name:  "auth",
				Usage: "authorize users and node agents",
				Subcommands: []*cli.Command{
					{
						Name:   "authorize",
						Usage:  "Authorize b3scalectl for the current API host",
						Action: c.authorizeAPI,
					},
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

	token, err := auth.NewClaims(sub).
		WithScopesCSV(scopes).
		Sign(secret)
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
		ref = auth.GenerateRef(3)
	}

	secret, err := readSecretOrEnv(ctx)
	if err != nil {
		return err
	}

	token, err := auth.NewClaims(ref).
		WithScopes(auth.ScopeNode).
		Sign(secret)
	if err != nil {
		return err
	}

	fmt.Println(token)

	return nil
}

func (c *Cli) showStatusHelp(ctx *cli.Context) error {
	// Show help text
	cli.ShowAppHelp(ctx)
	fmt.Println("")
	return c.showStatus(ctx)
}

func (c *Cli) showStatus(ctx *cli.Context) error {
	apiHost := ctx.String("api")
	// Check if the token exists
	if !apiTokenExists(ctx) {
		fmt.Println("b3scalectl is not authorized: An API access token for this host is not present.")
		fmt.Println("API host:", apiHost)
		fmt.Println("\nUse `b3scalectl auth authorize` to create a new access token for this host.")
		fmt.Println("")
		return nil
	}

	// Show status
	client, err := apiClient(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "invalid or expired") {
			fmt.Println("Authentication Error: The access token is not longer valid.")
			fmt.Println("Reauthenticate using: `b3scalectl auth authorize`")
			fmt.Println("")
		}
		return err
	}
	status, err := client.Status(ctx.Context)
	if err != nil {
		return err
	}

	// Print server info
	fmt.Println("b3scale @", apiHost)
	fmt.Println("")
	fmt.Println("server version:", status.Version, "\tbuild:", status.Build)
	fmt.Println("   api version:", status.API)
	fmt.Println("")

	// Show migrations status
	dbStatus := status.Database
	dbVersion := "not initialized"
	if dbStatus.Error != nil {
		fmt.Println("Database Error:", *dbStatus.Error)
	}
	if dbStatus.Migration != nil {
		m := dbStatus.Migration
		dbVersion = fmt.Sprintf("v%d, '%s', applied at: %s", m.Version, m.Description, m.AppliedAt)
	}

	fmt.Println("Database:", dbStatus.Database)
	fmt.Println("Version: ", dbVersion)
	if dbStatus.PendingMigrations > 0 {
		fmt.Println("")
		fmt.Println("WARNING: The database is not migrated.")
		fmt.Println("         There are", dbStatus.PendingMigrations, "pending migrations.")
		fmt.Println("")
		fmt.Println("Use `b3scalectl db migrate` to apply all pending migrations.")
	}
	fmt.Println("")

	return nil
}

// applyMigrations applies all pending migrations
func (c *Cli) applyMigrations(ctx *cli.Context) error {
	client, err := apiClient(ctx)
	if err != nil {
		return nil
	}

	status, err := client.Status(ctx.Context)
	if err != nil {
		return err
	}
	dbStatus := status.Database

	if dbStatus.PendingMigrations == 0 {
		fmt.Println("There are currently no migrations to apply.")
		fmt.Println("The database is up to date.")
		return nil
	}
	fmt.Println("Applying", dbStatus.PendingMigrations, "pending migrations to the database.")

	if _, err := client.CtrlMigrate(ctx.Context); err != nil {
		return err
	}
	fmt.Println("Migration successful.")
	fmt.Println("")

	return c.showStatus(ctx)
}

// authorizeAPI b3scalectl for the current API host
func (c *Cli) authorizeAPI(ctx *cli.Context) error {
	apiHost := ctx.String("api")
	tokenFilename := apiTokenFilename(ctx)

	var accessToken string
	for {
		token, err := acquireToken(apiHost)
		if err != nil {
			return err
		}

		client := client.New(apiHost, token)
		_, err = client.Status(ctx.Context)
		if err != nil {
			fmt.Println("error using the token:", err)
		} else {
			accessToken = token
			break
		}
	}

	fmt.Println("Secret accepted, b3scalectl is now authorized.")

	// Persist token
	if err := config.UserDirPut(tokenFilename, []byte(accessToken)); err != nil {
		return err
	}

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
