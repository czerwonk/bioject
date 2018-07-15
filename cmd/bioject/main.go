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
	v := flag.Bool("v", false, "Show version info")

	if *v {
		showVersion()
		os.Exit(0)
	}

	cfg, err := loadConfigFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	err = server.Start(cfg)
	if err != nil {
		log.Fatal(err)
	}

	select {}
}

func showVersion() {
	fmt.Println("bioject - Route Injector based on BIO Routing Daemon")
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
