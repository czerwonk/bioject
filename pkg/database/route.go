// SPDX-FileCopyrightText: (c) 2018 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package database

type Route struct {
	ID               uint
	Prefix           string
	NextHop          string
	LocalPref        uint
	MED              uint
	Communities      []*Community
	LargeCommunities []*LargeCommunity
}

type Community struct {
	ID      uint
	ASN     uint16
	Value   uint16
	RouteID uint
}

type LargeCommunity struct {
	ID      uint
	Global  uint32
	Data1   uint32
	Data2   uint32
	RouteID uint
}

func NewRoute(prefix, nextHop string) *Route {
	return &Route{
		Prefix:           prefix,
		NextHop:          nextHop,
		Communities:      make([]*Community, 0),
		LargeCommunities: make([]*LargeCommunity, 0),
	}
}

func (r *Route) AddCommunity(asn uint16, value uint16) *Route {
	r.Communities = append(r.Communities, &Community{
		ASN:     asn,
		Value:   value,
		RouteID: r.ID,
	})

	return r
}

func (r *Route) AddLargeCommunity(global uint32, local1 uint32, local2 uint32) *Route {
	r.LargeCommunities = append(r.LargeCommunities, &LargeCommunity{
		Global:  global,
		Data1:   local1,
		Data2:   local2,
		RouteID: r.ID,
	})

	return r
}
