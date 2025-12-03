package domain

import "time"

type Image struct {
	ID        uint
	UserID    uint
	DiaryID   *uint
	Path      string
	CreatedAt time.Time
	IsDeleted bool
	DeleteTime time.Time
}