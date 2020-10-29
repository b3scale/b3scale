package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	//	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	//	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
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

	states, err := store.GetBackendStates(dbConn, store.NewQuery())
	if err != nil {
		log.Fatal(err)
	}

	for _, state := range states {
		res, _ := json.Marshal(state)
		fmt.Println(string(res))
		fmt.Println("L:", state.Latency)
	}

	/*
		cmd := cluster.AddBackend(&cluster.AddBackendRequest{
			Backend: &bbb.Backend{
				Host:   "https://bbbackend01.bastelgenosse.de/bigbluebutton/aapi",
				Secret: "nwlly97dniqQxKTWHdbGItgieGEBSnyL6s8I3FtM28",
			},
			Tags: []string{"sip", "2.0.0"},
		})
		queue := store.NewCommandQueue(dbConn)
		err = queue.Queue(cmd)
		if err != nil {
			fmt.Println(err)
		}
	*/
}
