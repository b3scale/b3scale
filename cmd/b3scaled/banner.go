package main

import (
	"fmt"

	"github.com/b3scale/b3scale/pkg/config"
)

func banner() {
	fmt.Printf("    __   _____                      __        v.%s\n"+
		"   / /_ |__  /     ______________  / /__  \n"+
		"  / __ \\ /_ <     / ___/ ___/ __ \\/ / _ \\ \n"+
		" / /_/ /__/ /    (__  ) /__/ /_/ / /  __/ \n"+
		"/_.___/____/____/____/\\___/\\__,_/_/\\___/  \n\n", config.Version)
}
