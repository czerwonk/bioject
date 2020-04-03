package server

import (
	"strconv"
	"strings"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/czerwonk/bioject/database"
)

func emptyASPath() *types.ASPath {
	p := make(types.ASPath, 0)
	return &p
}

func convertToDatabaseRoute(prefix *bnet.Prefix, path *route.Path) *database.Route {
	r := &database.Route{
		Prefix:           prefix.String(),
		NextHop:          path.BGPPath.BGPPathA.NextHop.String(),
		LocalPref:        uint(path.BGPPath.BGPPathA.LocalPref),
		MED:              uint(path.BGPPath.BGPPathA.MED),
		Communities:      communitiesFromBioRoute(path.BGPPath.Communities),
		LargeCommunities: largeCommunitiesFromBioRoute(path.BGPPath.LargeCommunities),
	}

	return r
}

func convertToBioRoute(r *database.Route) (pfx *bnet.Prefix, path *route.Path, err error) {
	t := strings.Split(r.Prefix, "/")
	net, err := bnet.IPFromString(t[0])
	if err != nil {
		return pfx, path, err
	}

	length, err := strconv.Atoi(t[1])
	if err != nil {
		return pfx, path, err
	}
	pfx = bnet.NewPfx(net, uint8(length)).Ptr()

	nextHop, err := bnet.IPFromString(r.NextHop)
	if err != nil {
		return &bnet.Prefix{}, path, err
	}

	return pfx, &route.Path{
		Type: route.BGPPathType,
		BGPPath: &route.BGPPath{
			BGPPathA: &route.BGPPathA{
				Source:    &bnet.IP{},
				LocalPref: uint32(r.LocalPref),
				MED:       uint32(r.MED),
				NextHop:   &nextHop,
				EBGP:      true,
			},
			Communities:      communitiesFromDatabaseRoute(r.Communities),
			LargeCommunities: largeCommunitiesFromDatabaseRoute(r.LargeCommunities),
			ASPath:           emptyASPath(),
		},
	}, nil
}

func communitiesFromDatabaseRoute(coms []*database.Community) *types.Communities {
	res := make(types.Communities, len(coms))

	for i, c := range coms {
		res[i] = uint32(c.ASN)<<16 + uint32(c.Value)
	}

	return &res
}

func largeCommunitiesFromDatabaseRoute(coms []*database.LargeCommunity) *types.LargeCommunities {
	res := make(types.LargeCommunities, len(coms))

	for i, c := range coms {
		res[i] = types.LargeCommunity{
			GlobalAdministrator: c.Global,
			DataPart1:           c.Data1,
			DataPart2:           c.Data2,
		}
	}

	return &res
}

func communitiesFromBioRoute(coms *types.Communities) []*database.Community {
	if coms == nil {
		return []*database.Community{}
	}

	res := make([]*database.Community, len(*coms))

	for i, c := range *coms {
		res[i] = &database.Community{
			ASN:   uint16((c & 0xFFFF0000) >> 16),
			Value: uint16(c & 0x0000FFFF),
		}
	}

	return res
}

func largeCommunitiesFromBioRoute(coms *types.LargeCommunities) []*database.LargeCommunity {
	if coms == nil {
		return []*database.LargeCommunity{}
	}

	res := make([]*database.LargeCommunity, len(*coms))

	for i, c := range *coms {
		res[i] = &database.LargeCommunity{
			Global: c.GlobalAdministrator,
			Data1:  c.DataPart1,
			Data2:  c.DataPart2,
		}
	}

	return res
}
