package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
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
	listenAddress := flag.String("listen-address", ":1337", "Listen address to listen for GRPC calls")
	dbFile := flag.String("db-file", "routes.db", "Path to the database persisting routes")
	v := flag.Bool("v", false, "Show version info")

	flag.Parse()

	if *v {
		showVersion()
		os.Exit(0)
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

	log.Fatal(server.Start(cfg, *listenAddress, db, metrics))
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
