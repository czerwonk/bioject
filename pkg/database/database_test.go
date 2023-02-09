// SPDX-FileCopyrightText: (c) 2018 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package database

import (
	"context"
	"os"
	"testing"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
)

const testDB = "test.db"

func TestSave(t *testing.T) {
	route := &Route{
		Prefix:    "185.138.52.0/32",
		NextHop:   "192.168.2.1",
		LocalPref: 200,
		MED:       1,
		Communities: []*Community{
			{
				ASN:   48821,
				Value: 6500,
			},
		},
		LargeCommunities: []*LargeCommunity{
			{
				Global: 202739,
				Data1:  456,
				Data2:  789,
			},
			{
				Global: 48821,
				Data1:  689,
				Data2:  234,
			},
		},
	}

	d := connectTestDB()
	defer d.Close()

	d.Save(context.Background(), route)

	var r Route
	d.db.Preload("Communities").Preload("LargeCommunities").First(&r)
	if d.db.Error != nil {
		t.Fatal(d.db.Error)
	}

	assert.Equal(t, route.Prefix, r.Prefix, "Prefix")
	assert.Equal(t, route.NextHop, r.NextHop, "Next-Hop")
	assert.Equal(t, route.LocalPref, r.LocalPref, "Local-Pref")
	assert.Equal(t, route.MED, r.MED, "MED")
	assert.Equal(t, 1, len(r.Communities), "Communities")
	assert.Equal(t, 2, len(r.LargeCommunities), "Large Communities")
}

func TestDelete(t *testing.T) {
	d := connectTestDB()
	defer d.Close()

	r1 := insert(d, NewRoute("185.138.52.1/32", "192.168.2.1"), t)
	insert(d, NewRoute("185.138.52.0/32", "192.168.2.1"), t)
	r2 := insert(d, NewRoute("185.138.52.0/32", "185.138.53.1"), t)

	route := NewRoute("185.138.52.0/32", "192.168.2.1")
	if err := d.Delete(context.Background(), route); err != nil {
		t.Fatal(err)
	}

	var routes []*Route
	d.db.Find(&routes)
	if d.db.Error != nil {
		t.Fatal(d.db.Error)
	}

	assert.Equal(t, 2, len(routes))
	assert.Equal(t, r1.Prefix, routes[0].Prefix, "first prefix")
	assert.Equal(t, r1.NextHop, routes[0].NextHop, "first nexthop")
	assert.Equal(t, r2.Prefix, routes[1].Prefix, "second prefix")
	assert.Equal(t, r2.NextHop, routes[1].NextHop, "second nexthop")
}

func TestRoutes(t *testing.T) {
	d := connectTestDB()
	defer d.Close()

	r1 := insert(d, NewRoute("185.138.52.1/32", "192.168.2.1").AddLargeCommunity(202739, 123, 456), t)
	r2 := insert(d, NewRoute("185.138.52.0/32", "192.168.2.1").AddCommunity(48821, 123), t)

	routes, err := d.Routes()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(routes))

	assert.Equal(t, r1.Prefix, routes[0].Prefix, "route 1 prefix")
	assert.Equal(t, r1.NextHop, routes[0].NextHop, "route 1 next-hop")
	assert.Equal(t, 0, len(routes[0].Communities), "route 1 communities")
	assert.Equal(t, 1, len(routes[0].LargeCommunities), "route 1 large communities")

	assert.Equal(t, r2.Prefix, routes[1].Prefix, "route 2 prefix")
	assert.Equal(t, r2.NextHop, routes[1].NextHop, "route 2 next-hop")
	assert.Equal(t, 1, len(routes[1].Communities), "route 2 communities")
	assert.Equal(t, 0, len(routes[1].LargeCommunities), "route 2 large communities")
}

func insert(d *Database, r *Route, t *testing.T) *Route {
	d.db.Save(r)
	if d.db.Error != nil {
		t.Fatal(d.db.Error)
	}

	return r
}

func connectTestDB() *Database {
	stat, _ := os.Stat(testDB)
	if stat != nil {
		if err := os.Remove(testDB); err != nil {
			panic(err)
		}
	}

	d, err := Connect("sqlite3", testDB)
	if err != nil {
		panic(err)
	}

	d.db.LogMode(true)

	return d
}
