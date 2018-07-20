package server

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/czerwonk/bioject/proto"

	log "github.com/sirupsen/logrus"
)

type apiServer struct {
}

func startAPIServer(listenAddress string) error {
	lis, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterBioJectServiceServer(s, &apiServer{})

	reflection.Register(s)

	log.Println("Starting API server on", listenAddress)
	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

func (s *apiServer) AddRoute(ctx context.Context, req *pb.AddRouteRequest) (*pb.Result, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *apiServer) WithdrawRoute(ctx context.Context, req *pb.WithdrawRouteRequest) (*pb.Result, error) {
	return nil, fmt.Errorf("not implemented")
}
