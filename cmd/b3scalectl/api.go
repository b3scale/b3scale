package main

import (
	"fmt"
	"syscall"

	"github.com/urfave/cli/v2"
	"github.com/b3scale/b3scale/pkg/config"
	v1 "github.com/b3scale/b3scale/pkg/http/api/v1"
	"golang.org/x/term"
)

// apiClient initializes the applications API client
func apiClient(ctx *cli.Context) (v1.Client, error) {
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
	client := v1.NewJWTClient(apiHost, token)
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

	return v1.SignAdminAccessToken("b3scalectl", secret)
}
