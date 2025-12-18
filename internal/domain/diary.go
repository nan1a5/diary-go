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
	Music       string
	IsPinned    bool
	ContentEnc  []byte
	IV          []byte
	Summary     string
	Properties  map[string]interface{}
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Images      []Image
	Tags        []Tag
	PlainContent string
	ContentHTML  string
}

type MonthlyTrendItem struct {
	Month string
	Count int64
}

type TopTagItem struct {
	Tag   string
	Count int64
}