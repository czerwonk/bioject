// SPDX-FileCopyrightText: (c) 2018 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		expected   *Config
		wantsFail  bool
	}{
		{
			name:       "config1",
			configFile: "config1.yml",
			expected: &Config{
				LocalAS:  65600,
				RouterID: "127.0.0.1",
				Filters: []*RouteFilter{
					{
						Net:    "2001:678:1e0::",
						Length: 48,
						Min:    127,
						Max:    128,
					},
				},
				Sessions: []*Session{
					{
						Name:                       "session1",
						RemoteAS:                   202739,
						LocalIP:                    "2001:678::1",
						PeerIP:                     "2001:678:1e0:b::1",
						Passive:                    true,
						AdvertiseIPv4MultiProtocol: true,
					},
				},
				Debug: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b, err := os.ReadFile("tests/" + test.configFile)
			if err != nil {
				t.Fatal(err)
			}

			r := bytes.NewReader(b)
			c, err := Load(r)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			assert.Equal(t, test.expected, c)
		})
	}
}
