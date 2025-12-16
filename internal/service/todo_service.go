package service

import (
	"context"
	"errors"
	"time"

	"diary/internal/domain"
)

var (
	ErrTodoNotFound = errors.New("待办事项不存在")
)

type TodoService interface {
	Create(ctx context.Context, userID uint, title, description string, dueDate *time.Time) (*domain.Todo, error)
	GetByID(ctx context.Context, id uint) (*domain.Todo, error)
	Update(ctx context.Context, id uint, title, description string, dueDate *time.Time) error
	Delete(ctx context.Context, id uint) error
	MarkAsDone(ctx context.Context, id uint) error
	MarkAsUndone(ctx context.Context, id uint) error
	ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]domain.Todo, int64, error)
	ListByStatus(ctx context.Context, userID uint, done bool, page, pageSize int) ([]domain.Todo, int64, error)
	ListByDueDate(ctx context.Context, userID uint, startDate, endDate time.Time) ([]domain.Todo, error)
	GetStats(ctx context.Context, userID uint) (total, pending int64, err error)
}

type todoService struct {
	todoRepo domain.TodoRepository
}

func NewTodoService(todoRepo domain.TodoRepository) TodoService {
	return &todoService{
		todoRepo: todoRepo,
	}
}

func (s *todoService) Create(ctx context.Context, userID uint, title, description string, dueDate *time.Time) (*domain.Todo, error) {
	todo := &domain.Todo{
		UserID:      userID,
		Title:       title,
		Description: description,
		Done:        false,
		DueDate:     dueDate,
	}

	if err := s.todoRepo.Create(ctx, todo); err != nil {
		return nil, err
	}

	return todo, nil
}

func (s *todoService) GetByID(ctx context.Context, id uint) (*domain.Todo, error) {
	todo, err := s.todoRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTodoNotFound
	}
	return todo, nil
}

func (s *todoService) Update(ctx context.Context, id uint, title, description string, dueDate *time.Time) error {
	todo, err := s.todoRepo.GetByID(ctx, id)
	if err != nil {
		return ErrTodoNotFound
	}

	todo.Title = title
	todo.Description = description
	todo.DueDate = dueDate

	return s.todoRepo.Update(ctx, todo)
}

func (s *todoService) Delete(ctx context.Context, id uint) error {
	return s.todoRepo.Delete(ctx, id)
}

func (s *todoService) MarkAsDone(ctx context.Context, id uint) error {
	return s.todoRepo.MarkAsDone(ctx, id)
}

func (s *todoService) MarkAsUndone(ctx context.Context, id uint) error {
	return s.todoRepo.MarkAsUndone(ctx, id)
}

func (s *todoService) ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]domain.Todo, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.todoRepo.ListByUserID(ctx, userID, offset, pageSize)
}

func (s *todoService) ListByStatus(ctx context.Context, userID uint, done bool, page, pageSize int) ([]domain.Todo, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.todoRepo.ListByStatus(ctx, userID, done, offset, pageSize)
}

func (s *todoService) ListByDueDate(ctx context.Context, userID uint, startDate, endDate time.Time) ([]domain.Todo, error) {
	return s.todoRepo.ListByDueDate(ctx, userID, startDate, endDate)
}

func (s *todoService) GetStats(ctx context.Context, userID uint) (int64, int64, error) {
	total, err := s.todoRepo.CountByUserID(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	pending, err := s.todoRepo.CountPending(ctx, userID)
	if err != nil {
		return 0, 0, err
	}
	return total, pending, nil
}
