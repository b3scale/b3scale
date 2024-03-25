package main

import (
	_ "embed"
	"fmt"

	"github.com/urfave/cli/v2"
)

//go:embed completions/bash_autocomplete
var bashAutocomplete string

//go:embed completions/zsh_autocomplete
var zshAutocomplete string

func (c *Cli) printBashCompletion(ctx *cli.Context) error {
	fmt.Print(string(bashAutocomplete))
	return nil
}

func (c *Cli) printZshCompletion(ctx *cli.Context) error {
	fmt.Print(string(zshAutocomplete))
	return nil
}
