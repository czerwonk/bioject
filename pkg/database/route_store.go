// SPDX-FileCopyrightText: (c) 2018 Daniel Czerwonk
//
// SPDX-License-Identifier: MIT

package database

import "context"

type RouteStore interface {
	Save(ctx context.Context, route *Route) error
	Delete(ctx context.Context, route *Route) error
}
