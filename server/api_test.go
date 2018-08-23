package server

import (
	"context"
	"fmt"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
	"github.com/bio-routing/bio-rd/route"
	"github.com/czerwonk/bioject/api"
	"github.com/czerwonk/bioject/database"
	pb "github.com/czerwonk/bioject/proto"
	"github.com/stretchr/testify/assert"
)

type bgpMock struct {
	addResult error
	addCalled bool

	removeResult bool
	removeCalled bool
}

func (m *bgpMock) addPath(pfx bnet.Prefix, p *route.Path) error {
	m.addCalled = true
	return m.addResult
}

func (m *bgpMock) removePath(pfx bnet.Prefix, p *route.Path) bool {
	m.removeCalled = true
	return m.removeResult
}

type dbMock struct {
	saveCalled   bool
	deleteCalled bool
}

func (m *dbMock) Save(route *database.Route) error {
	m.saveCalled = true
	return nil
}

func (m *dbMock) Delete(route *database.Route) error {
	m.deleteCalled = true
	return nil
}

func TestAddRoute(t *testing.T) {
	tests := []struct {
		name         string
		req          *pb.AddRouteRequest
		addResult    error
		expectedCode uint32
		wantBGPCall  bool
		wantDBCall   bool
		wantFail     bool
	}{
		{
			name: "valid route IPv4",
			req: &pb.AddRouteRequest{
				Route: &pb.Route{
					Prefix: &pb.Prefix{
						Ip:     []byte{194, 48, 228, 0},
						Length: 24,
					},
					NextHop:   []byte{192, 168, 2, 1},
					LocalPref: 200,
					Med:       1,
				},
			},
			wantBGPCall:  true,
			wantDBCall:   true,
			expectedCode: api.StatusCodeOK,
		},
		{
			name: "valid route IPv6",
			req: &pb.AddRouteRequest{
				Route: &pb.Route{
					Prefix: &pb.Prefix{
						Ip:     []byte{0x20, 0x01, 0x06, 0x78, 0x01, 0xe0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						Length: 48,
					},
					NextHop: []byte{0x20, 0x01, 0x06, 0x78, 0x01, 0xe0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff},
				},
			},
			wantBGPCall:  true,
			wantDBCall:   true,
			expectedCode: api.StatusCodeOK,
		},
		{
			name: "invalid prefix",
			req: &pb.AddRouteRequest{
				Route: &pb.Route{
					Prefix: &pb.Prefix{
						Ip:     []byte{194, 48},
						Length: 24,
					},
					NextHop:   []byte{192, 168, 2, 1},
					LocalPref: 200,
					Med:       1,
				},
			},
			wantFail:     true,
			expectedCode: api.StatusCodeRequestError,
		},
		{
			name: "invalid next hop",
			req: &pb.AddRouteRequest{
				Route: &pb.Route{
					Prefix: &pb.Prefix{
						Ip:     []byte{194, 48, 228, 0},
						Length: 24,
					},
					NextHop:   []byte{192, 168},
					LocalPref: 200,
					Med:       1,
				},
			},
			wantFail:     true,
			expectedCode: api.StatusCodeRequestError,
		},
		{
			name: "error on add",
			req: &pb.AddRouteRequest{
				Route: &pb.Route{
					Prefix: &pb.Prefix{
						Ip:     []byte{194, 48, 228, 0},
						Length: 24,
					},
					NextHop:   []byte{192, 168, 2, 1},
					LocalPref: 200,
					Med:       1,
				},
			},
			wantFail:     true,
			wantBGPCall:  true,
			wantDBCall:   false,
			addResult:    fmt.Errorf("test"),
			expectedCode: api.StatusCodeProcessingError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := &bgpMock{
				addResult: test.addResult,
			}

			db := &dbMock{}

			api := &apiServer{
				bgp: b,
				db:  db,
			}

			res, err := api.AddRoute(context.Background(), test.req)
			if err != nil {
				assert.True(t, test.wantFail, "unexpected error:", err)
				return
			}

			assert.Equal(t, test.wantBGPCall, b.addCalled, "add called on BGP")
			assert.Equal(t, test.wantDBCall, db.saveCalled, "save called on DB")
			assert.Equal(t, test.expectedCode, res.Code, "code")
		})
	}
}

func TestWithdrawRoute(t *testing.T) {
	tests := []struct {
		name         string
		req          *pb.WithdrawRouteRequest
		removeResult bool
		expectedCode uint32
		wantBGPCall  bool
		wantDBCall   bool
		wantFail     bool
	}{
		{
			name: "valid route IPv4",
			req: &pb.WithdrawRouteRequest{
				Route: &pb.Route{
					Prefix: &pb.Prefix{
						Ip:     []byte{194, 48, 228, 0},
						Length: 24,
					},
					NextHop: []byte{192, 168, 2, 1},
				},
			},
			wantBGPCall:  true,
			wantDBCall:   true,
			removeResult: true,
			expectedCode: api.StatusCodeOK,
		},
		{
			name: "valid route IPv6",
			req: &pb.WithdrawRouteRequest{
				Route: &pb.Route{
					Prefix: &pb.Prefix{
						Ip:     []byte{0x20, 0x01, 0x06, 0x78, 0x01, 0xe0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
						Length: 48,
					},
					NextHop: []byte{0x20, 0x01, 0x06, 0x78, 0x01, 0xe0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff},
				},
			},
			wantBGPCall:  true,
			wantDBCall:   true,
			removeResult: true,
			expectedCode: api.StatusCodeOK,
		},
		{
			name: "invalid prefix",
			req: &pb.WithdrawRouteRequest{
				Route: &pb.Route{
					Prefix: &pb.Prefix{
						Ip:     []byte{194, 48},
						Length: 24,
					},
					NextHop: []byte{192, 168, 2, 1},
				},
			},
			wantFail:     true,
			expectedCode: api.StatusCodeRequestError,
		},
		{
			name: "invalid next hop",
			req: &pb.WithdrawRouteRequest{
				Route: &pb.Route{
					Prefix: &pb.Prefix{
						Ip:     []byte{194, 48, 228, 0},
						Length: 24,
					},
					NextHop: []byte{192, 168},
				},
			},
			wantFail:     true,
			expectedCode: api.StatusCodeRequestError,
		},
		{
			name: "error on add",
			req: &pb.WithdrawRouteRequest{
				Route: &pb.Route{
					Prefix: &pb.Prefix{
						Ip:     []byte{194, 48, 228, 0},
						Length: 24,
					},
					NextHop: []byte{192, 168, 2, 1},
				},
			},
			wantFail:     true,
			wantBGPCall:  true,
			wantDBCall:   false,
			removeResult: false,
			expectedCode: api.StatusCodeProcessingError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := &bgpMock{
				removeResult: test.removeResult,
			}

			db := &dbMock{}

			api := &apiServer{
				bgp: b,
				db:  db,
			}

			res, err := api.WithdrawRoute(context.Background(), test.req)
			if err != nil {
				assert.True(t, test.wantFail, "unexpected error:", err)
				return
			}

			assert.Equal(t, test.wantBGPCall, b.removeCalled, "remove called on BGP")
			assert.Equal(t, test.wantDBCall, db.deleteCalled, "delete called on DB")
			assert.Equal(t, test.expectedCode, res.Code, "code")
		})
	}
}

func TestPathForRoute(t *testing.T) {
	tests := []struct {
		name     string
		route    *pb.Route
		expected *route.BGPPath
		wantFail bool
	}{
		{
			name: "valid path with IPv4 nexthop",
			route: &pb.Route{
				NextHop:   []byte{192, 168, 2, 1},
				LocalPref: 200,
				Med:       1,
			},
			expected: &route.BGPPath{
				ASPath:    make(types.ASPath, 0),
				EBGP:      true,
				LocalPref: 200,
				MED:       1,
				NextHop:   bnet.IPv4FromOctets(192, 168, 2, 1),
			},
		},
		{
			name: "valid path with IPv6 nexthop",
			route: &pb.Route{
				NextHop:   []byte{0x20, 0x01, 0x06, 0x78, 0x01, 0xe0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
				LocalPref: 200,
				Med:       1,
			},
			expected: &route.BGPPath{
				ASPath:    make(types.ASPath, 0),
				EBGP:      true,
				LocalPref: 200,
				MED:       1,
				NextHop:   bnet.IPv6FromBlocks(0x2001, 0x0678, 0x01e0, 0, 0, 0, 0, 1),
			},
		},
		{
			name: "invalid nexthop",
			route: &pb.Route{
				NextHop:   []byte{65, 66, 67},
				LocalPref: 200,
				Med:       1,
			},
			wantFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &apiServer{}
			path, err := s.pathForRoute(test.route)
			if err != nil {
				assert.True(t, test.wantFail, "unexpected error:", err)
				return
			}

			assert.Equal(t, test.expected, path.BGPPath)
		})
	}
}
