package dto

import "time"

type CreateTagRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

type UpdateTagRequest struct {
	Name string `json:"name" binding:"required,min=1,max=50"`
}

type TagResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type TagListResponse struct {
	Tags     []TagResponse `json:"tags"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}
