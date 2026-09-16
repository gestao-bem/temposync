package models

import "time"

type Post struct {
	ID          int64
	Slug        string
	Title       string
	Excerpt     string
	Body        string
	Category    string
	ReadMin     int
	Author      string
	PublishedAt time.Time
}
