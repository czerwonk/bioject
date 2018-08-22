package server

import (
	"context"

	"github.com/czerwonk/bioject/api"
	pb "github.com/czerwonk/bioject/proto"
)

type metricAPIAdapter struct {
	api     *apiServer
	metrics *Metrics
}

func newMetricAPIAdapter(api *apiServer, metrics *Metrics) *metricAPIAdapter {
	return &metricAPIAdapter{
		api:     api,
		metrics: metrics,
	}
}

func (m *metricAPIAdapter) AddRoute(ctx context.Context, req *pb.AddRouteRequest) (*pb.Result, error) {
	m.metrics.requestsTotal.Inc()

	res, err := m.api.AddRoute(ctx, req)
	if err != nil || res.Code != api.StatusCodeOK {
		m.metrics.requestsFailed.Inc()
	}

	return res, err
}

func (m *metricAPIAdapter) WithdrawRoute(ctx context.Context, req *pb.WithdrawRouteRequest) (*pb.Result, error) {
	m.metrics.requestsTotal.Inc()

	res, err := m.api.WithdrawRoute(ctx, req)
	if err != nil || res.Code != api.StatusCodeOK {
		m.metrics.requestsFailed.Inc()
	}

	return res, err
}
