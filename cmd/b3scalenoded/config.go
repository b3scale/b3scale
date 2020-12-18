package main

import (
	"fmt"

	"gitlab.com/infra.run/public/b3scale/pkg/config"
)

// Make a redis url from the BBB config
func redisURL(conf config.Properties) string {
	host, ok := conf.Get("redisHost")
	if !ok {
		host = "localhost"
	}
	port, ok := conf.Get("redisPort")
	if !ok {
		port = "6379"
	}
	pass, _ := conf.Get("redisPassword")

	return fmt.Sprintf(
		"redis://:%s@%s:%s/1",
		pass, host, port)
}
