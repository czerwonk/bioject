package database

type RouteStore interface {
	Save(route *Route) error
	Delete(route *Route) error
}
