package database

import (
	"github.com/jinzhu/gorm"
)

type Route struct {
	gorm.Model
	Prefix  string
	NextHop string
}
