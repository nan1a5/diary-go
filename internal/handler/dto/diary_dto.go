package dto

import "time"

type CreateDiaryRequest struct {
	Title    string    `json:"title" binding:"required,max=255"`
	Content  string    `json:"content"`
	Weather  string    `json:"weather"`
	Mood     string    `json:"mood"`
	Location string    `json:"location"`
	Date     time.Time `json:"date" binding:"required"`
	IsPublic bool      `json:"is_public"`
	Tags     []string  `json:"tags"`
	ImageIDs []uint    `json:"image_ids"`
}

type UpdateDiaryRequest struct {
	Title    string    `json:"title" binding:"required,max=255"`
	Content  string    `json:"content"`
	Weather  string    `json:"weather"`
	Mood     string    `json:"mood"`
	Location string    `json:"location"`
	Date     time.Time `json:"date" binding:"required"`
	IsPublic bool      `json:"is_public"`
	Tags     []string  `json:"tags"`
}

type DiaryResponse struct {
	ID        uint            `json:"id"`
	Title     string          `json:"title"`
	Content   string          `json:"content,omitempty"` // 仅详情返回
	Summary   string          `json:"summary"`           // 列表返回
	Weather   string          `json:"weather"`
	Mood      string          `json:"mood"`
	Location  string          `json:"location"`
	Date      time.Time       `json:"date"`
	IsPublic  bool            `json:"is_public"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Tags      []TagResponse   `json:"tags,omitempty"`
	Images    []ImageResponse `json:"images,omitempty"`
}

type DiaryListResponse struct {
	Diaries  []DiaryResponse `json:"diaries"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}
