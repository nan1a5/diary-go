package domain

import "time"

type Diary struct {
	ID          uint
	UserID      uint
	Title       string
	Weather     string
	Location    string
	Date        time.Time
	IsPublic    bool
	IsDeleted   bool
	DeleteTime 	time.Time
	Mood        string
	ContentEnc  []byte
	IV          []byte
	Summary     string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Images      []Image
	Tags        []Tag
	PlainContent string
	ContentHTML  string
}