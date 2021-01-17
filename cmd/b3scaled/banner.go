package main

import (
	"fmt"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

func banner() {
	fmt.Printf("    __   _____                      __        v.%s\n", config.Version)
	fmt.Println("   / /_ |__  /     ______________ _/ /__  ")
	fmt.Println("  / __ \\ /_ <     / ___/ ___/ __ `/ / _ \\ ")
	fmt.Println(" / /_/ /__/ /    (__  ) /__/ /_/ / /  __/ ")
	fmt.Println("/_.___/____/____/____/\\___/\\__,_/_/\\___/  ")
	fmt.Println("")
}
