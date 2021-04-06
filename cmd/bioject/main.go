package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"contrib.go.opencensus.io/exporter/zipkin"
	"github.com/czerwonk/bioject/pkg/config"
	"github.com/czerwonk/bioject/pkg/database"
	"github.com/czerwonk/bioject/pkg/server"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

const version = "0.3.0"

func main() {
	configFile := flag.String("config-file", "config.yml", "Path to config file")
	listenAddress := flag.String("api-listen-address", ":1337", "Listen address to listen for GRPC calls")
	bgpListenAddress := flag.String("bgp-listen-address", "0.0.0.0", "Listen address to listen for BGP connections")
	zipkinEndpoint := flag.String("zipkin-endpoint", "", "Zipkin endpoint for tracing information")
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

	if len(*zipkinEndpoint) > 0 {
		err := enableTracing(*zipkinEndpoint)
		if err != nil {
			log.Fatalf("could not register zipkin exporter: %v", err)
		}
	}

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

func enableTracing(endpoint string) error {
	localEndpoint, err := openzipkin.NewEndpoint("bioject", ":0")
	if err != nil {
		return err
	}

	reporter := zipkinHTTP.NewReporter(endpoint)
	exporter := zipkin.NewExporter(reporter, localEndpoint)
	trace.RegisterExporter(exporter)

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	return nil
}
