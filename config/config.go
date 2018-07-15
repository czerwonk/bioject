package config

import (
	"fmt"
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config respresents the server configuration
type Config struct {
	// Our ASN
	LocalAS uint32 `yaml:"local_as"`

	// RouterID is the BGP router identifier of our server
	RouterID string `yaml:"router_id"`

	// Filters to match incoming (via API) routes against
	Filters []*RouteFilter `yaml:"route_filters"`

	// Sessions to BGP peers
	Sessions []*Session `yaml:"sessions"`
}

// Session defines all parameters needed to establish a BGP session with a peer
type Session struct {
	// Name of session
	Name string `yaml:"name"`

	// ASN of the peer
	RemoteAS uint32 `yaml:"remote_as"`

	// IP of the peer
	IP string `yaml:"ip"`
}

// RouteFilter defines all parameters needed to decide wether to accept or to drop a route for a prefix
type RouteFilter struct {
	// Net is the network address to match for
	Net string

	// Length is the prefix length
	Length uint8

	// Prefix length has to be larger or equal `Min`
	Min uint8

	// Prefix length has to be less or equal `Max`
	Max uint8
}

// Load loads a configuration from a reader
func Load(r io.Reader) (*Config, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("could not read config: %s", err)
	}

	c := &Config{}
	err = yaml.Unmarshal(b, c)
	if err != nil {
		return nil, fmt.Errorf("could not parse config: %s", err)
	}

	return c, nil
}