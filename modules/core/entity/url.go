package entity

import "time"

type URL struct {
	ID        int64
	LongURL   string
	ExpiresAt time.Time
}
