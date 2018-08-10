package server

import (
	"fmt"

	"github.com/czerwonk/bioject/config"
	"github.com/czerwonk/bioject/database"
	log "github.com/sirupsen/logrus"
)

// Server is an server hosting BGP and API endoints
type Server struct {
}

// Start starts the server listening for BGP and API calls
func Start(c *config.Config, listenAddress, dbFile string) error {
	bgp := newBGPserver()

	db, err := database.Connect("sqlite3", dbFile)
	if err != nil {
		return fmt.Errorf("could not connect to database: %v", err)
	}
	defer db.Close()

	err = bgp.start(c)
	if err != nil {
		return fmt.Errorf("could not start BGP speaker: %v", err)
	}

	log.Info("BGP server started")

	err = restoreRoutes(bgp, db)
	if err != nil {
		return fmt.Errorf("could not restore routes: %v", err)
	}

	return startAPIServer(listenAddress, bgp, db)
}

func restoreRoutes(bgp *bgpServer, db *database.Database) error {
	routes, err := db.Routes()
	if err != nil {
		return err
	}

	for _, r := range routes {
		log.Infof("Restoring route: %s via %s", r.Prefix, r.NextHop)
		pfx, path, err := convertToBioRoute(r)
		if err != nil {
			return fmt.Errorf("could not convert %s via %s: %v", r.Prefix, r.NextHop, err)
		}

		err = bgp.addPath(pfx, path)
		if err != nil {
			return fmt.Errorf("could not restore route %s via %s: %v", r.Prefix, r.NextHop, err)
		}
	}

	return nil
}
