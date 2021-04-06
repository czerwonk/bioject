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
