package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/infra.run/public/b3scale/pkg/cluster"
)

// When receiving a SIGUSR1 signal, we trigger
// a reload. For now this is a pragmatic solution
// over providing a control socket.
// The drawback is that it is async. However this is
// already the way we are dealing with configs.

// SigCtl is a signal handler for a cluster controller
type SigCtl struct {
	controller *cluster.Controller
}

// NewSigCtl creates a new signal handler
func NewSigCtl(controller *cluster.Controller) *SigCtl {
	return &SigCtl{
		controller: controller,
	}
}

// Start the signal handler
func (ctl *SigCtl) Start() {
	log.Println("Starting signal handler: SIGUSR1")

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGUSR1)

	for {
		<-c // Await signal
		ctl.controller.Reload()
	}
}
