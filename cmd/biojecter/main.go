package main

import (
	"context"
	"flag"
	"fmt"
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

	conn, err := grpc.Dial(*apiAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

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

	return &pb.Prefix{
		Ip:     ipBytes(ip),
		Length: uint32(ones),
	}, nil
}

func ipBytes(ip net.IP) []byte {
	b := ip.To4()
	if b == nil {
		b = ip.To16()
	}

	return b
}

func sendUpdate(client pb.BioJectServiceClient, pfx *pb.Prefix, nextHop net.IP, p *requestParameters) error {
	req := createAddRouteRequest(pfx, nextHop, p)

	res, err := client.AddRoute(context.Background(), req)
	if err != nil {
		return err
	}

	if res.Code != api.StatusCodeOK {
		return fmt.Errorf("error #%d: %s", res.Code, res.Message)
	}

	return nil
}

func createAddRouteRequest(pfx *pb.Prefix, nextHop net.IP, p *requestParameters) *pb.AddRouteRequest {
	req := &pb.AddRouteRequest{
		Route: &pb.Route{
			Prefix:    pfx,
			NextHop:   nextHop,
			LocalPref: uint32(p.localPref),
			Med:       uint32(p.med),
		},
		Communities:      make([]*pb.Community, 0),
		LargeCommunities: make([]*pb.LargeCommunity, 0),
	}

	matches := communityRegex.FindAllStringSubmatch(p.community, -1)
	for _, m := range matches {
		if m[3] != "" {
			req.LargeCommunities = append(req.LargeCommunities, largeCommunityForMatch(m))
		} else {
			req.Communities = append(req.Communities, communityForMatch(m))
		}
	}

	return req
}

func largeCommunityForMatch(groups []string) *pb.LargeCommunity {
	global, _ := strconv.Atoi(groups[1])
	p1, _ := strconv.Atoi(groups[2])
	p2, _ := strconv.Atoi(groups[3])

	return &pb.LargeCommunity{
		GlobalAdministrator: uint32(global),
		LocalDataPart1:      uint32(p1),
		LocalDataPart2:      uint32(p2),
	}
}

func communityForMatch(groups []string) *pb.Community {
	asn, _ := strconv.Atoi(groups[1])
	value, _ := strconv.Atoi(groups[2])

	return &pb.Community{
		Asn:   uint32(asn),
		Value: uint32(value),
	}
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
