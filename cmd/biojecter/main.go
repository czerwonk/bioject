// SPDX-FileCopyrightText: (c) 2018 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"net"
	"os"
	"regexp"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	log "github.com/sirupsen/logrus"

	"github.com/czerwonk/bioject/pkg/api"
	pb "github.com/czerwonk/bioject/proto"
)

const version = "0.1.4"

var (
	communityRegex = regexp.MustCompile(`(\d+)\:(\d+)(?:\:(\d+))?`)
)

type requestParameters struct {
	prefix    string
	nextHop   string
	localPref int
	med       int
	community string
	withdraw  bool
}

func main() {
	p := &requestParameters{}
	apiAddress := flag.String("api", "[::1]:1337", "Address to the bioject GRPC API")
	flag.StringVar(&p.prefix, "prefix", "", "Prefix")
	flag.StringVar(&p.nextHop, "next-hop", "", "Next hop IP")
	flag.IntVar(&p.localPref, "local-pref", 100, "Local preference of the route")
	flag.IntVar(&p.med, "med", 0, "Multiple Exit Discriminator of the route")
	flag.StringVar(&p.community, "community", "", "BGP Community to tag the route with (Format: a:b for RFC1997 or a:b:c for RFC8195)")
	flag.BoolVar(&p.withdraw, "withdraw", false, "Withdraws route instead of adding it")
	v := flag.Bool("v", false, "Show version info")

	flag.Parse()

	if *v {
		showVersion()
		os.Exit(0)
	}

	conn, err := grpc.NewClient(*apiAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Panic(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Errorf("could not close connection: %v", err)
		}
	}()

	client := pb.NewBioJectServiceClient(conn)
	err = sendRequest(client, p)
	if err != nil {
		log.Panic(err)
	}
}

func showVersion() {
	fmt.Println("biojecter - Simple client for bioject route injector")
	fmt.Println("Version:", version)
	fmt.Println("Author(s): Daniel Czerwonk")
}

func sendRequest(client pb.BioJectServiceClient, p *requestParameters) error {
	pfx, err := parsePrefix(p.prefix)
	if err != nil {
		return err
	}

	nextHopIP := net.ParseIP(p.nextHop)
	if nextHopIP == nil {
		return fmt.Errorf("could not parse next hop IP address: %s", p.nextHop)
	}

	if p.withdraw {
		return sendWithdraw(client, pfx, ipBytes(nextHopIP))
	}

	return sendUpdate(client, pfx, ipBytes(nextHopIP), p)
}

func parsePrefix(s string) (*pb.Prefix, error) {
	ip, net, err := net.ParseCIDR(s)
	if err != nil {
		return nil, fmt.Errorf("could not parse prefix %v", err)
	}

	ones, _ := net.Mask.Size()
	length, err := toUint32(ones, "prefix length")
	if err != nil {
		return nil, err
	}

	return &pb.Prefix{
		Ip:     ipBytes(ip),
		Length: length,
	}, nil
}

func toUint32(v int, name string) (uint32, error) {
	if v < 0 || v > math.MaxUint32 {
		return 0, fmt.Errorf("%s out of range: %d", name, v)
	}

	return uint32(v), nil
}

func ipBytes(ip net.IP) []byte {
	b := ip.To4()
	if b == nil {
		b = ip.To16()
	}

	return b
}

func sendUpdate(client pb.BioJectServiceClient, pfx *pb.Prefix, nextHop net.IP, p *requestParameters) error {
	req, err := createAddRouteRequest(pfx, nextHop, p)
	if err != nil {
		return err
	}

	res, err := client.AddRoute(context.Background(), req)
	if err != nil {
		return err
	}

	if res.Code != api.StatusCodeOK {
		return fmt.Errorf("error #%d: %s", res.Code, res.Message)
	}

	return nil
}

func createAddRouteRequest(pfx *pb.Prefix, nextHop net.IP, p *requestParameters) (*pb.AddRouteRequest, error) {
	localPref, err := toUint32(p.localPref, "local pref")
	if err != nil {
		return nil, err
	}

	med, err := toUint32(p.med, "MED")
	if err != nil {
		return nil, err
	}

	req := &pb.AddRouteRequest{
		Route: &pb.Route{
			Prefix:    pfx,
			NextHop:   nextHop,
			LocalPref: localPref,
			Med:       med,
		},
		Communities:      make([]*pb.Community, 0),
		LargeCommunities: make([]*pb.LargeCommunity, 0),
	}

	matches := communityRegex.FindAllStringSubmatch(p.community, -1)
	for _, m := range matches {
		if m[3] != "" {
			c, err := largeCommunityForMatch(m)
			if err != nil {
				return nil, err
			}
			req.LargeCommunities = append(req.LargeCommunities, c)
		} else {
			c, err := communityForMatch(m)
			if err != nil {
				return nil, err
			}
			req.Communities = append(req.Communities, c)
		}
	}

	return req, nil
}

func parseUint32(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("could not parse community part %q: %w", s, err)
	}

	return uint32(v), nil
}

func largeCommunityForMatch(groups []string) (*pb.LargeCommunity, error) {
	global, err := parseUint32(groups[1])
	if err != nil {
		return nil, err
	}

	p1, err := parseUint32(groups[2])
	if err != nil {
		return nil, err
	}

	p2, err := parseUint32(groups[3])
	if err != nil {
		return nil, err
	}

	return &pb.LargeCommunity{
		GlobalAdministrator: global,
		LocalDataPart1:      p1,
		LocalDataPart2:      p2,
	}, nil
}

func communityForMatch(groups []string) (*pb.Community, error) {
	asn, err := parseUint32(groups[1])
	if err != nil {
		return nil, err
	}

	value, err := parseUint32(groups[2])
	if err != nil {
		return nil, err
	}

	return &pb.Community{
		Asn:   asn,
		Value: value,
	}, nil
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

	if res.Code != api.StatusCodeOK {
		return fmt.Errorf("error #%d: %s", res.Code, res.Message)
	}

	return nil
}
