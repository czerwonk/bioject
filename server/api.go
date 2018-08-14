package server

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"

	"github.com/czerwonk/bioject/api"
	"github.com/czerwonk/bioject/database"
	pb "github.com/czerwonk/bioject/proto"

	log "github.com/sirupsen/logrus"
)

type bgpService interface {
	addPath(pfx bnet.Prefix, p *route.Path) error
	removePath(pfx bnet.Prefix, p *route.Path) bool
}

type apiServer struct {
	bgp bgpService
	db  *database.Database
}

func startAPIServer(listenAddress string, bgp bgpService, db *database.Database, metrics *Metrics) error {
	lis, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	api := &apiServer{
		bgp: bgp,
		db:  db,
	}

	s := grpc.NewServer()
	pb.RegisterBioJectServiceServer(s, newMetricAPIAdapter(api, metrics))
	reflection.Register(s)

	log.Println("Starting API server on", listenAddress)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func (s *apiServer) AddRoute(ctx context.Context, req *pb.AddRouteRequest) (*pb.Result, error) {
	log.Info("Received AddRoute request:", req)

	pfx, err := s.prefixForRequest(req.Route.Prefix)
	if err != nil {
		return s.errorResult(api.StatusCodeRequestError, err.Error()), nil
	}

	p, err := s.pathForRoute(req.Route)
	if err != nil {
		return s.errorResult(api.StatusCodeRequestError, err.Error()), nil
	}

	if err := s.bgp.addPath(pfx, p); err != nil {
		return s.errorResult(api.StatusCodeProcessingError, err.Error()), nil
	}

	if err := s.db.Save(convertToDatabaseRoute(pfx, p)); err != nil {
		return s.errorResult(api.StatusCodeProcessingError, err.Error()), nil
	}

	return &pb.Result{Code: api.StatusCodeOK}, nil
}

func (s *apiServer) WithdrawRoute(ctx context.Context, req *pb.WithdrawRouteRequest) (*pb.Result, error) {
	log.Info("Received WithdrawRoute request:", req)

	pfx, err := s.prefixForRequest(req.Route.Prefix)
	if err != nil {
		return s.errorResult(api.StatusCodeRequestError, err.Error()), nil
	}

	p, err := s.pathForRoute(req.Route)
	if err != nil {
		return s.errorResult(api.StatusCodeRequestError, err.Error()), nil
	}

	if !s.bgp.removePath(pfx, p) {
		return s.errorResult(api.StatusCodeProcessingError, "did not remove path"), nil
	}

	if err := s.db.Delete(convertToDatabaseRoute(pfx, p)); err != nil {
		return s.errorResult(api.StatusCodeProcessingError, err.Error()), nil
	}

	return &pb.Result{Code: api.StatusCodeOK}, nil
}

func (s *apiServer) pathForRoute(r *pb.Route) (*route.Path, error) {
	nextHopIP, err := bnet.IPFromBytes(r.NextHop)
	if err != nil {
		return nil, err
	}

	return &route.Path{
		Type: route.BGPPathType,
		BGPPath: &route.BGPPath{
			ASPath:    make(types.ASPath, 0),
			LocalPref: 100,
			NextHop:   nextHopIP,
			Origin:    0,
		},
	}, nil
}

func (s *apiServer) prefixForRequest(pfx *pb.Prefix) (bnet.Prefix, error) {
	ip, err := bnet.IPFromBytes(pfx.Ip)
	if err != nil {
		return bnet.Prefix{}, err
	}

	return bnet.NewPfx(ip, uint8(pfx.Length)), nil
}

func (s *apiServer) errorResult(code uint32, msg string) *pb.Result {
	log.Error("Error:", msg)
	return &pb.Result{
		Code:    code,
		Message: msg,
	}
}
