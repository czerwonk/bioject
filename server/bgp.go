package server

import (
	"context"
	"github.com/bio-routing/bio-rd/routingtable/locRIB"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/pkg/errors"
	"net"
	"time"

	bconfig "github.com/bio-routing/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
	bgp "github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/filter/actions"
	"github.com/czerwonk/bioject/config"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.opencensus.io/trace"
)

type bgpServer struct {
	vrf           *vrf.VRF
	metrics       *Metrics
	listenAddress net.IP
}

func newBGPserver(metrics *Metrics, listenAddress net.IP) *bgpServer {
	v, _ := vrf.New("master")

	s := &bgpServer{
		vrf:           v,
		metrics:       metrics,
		listenAddress: listenAddress,
	}

	return s
}

func (bs *bgpServer) start(c *config.Config) error {
	b := bgp.NewBgpServer()

	routerID, err := bnet.IPFromString(c.RouterID)
	if err != nil {
		return errors.Wrap(err, "could not parse router id")
	}

	err = b.Start(&bconfig.Global{
		Listen:           true,
		LocalASN:         c.LocalAS,
		RouterID:         routerID.ToUint32(),
		LocalAddressList: []net.IP{bs.listenAddress},
	})
	if err != nil {
		return errors.Wrap(err, "unable to start BGP server")
	}

	f, err := bs.exportFilter(c)
	if err != nil {
		return errors.Wrap(err, "could not create export filter from config")
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

func (bs *bgpServer) peerForSession(sess *config.Session, f *filter.Filter, routerID uint32) (bconfig.Peer, error) {
	ip, err := bnet.IPFromString(sess.IP)
	if err != nil {
		return bconfig.Peer{}, errors.Wrapf(err, "could not parse IP for session %s", sess.Name)
	}

	p := bconfig.Peer{
		AdminEnabled:      true,
		PeerAS:            sess.RemoteAS,
		PeerAddress:       ip,
		ReconnectInterval: time.Second * 15,
		HoldTime:          time.Second * 90,
		KeepAlive:         time.Second * 30,
		Passive:           sess.Passive,
		RouterID:          routerID,
		VRF:               bs.vrf,
	}

	addressFamily := &bconfig.AddressFamilyConfig{
		ExportFilter: f,
		ImportFilter: filter.NewDrainFilter(),
		AddPathSend: routingtable.ClientOptions{
			BestOnly: true,
		},
		AddPathRecv: false,
	}

	if ip.IsIPv4() {
		p.IPv4 = addressFamily
	} else {
		p.IPv6 = addressFamily
	}

	return p, nil
}

func (bs *bgpServer) addPath(ctx context.Context, pfx bnet.Prefix, p *route.Path) error {
	ctx, span := trace.StartSpan(ctx, "BGP.AddPath")
	defer span.End()

	rib := bs.ribForPrefix(pfx)

	if rib.ContainsPfxPath(pfx, p) {
		return nil
	}

	err := rib.AddPath(pfx, p)
	if err == nil {
		log.Infof("Added route: %s via %s\n", pfx, p.BGPPath.NextHop)
		stats.Record(ctx, bs.metrics.routesAdded.M(1))
	}

	return err
}

func (bs *bgpServer) removePath(ctx context.Context, pfx bnet.Prefix, p *route.Path) bool {
	ctx, span := trace.StartSpan(ctx, "BGP.RemovePath")
	defer span.End()

	rib := bs.ribForPrefix(pfx)

	if !rib.ContainsPfxPath(pfx, p) {
		return true
	}

	res := rib.RemovePath(pfx, p)
	if res {
		log.Infof("Removed route: %s via %s\n", pfx, p.BGPPath.NextHop)
		stats.Record(ctx, bs.metrics.routesWithdrawn.M(1))
	}

	return res
}

func (bs *bgpServer) ribForPrefix(pfx bnet.Prefix) *locRIB.LocRIB {
	if pfx.Addr().IsIPv4() {
		return bs.vrf.IPv4UnicastRIB()
	}

	return bs.vrf.IPv6UnicastRIB()
}
