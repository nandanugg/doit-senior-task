package entity

import "time"

type URLAnalytic struct {
	ID             int64
	URLID          int64
	LongURL        string
	CreatedAt      time.Time
	ExpiresAt      time.Time
	ClickCount     int64
	LastAccessedAt *time.Time
}
