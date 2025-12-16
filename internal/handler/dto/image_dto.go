package dto

import "time"

// ImageResponse 图片响应
type ImageResponse struct {
	ID        uint      `json:"id"`
	Path      string    `json:"path"`
	DiaryID   *uint     `json:"diary_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ImageListResponse struct {
	Images   []ImageResponse `json:"images"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

type AttachImageRequest struct {
	DiaryID uint `json:"diary_id" binding:"required"`
}
