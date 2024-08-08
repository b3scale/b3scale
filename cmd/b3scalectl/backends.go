package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/http/api"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/urfave/cli/v2"
)

// Backend retrieval helper
func getBackendByHost(
	ctx context.Context, c api.Client, key string,
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

func getBackendByID(
	ctx context.Context, c api.Client, key string,
) (*store.BackendState, error) {
	backends, err := c.BackendsList(ctx, url.Values{
		"id": []string{key},
	})
	if err != nil {
		return nil, err
	}
	if len(backends) > 0 {
		return backends[0], nil
	}
	return nil, nil
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
		return nil // we are done here
	}

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
		changes = true
	}
	if changes {
		if !dry {
			if ctx.IsSet("opts") {
				// Update backend settings using raw payload to
				// convey explicit null values.
				payload, err := json.Marshal(map[string]json.RawMessage{
					"settings": []byte(ctx.String("opts")),
				})
				if err != nil {
					return err
				}
				_, err = client.BackendUpdateRaw(ctx.Context, state.ID, payload)
				if err != nil {
					return err
				}
			} else {
				state, err = client.BackendUpdate(ctx.Context, state)
				if err != nil {
					return err
				}
			}
			fmt.Println("updated backend")
		} else {
			fmt.Println("skipping backend update")
		}
	} else {
		fmt.Println("no changes")
		c.returnCode = RetNoChange
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
	backends, err := client.BackendsList(ctx.Context)
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
		return fmt.Errorf("backend not found")
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

func (c *Cli) completeBackend(ctx *cli.Context) {
	// This will complete if no args are passed
	if ctx.NArg() > 0 {
		return
	}

	client, err := apiClient(ctx)
	if err != nil {
		return
	}
	// Check if backend exists
	backends, err := client.BackendsList(ctx.Context)
	if err != nil {
		return
	}
	for _, b := range backends {
		fmt.Println(b.Backend.Host)
	}
}
