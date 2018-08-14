package server

import (
	log "github.com/sirupsen/logrus"

	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"
)

const ns = "bioject"

type Metrics struct {
	requestsTotal   prometheus.Counter
	requestsFailed  prometheus.Counter
	routesAdded     prometheus.Counter
	routesWithdrawn prometheus.Counter
}

func NewMetrics() *Metrics {
	m := &Metrics{
		requestsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: "requests",
			Name:      "total",
			Help:      "Number of API requests received.",
		}),
		requestsFailed: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: "requests",
			Name:      "failed",
			Help:      "Number of failed API requests.",
		}),
		routesAdded: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: "routes",
			Name:      "added",
			Help:      "Number of routes added to RIB.",
		}),
		routesWithdrawn: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: "routes",
			Name:      "withdrawn",
			Help:      "Number of routes removed from RIB.",
		}),
	}
	return m
}

func (m *Metrics) RegisterEndpoint(addr string) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(m.requestsTotal)
	reg.MustRegister(m.requestsFailed)
	reg.MustRegister(m.routesAdded)
	reg.MustRegister(m.routesWithdrawn)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
	}))

	go func() {
		log.Infof("Start listening for metrics requests at http://%s/metrics", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
	}()
}
