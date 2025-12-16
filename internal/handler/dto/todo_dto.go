package dto

import "time"

type CreateTodoRequest struct {
	Title       string     `json:"title" binding:"required,min=1,max=255"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateTodoRequest struct {
	Title       string     `json:"title" binding:"required,min=1,max=255"`
	Description string     `json:"description"`
	DueDate     *time.Time `json:"due_date"`
}

type TodoResponse struct {
	ID          uint       `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Done        bool       `json:"done"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TodoListResponse struct {
	Todos    []TodoResponse `json:"todos"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

type TodoStatsResponse struct {
	Total   int64 `json:"total"`
	Pending int64 `json:"pending"`
}
