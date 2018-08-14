package server

import (
	"strconv"
	"strings"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/czerwonk/bioject/database"
)

func convertToDatabaseRoute(prefix bnet.Prefix, path *route.Path) *database.Route {
	r := &database.Route{
		Prefix:           prefix.String(),
		NextHop:          path.BGPPath.NextHop.String(),
		Communities:      communitiesFromBioRoute(path.BGPPath.Communities),
		LargeCommunities: largeCommunitiesFromBioRoute(path.BGPPath.LargeCommunities),
	}

	return r
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
		Type: route.BGPPathType,
		BGPPath: &route.BGPPath{
			LocalPref:        100,
			NextHop:          nextHop,
			Communities:      communitiesFromDatabaseRoute(r.Communities),
			LargeCommunities: largeCommunitiesFromDatabaseRoute(r.LargeCommunities),
		},
	}, nil
}

func communitiesFromDatabaseRoute(coms []*database.Community) []uint32 {
	res := make([]uint32, len(coms))

	for i, c := range coms {
		res[i] = uint32(c.ASN)<<16 + uint32(c.Value)
	}

	return res
}

func largeCommunitiesFromDatabaseRoute(coms []*database.LargeCommunity) []types.LargeCommunity {
	res := make([]types.LargeCommunity, len(coms))

	for i, c := range coms {
		res[i] = types.LargeCommunity{
			GlobalAdministrator: c.Global,
			DataPart1:           c.Data1,
			DataPart2:           c.Data2,
		}
	}

	return res
}

func communitiesFromBioRoute(coms []uint32) []*database.Community {
	res := make([]*database.Community, len(coms))

	for i, c := range coms {
		res[i] = &database.Community{
			ASN:   uint16((c & 0xFFFF0000) >> 16),
			Value: uint16(c & 0x0000FFFF),
		}
	}

	return res
}

func largeCommunitiesFromBioRoute(coms []types.LargeCommunity) []*database.LargeCommunity {
	res := make([]*database.LargeCommunity, len(coms))

	for i, c := range coms {
		res[i] = &database.LargeCommunity{
			Global: c.GlobalAdministrator,
			Data1:  c.DataPart1,
			Data2:  c.DataPart2,
		}
	}

	return res
}
