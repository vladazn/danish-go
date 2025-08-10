package server

import (
	"github.com/vladazn/danish/app/classroom"
)

type handler struct {
	dict *classroom.Dictionary
	pool *classroom.WordPool
	set  *classroom.SetService
}
