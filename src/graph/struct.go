package graph

import ()

type Sub struct {
	Name        string
	Subscribers int64
	Out         int
	In          int
}

type Link struct {
	From string
	To   string
}

var Subs map[string]*Sub
var Links []Link
var Failed []string
var Read map[string]bool

func structInit() {
	Subs = make(map[string]*Sub)
	Links = make([]Link, 0, 4)
	Failed = make([]string, 0, 4)
	Read = make(map[string]bool)
}
