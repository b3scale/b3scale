package main

import (
	"fmt"
)

func banner() {
	fmt.Printf("    __   _____                      __        v.%s\n", version)
	fmt.Println("   / /_ |__  /     ______________ _/ /__  ")
	fmt.Println("  / __ \\ /_ <     / ___/ ___/ __ `/ / _ \\ ")
	fmt.Println(" / /_/ /__/ /    (__  ) /__/ /_/ / /  __/ ")
	fmt.Println("/_.___/____/____/____/\\___/\\__,_/_/\\___/  ")
	fmt.Println("")
}
