package server

import (
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"testing"
	"time"

	"github.com/bio-routing/bio-rd/routingtable"

	bconfig "github.com/bio-routing/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/stretchr/testify/assert"

	"github.com/czerwonk/bioject/config"
)

func TestExportFilter(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		expectAccept []bnet.Prefix
		expectReject []bnet.Prefix
	}{
		{
			name: "2 route filters",
			config: &config.Config{
				Filters: []*config.RouteFilter{
					{
						Net:    "2001:678:1e0::",
						Length: 48,
						Min:    127,
						Max:    128,
					},
					{
						Net:    "192.168.0.0",
						Length: 24,
						Min:    30,
						Max:    32,
					},
				},
			},
			expectAccept: []bnet.Prefix{
				bnet.NewPfx(bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 1), 127),
				bnet.NewPfx(bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 1), 128),
				bnet.NewPfx(bnet.IPv4FromOctets(192, 168, 0, 4), 30),
				bnet.NewPfx(bnet.IPv4FromOctets(192, 168, 0, 0), 32),
			},
			expectReject: []bnet.Prefix{
				bnet.NewPfx(bnet.IPv4(0), 0),
				bnet.NewPfx(bnet.IPv6(0, 0), 0),
				bnet.NewPfx(bnet.IPv4FromOctets(127, 0, 0, 1), 8),
				bnet.NewPfx(bnet.IPv6(0, 1), 128),
				bnet.NewPfx(bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0x100, 0, 0, 0, 0), 56),
				bnet.NewPfx(bnet.IPv4FromOctets(192, 168, 0, 0), 29),
			},
		},
		{
			name: "no route filter",
			config: &config.Config{
				Filters: []*config.RouteFilter{},
			},
			expectAccept: []bnet.Prefix{
				bnet.NewPfx(bnet.IPv4(0), 0),
				bnet.NewPfx(bnet.IPv6(0, 0), 0),
				bnet.NewPfx(bnet.IPv4FromOctets(127, 0, 0, 1), 8),
				bnet.NewPfx(bnet.IPv6(0, 1), 128),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs := &bgpServer{}

			f, err := bs.exportFilter(test.config)
			if err != nil {
				t.Fatal(err)
			}

			for _, pfx := range test.expectAccept {
				if _, rejected := f.ProcessTerms(pfx, &route.Path{}); rejected {
					t.Fatalf("expected prefix %s to be accepted", pfx)
				}
			}

			for _, pfx := range test.expectReject {
				if _, rejected := f.ProcessTerms(pfx, &route.Path{}); !rejected {
					t.Fatalf("expected prefix %s to be rejected", pfx)
				}
			}
		})
	}
}

func TestPeerForSession(t *testing.T) {
	exportFilter := filter.NewDrainFilter()
	vrf, _ := vrf.New("master")

	routerID := bnet.IPv4FromOctets(127, 0, 0, 1).ToUint32()

	tests := []struct {
		name     string
		session  *config.Session
		expected bconfig.Peer
	}{
		{
			name: "IPv4 peer",
			session: &config.Session{
				Name:     "test",
				IP:       "192.168.1.1",
				RemoteAS: 65500,
			},
			expected: bconfig.Peer{
				AdminEnabled:      true,
				PeerAS:            65500,
				PeerAddress:       bnet.IPv4FromOctets(192, 168, 1, 1),
				ReconnectInterval: time.Second * 15,
				HoldTime:          time.Second * 90,
				KeepAlive:         time.Second * 30,
				RouterID:          routerID,
				IPv4: &bconfig.AddressFamilyConfig{
					ImportFilter: filter.NewDrainFilter(),
					ExportFilter: exportFilter,
					AddPathSend: routingtable.ClientOptions{
						BestOnly: true,
					},
				},
				VRF: vrf,
			},
		},
		{
			name: "IPv6 peer",
			session: &config.Session{
				Name:     "test",
				IP:       "2001:678:1e0::1",
				RemoteAS: 202739,
			},
			expected: bconfig.Peer{
				AdminEnabled:      true,
				PeerAS:            202739,
				PeerAddress:       bnet.IPv6FromBlocks(0x2001, 0x678, 0x1e0, 0, 0, 0, 0, 1),
				ReconnectInterval: time.Second * 15,
				HoldTime:          time.Second * 90,
				KeepAlive:         time.Second * 30,
				RouterID:          routerID,
				IPv6: &bconfig.AddressFamilyConfig{
					ImportFilter: filter.NewDrainFilter(),
					ExportFilter: exportFilter,
					AddPathSend: routingtable.ClientOptions{
						BestOnly: true,
					},
				},
				VRF: vrf,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bs := &bgpServer{
				vrf: vrf,
			}

			p, err := bs.peerForSession(test.session, exportFilter, routerID)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.expected, p)
		})
	}
}
