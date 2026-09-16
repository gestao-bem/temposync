package models

import "time"

type Punch struct {
	ID         int64
	UserID     int64
	HappenedAt time.Time
	Kind       string
	Note       string
	Minutes    int
	CreatedAt  time.Time
}
