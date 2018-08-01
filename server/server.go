package server

import (
	"github.com/czerwonk/bioject/config"
	log "github.com/sirupsen/logrus"
)

// Server is an server hosting BGP and API endoints
type Server struct {
}

// Start starts the server listening for BGP and API calls
func Start(c *config.Config, listenAddress string) error {
	bgp := newBGPserver()

	err := bgp.start(c)
	if err != nil {
		return err
	}

	log.Info("BGP server started")

	return startAPIServer(listenAddress, bgp)
}
