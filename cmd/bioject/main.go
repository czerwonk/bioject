package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/czerwonk/bioject/config"
	"github.com/czerwonk/bioject/database"
	"github.com/czerwonk/bioject/server"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const version = "0.1"

func main() {
	configFile := flag.String("config-file", "config.yml", "Path to config file")
	listenAddress := flag.String("api-listen-address", ":1337", "Listen address to listen for GRPC calls")
	bgpListenAddress := flag.String("bgp-listen-address", "0.0.0.0", "Listen address to listen for BGP connections")
	dbFile := flag.String("db-file", "routes.db", "Path to the database persisting routes")
	debug := flag.Bool("debug", false, "Enable debug log output")
	v := flag.Bool("v", false, "Show version info")

	flag.Parse()

	if *v {
		showVersion()
		os.Exit(0)
	}

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	cfg, err := loadConfigFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	metrics := server.NewMetrics()
	metrics.RegisterEndpoint(":9500")

	db, err := database.Connect("sqlite3", *dbFile)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer db.Close()

	log.Fatal(server.Start(cfg, *listenAddress, net.ParseIP(*bgpListenAddress), db, metrics))
}

func showVersion() {
	fmt.Println("bioject - Route injector based on BIO routing daemon")
	fmt.Println("Version:", version)
	fmt.Println("Author(s): Daniel Czerwonk")
}

func loadConfigFile(path string) (*config.Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %s", err)
	}

	r := bytes.NewReader(b)
	return config.Load(r)
}
