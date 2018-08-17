package server

import (
	"fmt"
	"time"

	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	log "github.com/sirupsen/logrus"

	bconfig "github.com/bio-routing/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
	bgp "github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/filter/actions"

	"github.com/czerwonk/bioject/config"
)

type bgpServer struct {
	rib     *locRIB.LocRIB
	metrics *Metrics
}

func newBGPserver(metrics *Metrics) *bgpServer {
	s := &bgpServer{
		rib:     locRIB.New(),
		metrics: metrics,
	}

	return s
}

func (bs *bgpServer) start(c *config.Config) error {
	b := bgp.NewBgpServer()

	routerID, err := bnet.IPFromString(c.RouterID)
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
		net, err := bnet.IPFromString(f.Net)
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
			[]filter.Action{
				&actions.AcceptAction{},
			}),
		filter.NewTerm(
			[]*filter.TermCondition{},
			[]filter.Action{
				&actions.RejectAction{},
			}),
	}

	return filter.NewFilter(terms), nil
}

func (bs *bgpServer) addPeer(sess *config.Session, f *filter.Filter, b bgp.BGPServer) error {
	p, err := bs.peerForSession(sess, f, b.RouterID())
	if err != nil {
		return err
	}

	b.AddPeer(p)
	return nil
}

func (bs *bgpServer) peerForSession(sess *config.Session, f *filter.Filter, routerID bnet.IP) (bconfig.Peer, error) {
	ip, err := bnet.IPFromString(sess.IP)
	if err != nil {
		return bconfig.Peer{}, fmt.Errorf("could not parse IP for session %s: %v", sess.Name, err)
	}

	p := bconfig.Peer{
		AdminEnabled:      true,
		PeerAS:            sess.RemoteAS,
		PeerAddress:       ip,
		ReconnectInterval: time.Second * 15,
		HoldTime:          time.Second * 90,
		KeepAlive:         time.Second * 30,
		Passive:           true,
		RouterID:          routerID,
	}

	addressFamily := &bconfig.AddressFamilyConfig{
		RIB:          bs.rib,
		ExportFilter: f,
		ImportFilter: filter.NewDrainFilter(),
	}

	if ip.IsIPv4() {
		p.IPv4 = addressFamily
	} else {
		p.IPv6 = addressFamily
	}

	return p, nil
}

func (bs *bgpServer) addPath(pfx bnet.Prefix, p *route.Path) error {
	if bs.rib.ContainsPfxPath(pfx, p) {
		return nil
	}

	err := bs.rib.AddPath(pfx, p)
	if err == nil {
		log.Infof("Added route: %s via %s\n", pfx, p.BGPPath.NextHop)
		bs.metrics.routesAdded.Inc()
	}

	return err
}

func (bs *bgpServer) removePath(pfx bnet.Prefix, p *route.Path) bool {
	if !bs.rib.ContainsPfxPath(pfx, p) {
		return true
	}

	res := bs.rib.RemovePath(pfx, p)
	if res {
		log.Infof("Removed route: %s via %s\n", pfx, p.BGPPath.NextHop)
		bs.metrics.routesWithdrawn.Inc()
	}

	return res
}
