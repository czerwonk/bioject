// SPDX-FileCopyrightText: (c) 2018 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package server

import (
	"context"

	"github.com/czerwonk/bioject/pkg/api"
	pb "github.com/czerwonk/bioject/proto"
	"go.opencensus.io/stats"
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
	stats.Record(ctx, m.metrics.requestsTotal.M(1))

	res, err := m.api.AddRoute(ctx, req)
	if err != nil || res.Code != api.StatusCodeOK {
		stats.Record(ctx, m.metrics.requestsFailed.M(1))
	}

	return res, err
}

func (m *metricAPIAdapter) WithdrawRoute(ctx context.Context, req *pb.WithdrawRouteRequest) (*pb.Result, error) {
	stats.Record(ctx, m.metrics.requestsTotal.M(1))

	res, err := m.api.WithdrawRoute(ctx, req)
	if err != nil || res.Code != api.StatusCodeOK {
		stats.Record(ctx, m.metrics.requestsFailed.M(1))
	}

	return res, err
}
