package graph

import (
)

type Sub struct {
	Name string
	Subscribers int64
	Out int
	In int
}

type Link struct {
	From string
	To string
}

type aboutJSON struct {
	Kind string "kind"
	Data aboutDataJSON "data"
}

type aboutDataJSON struct {
	Name string "display_name"
	ID string "id"
	ThingID string "name"
	Image string "header_img"
	ImageSize [2]int32 "header_size"
	HeadTitle string "header_title"
	Desc string "description"
	DescHTML string "description_html"
	Title string "title"
	Url string "url"
	Created float64 "created"
	CreatedUTC float64 "created_utc"
	DescPub string "public_description"
	NumOnline int64 "accounts_active"
	NumSubs int64 "subscribers"
	Over18 bool "over18"
}

var Subs map[string]*Sub
var Links map[Link]bool
var Failed []string

func init() {
	Subs = make(map[string]*Sub)
	Links = make(map[Link]bool)
}