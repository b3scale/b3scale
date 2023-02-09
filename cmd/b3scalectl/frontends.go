package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/b3scale/b3scale/pkg/bbb"
	"github.com/b3scale/b3scale/pkg/http/api"
	"github.com/b3scale/b3scale/pkg/store"
	"github.com/urfave/cli/v2"
)

// Frontend retrieval helper
func getFrontendByKey(
	ctx context.Context, c api.Client, key string,
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

// setFrontend manages a frontend
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
	if state == nil && err == nil {
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
	}
	if err != nil { // Something else failed.
		return err
	}
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
		changes = true
	}

	if !changes {
		fmt.Println("no changes")
		c.returnCode = RetNoChange
	} else {
		if !dry {
			if ctx.IsSet("opts") {
				// Update frontend settings using raw payload to
				// convey explicit null values.
				payload, err := json.Marshal(map[string]json.RawMessage{
					"settings": []byte(ctx.String("opts")),
				})
				if err != nil {
					return err
				}
				_, err = client.FrontendUpdateRaw(ctx.Context, state.ID, payload)
				if err != nil {
					return err
				}
			} else {
				_, err := client.FrontendUpdate(ctx.Context, state)
				if err != nil {
					return err
				}
			}
			fmt.Println("updated frontend")
		} else {
			fmt.Println("skipped saving changes in frontend")
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
    active := "enabled"
    if !f.Active {
      active = "disabled"
    }
		fmt.Printf("%s\t%s\t%s\t%s\t%s\n", f.ID, f.Frontend.Key, f.Frontend.Secret, active, settings)
	}
	return nil
}

// enableFrontend activates a frontend from the cluster
func (c *Cli) enableFrontend(ctx *cli.Context) error {
	dry := ctx.Bool("dry")

	// Args should be host
	key := ctx.Args().Get(0)
	if key == "" {
		return fmt.Errorf("need frontend key for enable")
	}

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	state, err := getFrontendByKey(ctx.Context, client, key)
	if state == nil {
		return fmt.Errorf("frontend with key %s not found", key)
	}

	if dry {
		fmt.Println("skipping enable (dry)")
		return nil
	}

	fmt.Println("enable frontend:", state.ID)
  state.Active = true
  _, err = client.FrontendUpdate(ctx.Context, state)
	return err
}

// disableFrontend deactivates a frontend from the cluster
func (c *Cli) disableFrontend(ctx *cli.Context) error {
	dry := ctx.Bool("dry")

	// Args should be host
	key := ctx.Args().Get(0)
	if key == "" {
		return fmt.Errorf("need frontend key for disable")
	}

	client, err := apiClient(ctx)
	if err != nil {
		return err
	}
	state, err := getFrontendByKey(ctx.Context, client, key)
	if state == nil {
		return fmt.Errorf("frontend with key %s not found", key)
	}

	if dry {
		fmt.Println("skipping disable (dry)")
		return nil
	}

  fmt.Println("came here")
	fmt.Println("disable frontend:", state.ID)
  state.Active = false
  _, err = client.FrontendUpdate(ctx.Context, state)
	return err
}
