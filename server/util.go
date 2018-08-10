package server

import (
	"strconv"
	"strings"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/czerwonk/bioject/database"
)

func convertToDatabaseRoute(prefix bnet.Prefix, path *route.Path) *database.Route {
	return &database.Route{
		Prefix:  prefix.String(),
		NextHop: path.BGPPath.NextHop.String(),
	}
}

func convertToBioRoute(r *database.Route) (pfx bnet.Prefix, path *route.Path, err error) {
	t := strings.Split(r.Prefix, "/")
	net, err := bnet.IPFromString(t[0])
	if err != nil {
		return pfx, path, err
	}

	length, err := strconv.Atoi(t[1])
	if err != nil {
		return pfx, path, err
	}
	pfx = bnet.NewPfx(net, uint8(length))

	nextHop, err := bnet.IPFromString(r.NextHop)
	if err != nil {
		return bnet.Prefix{}, path, err
	}

	return pfx, &route.Path{
		BGPPath: &route.BGPPath{
			LocalPref: 100,
			NextHop:   nextHop,
		},
	}, nil
}
