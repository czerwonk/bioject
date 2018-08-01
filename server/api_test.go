package server

import (
	"context"
	"fmt"
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route"
	"github.com/czerwonk/bioject/api"
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

func TestAddRoute(t *testing.T) {
	tests := []struct {
		name         string
		req          *pb.AddRouteRequest
		addResult    error
		expectedCode uint32
		wantCall     bool
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
					NextHop: []byte{192, 168, 2, 1},
				},
			},
			wantCall:     true,
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
			wantCall:     true,
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
					NextHop: []byte{192, 168, 2, 1},
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
					NextHop: []byte{192, 168},
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
					NextHop: []byte{192, 168, 2, 1},
				},
			},
			wantFail:     true,
			wantCall:     true,
			addResult:    fmt.Errorf("test"),
			expectedCode: api.StatusCodeProcessingError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := &bgpMock{
				addResult: test.addResult,
			}

			api := &apiServer{bgp: m}

			res, err := api.AddRoute(context.Background(), test.req)
			if err != nil {
				assert.True(t, test.wantFail, "unexpected error:", err)
				return
			}

			assert.Equal(t, test.wantCall, m.addCalled, "add called")
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
		wantCall     bool
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
			wantCall:     true,
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
			wantCall:     true,
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
			wantCall:     true,
			removeResult: false,
			expectedCode: api.StatusCodeProcessingError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := &bgpMock{
				removeResult: test.removeResult,
			}

			api := &apiServer{bgp: m}

			res, err := api.WithdrawRoute(context.Background(), test.req)
			if err != nil {
				assert.True(t, test.wantFail, "unexpected error:", err)
				return
			}

			assert.Equal(t, test.wantCall, m.removeCalled, "remove called")
			assert.Equal(t, test.expectedCode, res.Code, "code")
		})
	}
}
