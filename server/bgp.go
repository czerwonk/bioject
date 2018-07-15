package server

import (
	"fmt"
	"net"
	"time"

	"github.com/bio-routing/bio-rd/routingtable/locRIB"

	bconfig "github.com/bio-routing/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
	bgp "github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/filter/actions"

	"github.com/czerwonk/bioject/config"
)

type bgpServer struct {
	rib *locRIB.LocRIB
}

func newBGPserver() *bgpServer {
	return &bgpServer{
		rib: locRIB.New(),
	}
}

func (bs *bgpServer) start(c *config.Config) error {
	b := bgp.NewBgpServer()

	routerID, err := bs.parseIP(c.RouterID)
	if err != nil {
		return fmt.Errorf("could not parse router id: %v", err)
	}

	err = b.Start(&bconfig.Global{
		Listen:   true,
		LocalASN: c.LocalAS,
		RouterID: routerID.ToUint32(),
	})
	if err != nil {
		return fmt.Errorf("unable to start BGP server: %v", err)
	}

	f, err := bs.exportFilter(c)
	if err != nil {
		return fmt.Errorf("could not create export filter from config: %v", err)
	}

	for _, sess := range c.Sessions {
		bs.addPeer(sess, f, b)
	}

	return nil
}

func (bs *bgpServer) exportFilter(c *config.Config) (*filter.Filter, error) {
	if len(c.Filters) == 0 {
		return filter.NewAcceptAllFilter(), nil
	}

	routeFilters := make([]*filter.RouteFilter, len(c.Filters))
	for i, f := range c.Filters {
		net, err := bs.parseIP(f.Net)
		if err != nil {
			return nil, err
		}

		pfx := bnet.NewPfx(net, f.Length)

		routeFilters[i] = filter.NewRouteFilter(pfx, filter.InRange(f.Min, f.Max))
	}

	terms := []*filter.Term{
		filter.NewTerm(
			[]*filter.TermCondition{
				filter.NewTermConditionWithRouteFilters(routeFilters...),
			},
			[]filter.FilterAction{
				&actions.AcceptAction{},
			}),
		filter.NewTerm(
			[]*filter.TermCondition{},
			[]filter.FilterAction{
				&actions.RejectAction{},
			}),
	}

	return filter.NewFilter(terms), nil
}

func (bs *bgpServer) addPeer(sess *config.Session, f *filter.Filter, b bgp.BGPServer) error {
	p, err := bs.peerForSession(sess)
	if err != nil {
		return err
	}

	b.AddPeer(p, bs.rib)
	return nil
}

func (bs *bgpServer) peerForSession(sess *config.Session) (bconfig.Peer, error) {
	ip, err := bs.parseIP(sess.IP)
	if err != nil {
		return bconfig.Peer{}, fmt.Errorf("could not parse IP for session %s: %v", sess.Name, err)
	}

	return bconfig.Peer{
		AdminEnabled:      true,
		PeerAS:            sess.RemoteAS,
		PeerAddress:       ip,
		ReconnectInterval: time.Second * 15,
		HoldTime:          time.Second * 90,
		KeepAlive:         time.Second * 30,
		Passive:           true,
		ImportFilter:      filter.NewDrainFilter(),
		IPv6:              !ip.IsIPv4(),
	}, nil
}

func (bs *bgpServer) parseIP(str string) (bnet.IP, error) {
	ip := net.ParseIP(str)
	if ip == nil {
		return bnet.IP{}, fmt.Errorf("%s is not a valid IP address", str)
	}

	ip4 := ip.To4()
	if ip4 != nil {
		return bnet.IPFromBytes(ip4)
	}

	return bnet.IPFromBytes(ip.To16())
}
