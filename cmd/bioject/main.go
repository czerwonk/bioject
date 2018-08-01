package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/czerwonk/bioject/config"
	"github.com/czerwonk/bioject/server"
)

const version = "0.1"

func main() {
	configFile := flag.String("config-file", "config.yml", "Path to config file")
	listenAddress := flag.String("listen-address", ":1337", "Listen address to listen for GRPC calls")
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

	err = server.Start(cfg, *listenAddress)
	if err != nil {
		log.Fatal(err)
	}
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
