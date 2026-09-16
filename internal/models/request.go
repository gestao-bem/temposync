package models

import (
	"database/sql"
	"time"
)

type Request struct {
	ID        int64
	UserID    int64
	Kind      string // folga | ajuste
	Day       time.Time
	Minutes   int
	Reason    string
	Status    string // pendente | aprovada | cancelada
	CreatedAt time.Time
	DecidedAt sql.NullTime
}
