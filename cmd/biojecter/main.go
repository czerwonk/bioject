package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"

	pb "github.com/czerwonk/bioject/proto"
)

const (
	successCode = 200
	version     = "0.1"
)

func main() {
	apiAddress := flag.String("api", "[::1]:1337", "address to the bioject GRPPC API")
	prefix := flag.String("prefix", "", "prefix")
	nextHop := flag.String("next-hop", "", "next hop IP")
	community := flag.String("community", "", "community to tag the route with")
	withdraw := flag.Bool("withdraw", false, "withdraws route instead of adding it")
	v := flag.Bool("v", false, "Show version info")

	flag.Parse()

	if *v {
		showVersion()
		os.Exit(0)
	}

	conn, err := grpc.Dial(*apiAddress, grpc.WithInsecure())
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	client := pb.NewBioJectServiceClient(conn)
	err = sendRequest(client, *prefix, *nextHop, *community, *withdraw)
	if err != nil {
		log.Panic(err)
	}
}

func showVersion() {
	fmt.Println("biojecter - Simle client for bioject route injector")
	fmt.Println("Version:", version)
	fmt.Println("Author(s): Daniel Czerwonk")
}

func sendRequest(client pb.BioJectServiceClient, prefix, nextHop, community string, withdraw bool) error {
	pfx, err := parsePrefix(prefix)
	if err != nil {
		return err
	}

	nextHopIP := net.ParseIP(nextHop)
	if nextHopIP == nil {
		return fmt.Errorf("could not parse next hop IP address: %s", nextHop)
	}

	if withdraw {
		return sendWithdraw(client, pfx, nextHopIP)
	}

	return sendUpdate(client, pfx, nextHopIP, community)
}

func parsePrefix(s string) (*pb.Prefix, error) {
	ip, net, err := net.ParseCIDR(s)
	if err != nil {
		return nil, fmt.Errorf("could not parse prefix %v", err)
	}

	ones, _ := net.Mask.Size()

	return &pb.Prefix{
		Ip:     ip,
		Length: uint32(ones),
	}, nil
}

func sendUpdate(client pb.BioJectServiceClient, pfx *pb.Prefix, nextHop net.IP, community string) error {
	req := &pb.AddRouteRequest{
		Route: &pb.Route{
			Prefix:  pfx,
			NextHop: nextHop,
		},
		Communities:      make([]*pb.Community, 0),
		LargeCommunities: make([]*pb.LargeCommunity, 0),
	}

	res, err := client.AddRoute(context.Background(), req)
	if err != nil {
		return err
	}

	if res.Code != successCode {
		return fmt.Errorf("Error #%d: %s", res.Code, res.Message)
	}

	return nil
}

func sendWithdraw(client pb.BioJectServiceClient, prefix *pb.Prefix, nextHop net.IP) error {
	req := &pb.WithdrawRouteRequest{
		Route: &pb.Route{
			Prefix:  prefix,
			NextHop: nextHop,
		},
	}

	res, err := client.WithdrawRoute(context.Background(), req)
	if err != nil {
		return err
	}

	if res.Code != successCode {
		return fmt.Errorf("Error #%d: %s", res.Code, res.Message)
	}

	return nil
}
