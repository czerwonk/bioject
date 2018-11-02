package server

import (
	"net/http"

	log "github.com/sirupsen/logrus"
	"go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)

const ns = "bioject"

type Metrics struct {
	requestsTotal   *stats.Int64Measure
	requestsFailed  *stats.Int64Measure
	routesAdded     *stats.Int64Measure
	routesWithdrawn *stats.Int64Measure
}

func NewMetrics() *Metrics {
	m := &Metrics{
		requestsTotal:   stats.Int64(ns+"/requests_total", "Number of API requests received.", stats.UnitDimensionless),
		requestsFailed:  stats.Int64(ns+"/requests_failed", "Number of failed API requests.", stats.UnitDimensionless),
		routesAdded:     stats.Int64(ns+"/routes_added", "Number of routes added to RIB.", stats.UnitDimensionless),
		routesWithdrawn: stats.Int64(ns+"/routes_withdrawn", "Number of routes removed from RIB.", stats.UnitDimensionless),
	}

	return m
}

func (m *Metrics) RegisterEndpoint(addr string) {
	exporter, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatal(err)
	}
	view.RegisterExporter(exporter)
	http.Handle("/metrics", exporter)

	if err := view.Register(
		&view.View{
			Name:        m.requestsTotal.Name(),
			Description: m.requestsTotal.Description(),
			Measure:     m.requestsTotal,
			Aggregation: view.Count(),
		},
		&view.View{
			Name:        m.requestsFailed.Name(),
			Description: m.requestsFailed.Description(),
			Measure:     m.requestsFailed,
			Aggregation: view.Count(),
		},
		&view.View{
			Name:        m.routesAdded.Name(),
			Description: m.routesAdded.Description(),
			Measure:     m.routesAdded,
			Aggregation: view.Count(),
		},
		&view.View{
			Name:        m.routesWithdrawn.Name(),
			Description: m.routesWithdrawn.Description(),
			Measure:     m.routesWithdrawn,
			Aggregation: view.Count(),
		}); err != nil {
		log.Fatalf("Failed to register view: %v", err)
	}

	go func() {
		log.Infof("Start listening for metrics requests at http://%s/metrics", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}()
}
