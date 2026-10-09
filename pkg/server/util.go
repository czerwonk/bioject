// SPDX-FileCopyrightText: (c) 2018 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package server

import (
	"fmt"
	"strconv"
	"strings"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/czerwonk/bioject/pkg/database"
)

func emptyASPath() *types.ASPath {
	p := make(types.ASPath, 0)
	return &p
}

func convertToDatabaseRoute(prefix *bnet.Prefix, path *route.Path) *database.Route {
	r := &database.Route{
		Prefix:           prefix.String(),
		NextHop:          path.BGPPath.BGPPathA.NextHop.String(),
		LocalPref:        path.BGPPath.BGPPathA.LocalPref,
		MED:              path.BGPPath.BGPPathA.MED,
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

	length, err := strconv.ParseUint(t[1], 10, 8)
	if err != nil {
		return pfx, path, err
	}

	pfxLen, err := prefixLength(net, length)
	if err != nil {
		return pfx, path, err
	}
	pfx = bnet.NewPfx(net, pfxLen).Ptr()

	nextHop, err := bnet.IPFromString(r.NextHop)
	if err != nil {
		return &bnet.Prefix{}, path, err
	}

	return pfx, &route.Path{
		Type: route.BGPPathType,
		BGPPath: &route.BGPPath{
			BGPPathA: &route.BGPPathA{
				Source:    &bnet.IP{},
				LocalPref: r.LocalPref,
				MED:       r.MED,
				NextHop:   &nextHop,
				EBGP:      true,
			},
			Communities:      communitiesFromDatabaseRoute(r.Communities),
			LargeCommunities: largeCommunitiesFromDatabaseRoute(r.LargeCommunities),
			ASPath:           emptyASPath(),
		},
	}, nil
}

func prefixLength(ip bnet.IP, length uint64) (uint8, error) {
	if ip.IsIPv4() && length > 32 {
		return 0, fmt.Errorf("invalid IPv4 prefix length: %d", length)
	}

	if length > 128 {
		return 0, fmt.Errorf("invalid IPv6 prefix length: %d", length)
	}

	return uint8(length), nil
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
