// SPDX-FileCopyrightText: (c) 2018 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package server

import (
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/stretchr/testify/assert"

	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/czerwonk/bioject/pkg/database"
)

func TestConvertToBioRoute(t *testing.T) {
	r := database.NewRoute("185.138.52.0/32", "192.168.2.1")
	r.LocalPref = 100
	r.MED = 1
	r.AddCommunity(48821, 123)
	r.AddLargeCommunity(202739, 123, 456)

	expectedPrefix := bnet.NewPfx(bnet.IPv4FromOctets(185, 138, 52, 0), 32)
	expectedPath := &route.Path{
		Type: route.BGPPathType,
		BGPPath: &route.BGPPath{
			BGPPathA: &route.BGPPathA{
				LocalPref: 100,
				MED:       1,
				NextHop:   bnet.IPv4FromOctets(192, 168, 2, 1).Ptr(),
				Source:    &bnet.IP{},
				EBGP:      true,
			},
			Communities: &types.Communities{
				3199533179,
			},
			LargeCommunities: &types.LargeCommunities{
				{
					GlobalAdministrator: 202739,
					DataPart1:           123,
					DataPart2:           456,
				},
			},
			ASPath: emptyASPath(),
		},
	}

	pfx, p, err := convertToBioRoute(r)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, &expectedPrefix, pfx, "Prefix")
	assert.Equal(t, expectedPath, p, "Path")
}

func TestConvertToDatabaseRoute(t *testing.T) {
	pfx := bnet.NewPfx(bnet.IPv4FromOctets(185, 138, 53, 0), 32)
	path := &route.Path{
		Type: route.BGPPathType,
		BGPPath: &route.BGPPath{
			BGPPathA: &route.BGPPathA{
				LocalPref: 200,
				MED:       1,
				NextHop:   bnet.IPv4FromOctets(192, 168, 2, 1).Ptr(),
				Source:    &bnet.IP{},
			},
			Communities: &types.Communities{
				3199533179,
			},
			LargeCommunities: &types.LargeCommunities{
				{
					GlobalAdministrator: 202739,
					DataPart1:           123,
					DataPart2:           456,
				},
			},
		},
	}

	expected := database.NewRoute("185.138.53.0/32", "192.168.2.1")
	expected.AddCommunity(48821, 123)
	expected.AddLargeCommunity(202739, 123, 456)
	expected.LocalPref = 200
	expected.MED = 1

	r := convertToDatabaseRoute(&pfx, path)

	assert.Equal(t, expected, r)
}

func TestConvertToBioRouteInvalidPrefixLength(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{name: "IPv4 length too large", prefix: "185.138.52.0/33"},
		{name: "IPv6 length too large", prefix: "2001:678:1e0::/129"},
		{name: "length overflowing uint8", prefix: "185.138.52.0/280"},
		{name: "negative length", prefix: "185.138.52.0/-1"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := database.NewRoute(test.prefix, "192.168.2.1")
			_, _, err := convertToBioRoute(r)
			assert.Error(t, err, "expected error for prefix %s", test.prefix)
		})
	}
}
