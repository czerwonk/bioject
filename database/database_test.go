package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const testDB = "test.db"

func TestSave(t *testing.T) {
	route := &Route{
		Prefix:  "185.138.52.0/32",
		NextHop: "192.168.2.1",
	}

	d := connectTestDB()
	defer d.Close()

	d.Save(route)

	var r Route
	d.db.First(&r)
	if d.db.Error != nil {
		t.Fatal(d.db.Error)
	}

	assert.Equal(t, route.Prefix, r.Prefix)
	assert.Equal(t, route.NextHop, r.NextHop)
}

func TestDelete(t *testing.T) {
	d := connectTestDB()
	defer d.Close()

	r1 := insert(d, &Route{
		Prefix:  "185.138.52.1/32",
		NextHop: "192.168.2.1",
	}, t)

	insert(d, &Route{
		Prefix:  "185.138.52.0/32",
		NextHop: "192.168.2.1",
	}, t)

	r2 := insert(d, &Route{
		Prefix:  "185.138.52.0/32",
		NextHop: "185.138.53.1",
	}, t)

	route := &Route{
		Prefix:  "185.138.52.0/32",
		NextHop: "192.168.2.1",
	}

	if err := d.Delete(route); err != nil {
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

	r1 := insert(d, &Route{
		Prefix:  "185.138.52.1/32",
		NextHop: "192.168.2.1",
	}, t)

	r2 := insert(d, &Route{
		Prefix:  "185.138.52.0/32",
		NextHop: "192.168.2.1",
	}, t)

	routes, err := d.Routes()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 2, len(routes))
	assert.Equal(t, r1.Prefix, routes[0].Prefix, "first prefix")
	assert.Equal(t, r1.NextHop, routes[0].NextHop, "first nexthop")
	assert.Equal(t, r2.Prefix, routes[1].Prefix, "second prefix")
	assert.Equal(t, r2.NextHop, routes[1].NextHop, "second nexthop")
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

	return d
}
