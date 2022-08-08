package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/b3scale/b3scale/pkg/config"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

func readSecret() ([]byte, error) {
	fmt.Fprintln(os.Stderr, "Please paste your shared secret.")
	fmt.Fprintf(os.Stderr, "Secret: ")
	secret, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, err
	}
	fmt.Fprintln(os.Stderr, "") // add missing newline
	return secret, nil
}

func readSecretOrEnv(ctx *cli.Context) ([]byte, error) {
	secretStr := ctx.String("secret")
	if secretStr == "" { // Try env
		secretStr = config.EnvOpt(config.EnvJWTSecret, "")
	}
	if secretStr == "" { // Read from STDIN
		return readSecret()
	}
	return []byte(secretStr), nil
}
