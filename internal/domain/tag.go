package domain

import "time"

type Tag struct {
	ID        uint
	Name      string
	CreatedAt time.Time
	IsDeleted bool
	DeleteTime time.Time
}