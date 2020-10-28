package main

import (
	"fmt"
	"log"
	"os"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
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
	dbConn, err := store.Connect(dbConnStr)
	if err != nil {
		log.Fatal(err)
	}

	cmd := cluster.AddBackend(&cluster.AddBackendRequest{
		Backend: &bbb.Backend{
			Host:   "https://bbbackend01.bastelgenosse.de/bigbluebutton/api",
			Secret: "nwlly97dniqQxKTWHdbGItgieGEBSnyL6s8I3FtM28",
		},
		Tags: []string{"sip", "2.0.0"},
	})
	queue := store.NewCommandQueue(dbConn)
	err = queue.Queue(cmd)
	if err != nil {
		fmt.Println(err)
	}
}
