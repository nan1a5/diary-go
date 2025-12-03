package domain

import "time"	

type Todo struct {
	ID          uint
	UserID      uint
	Title       string
	Description string
	Done        bool
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	IsDeleted   bool
	DeleteTime time.Time
}