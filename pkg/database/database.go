package database

import (
	"context"

	"github.com/jinzhu/gorm"
	"go.opencensus.io/trace"
)

type Database struct {
	db *gorm.DB
}

// Connect connects to the database
func Connect(dialect string, args ...interface{}) (*Database, error) {
	db, err := gorm.Open(dialect, args...)
	if err != nil {
		return nil, err
	}

	d := &Database{
		db: db,
	}
	err = d.autoMigrate()
	if err != nil {
		db.Close()
		return nil, err
	}

	return d, nil
}

func (d *Database) autoMigrate() error {
	for _, t := range []interface{}{&Route{}, &Community{}, &LargeCommunity{}} {
		d.db.AutoMigrate(t)
		if d.db.Error != nil {
			return d.db.Error
		}
	}

	return nil
}

// Save saves a route to the database
func (d *Database) Save(ctx context.Context, route *Route) error {
	ctx, span := trace.StartSpan(ctx, "Database.Save")
	defer span.End()

	d.db.Save(route)
	if d.db.Error != nil {
		return d.db.Error
	}

	return nil
}

// Delete removes a route from the database
func (d *Database) Delete(ctx context.Context, route *Route) error {
	ctx, span := trace.StartSpan(ctx, "Database.Delete")
	defer span.End()

	d.db.Delete(Route{}, "prefix = ? AND next_hop = ?", route.Prefix, route.NextHop)
	if d.db.Error != nil {
		return d.db.Error
	}

	return nil
}

// Routes returns all routes stored in the database
func (d *Database) Routes() ([]*Route, error) {
	var routes []*Route
	d.db.Preload("Communities").Preload("LargeCommunities").Find(&routes)
	return routes, d.db.Error
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}
