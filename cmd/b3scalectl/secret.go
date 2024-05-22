package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

func readSecret() (string, error) {
	fmt.Fprintln(os.Stderr, "Please paste your shared secret.")
	fmt.Fprintf(os.Stderr, "Secret: ")
	secret, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Fprintln(os.Stderr, "") // add missing newline
	return string(secret), nil
}

func readSecretOrEnv(ctx *cli.Context) (string, error) {
	secret := ctx.String("secret")
	if secret == "" { // Try env
		secret = config.EnvOpt(config.EnvJWTSecret, "")
	}
	if secret == "" { // Read from STDIN
		return readSecret()
	}
	return secret, nil
}
