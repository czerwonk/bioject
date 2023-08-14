// SPDX-FileCopyrightText: (c) 2018 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/czerwonk/bioject/pkg/config"
	"github.com/czerwonk/bioject/pkg/database"
	"github.com/czerwonk/bioject/pkg/server"
	"github.com/czerwonk/bioject/pkg/tracing"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

const version = "0.5.0"

func main() {
	configFile := flag.String("config-file", "config.yml", "Path to config file")
	listenAddress := flag.String("api-listen-address", ":1337", "Listen address to listen for GRPC calls")
	bgpListenAddress := flag.String("bgp-listen-address", "0.0.0.0", "Listen address to listen for BGP connections")
	tracingEnabled := flag.Bool("tracing.enabled", false, "Enables tracing using OpenTelemetry")
	tracingProvider := flag.String("tracing.provider", "", "Sets the tracing provider (stdout or collector)")
	tracingCollectorEndpoint := flag.String("tracing.collector.grpc-endpoint", "", "Sets the tracing provider (stdout or collector)")
	dbFile := flag.String("db-file", "routes.db", "Path to the database persisting routes")
	debug := flag.Bool("debug", false, "Enable debug log output")
	v := flag.Bool("v", false, "Show version info")

	flag.Parse()

	if *v {
		showVersion()
		os.Exit(0)
	}

  logrus.Infof("Staring bioject (Version: %s)", version)

	cfg, err := loadConfigFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	if *debug || cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdownTracing, err := tracing.Init(ctx, version, *tracingEnabled, *tracingProvider, *tracingCollectorEndpoint)
	if err != nil {
		log.Fatalf("could not initialize tracing: %v", err)
	}
	defer shutdownTracing()

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
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %s", err)
	}

	r := bytes.NewReader(b)
	return config.Load(r)
}
