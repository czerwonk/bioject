package database

import (
	"github.com/jinzhu/gorm"
)

type Route struct {
	gorm.Model
	Prefix           string
	NextHop          string
	Communities      []*Community
	LargeCommunities []*LargeCommunity
}

type Community struct {
	gorm.Model
	ASN   uint16
	Value uint16
}

type LargeCommunity struct {
	Global uint32
	Data1  uint32
	Data2  uint32
}

func NewRoute(prefix, nextHop string) *Route {
	return &Route{
		Prefix:           prefix,
		NextHop:          nextHop,
		Communities:      make([]*Community, 0),
		LargeCommunities: make([]*LargeCommunity, 0),
	}
}

func (r *Route) AddCommunity(asn uint16, value uint16) {
	r.Communities = append(r.Communities, &Community{
		ASN:   asn,
		Value: value,
	})
}

func (r *Route) AddLargeCommunity(global uint32, local1 uint32, local2 uint32) {
	r.LargeCommunities = append(r.LargeCommunities, &LargeCommunity{
		Global: global,
		Data1:  local1,
		Data2:  local2,
	})
}
