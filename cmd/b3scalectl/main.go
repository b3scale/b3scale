package main

import (
	"fmt"
	"os"

	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

func getopt(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func main() {
	dbConnStr := getopt(
		"B3SCALE_DB_URL",
		"postgres://postgres:postgres@localhost:5432/b3scale")
	dbConn := store.Connect(dbConnStr)

	cmd := &store.Command{
		Action: "huhu",
		Params: []string{"foo", "bar", "triggered?"},
	}

	queue := store.NewCommandQueue(dbConn)
	err := queue.Queue(cmd)
	if err != nil {
		fmt.Println(err)
	}
}
