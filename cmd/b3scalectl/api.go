package main

import (
	"fmt"
	"syscall"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/b3scale/b3scale/pkg/http/api"
	"github.com/b3scale/b3scale/pkg/http/api/client"
	"github.com/b3scale/b3scale/pkg/http/auth"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

// apiTokenFilename makes the full path and filename of the
// local access token for the host
func apiTokenFilename(ctx *cli.Context) string {
	apiHost := ctx.String("api")
	tokenFilename := apiHost + ".access_token"
	return tokenFilename
}

// apiTokenGet checks if the access token is present,
// and returns the token from the user dir.
func apiTokenExists(ctx *cli.Context) bool {
	filename := apiTokenFilename(ctx)
	token, _ := config.UserDirGetString(filename)
	return token != ""
}

// apiClient initializes the applications API client
func apiClient(ctx *cli.Context) (api.Client, error) {
	apiHost := ctx.String("api")
	tokenFilename := apiHost + ".access_token"

	var (
		token string
		err   error
	)

	// Check if we have an access token, otherwise acquire
	// one by requesting the shared JWT secret.
	token, _ = config.UserDirGetString(tokenFilename)
	if token == "" {
		token, err = acquireToken(apiHost)
		if err != nil {
			return nil, err
		}
		if err := config.UserDirPut(tokenFilename, []byte(token)); err != nil {
			return nil, err
		}
	}

	// Initialize client and test connection
	client := client.New(apiHost, token)
	status, err := client.Status(ctx.Context)
	if err != nil {
		return nil, err
	}

	if !status.IsAdmin {
		return nil, fmt.Errorf("authorization failed")
	}

	return client, nil
}

func accessTokenPath(apiHost string) (string, error) {
	return config.UserDirPath(apiHost + ".access_token")
}

// Create an access token
func acquireToken(apiHost string) (string, error) {
	// Check if we have an access token, otherwise acquire
	// one by requesting the shared JWT secret.
	tokenFullPath, err := accessTokenPath(apiHost)
	if err != nil {
		return "", err
	}

	fmt.Println("")
	fmt.Println("** Authorization required for", apiHost, "**")
	fmt.Println("")
	fmt.Println("Please paste your shared secret here. The generated")
	fmt.Println("access token will be stored in:")
	fmt.Println("")
	fmt.Println("  ", tokenFullPath)
	fmt.Println("")
	fmt.Print("Secret: ")
	secret, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println("") // add missing newline

	if len(secret) == 0 {
		return "", fmt.Errorf("secret should not be empty")
	}

	return auth.NewClaims("b3scalectl").
		WithScopes(auth.ScopeAdmin).
		Sign(string(secret))
}
