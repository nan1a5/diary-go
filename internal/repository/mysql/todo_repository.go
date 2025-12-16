package mysql

import (
	"context"
	"time"

	"diary/internal/domain"
	"diary/internal/models"
	"gorm.io/gorm"
)

type todoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) domain.TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(ctx context.Context, todo *domain.Todo) error {
	dbTodo := &models.Todo{
		UserID:      todo.UserID,
		Title:       todo.Title,
		Description: todo.Description,
		Done:        todo.Done,
		DueDate:     todo.DueDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsDeleted:   false,
	}

	if err := r.db.WithContext(ctx).Create(dbTodo).Error; err != nil {
		return err
	}

	todo.ID = dbTodo.ID
	todo.CreatedAt = dbTodo.CreatedAt
	todo.UpdatedAt = dbTodo.UpdatedAt
	return nil
}

func (r *todoRepository) GetByID(ctx context.Context, id uint) (*domain.Todo, error) {
	var dbTodo models.Todo
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&dbTodo).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&dbTodo), nil
}

func (r *todoRepository) Update(ctx context.Context, todo *domain.Todo) error {
	updates := map[string]interface{}{
		"title":       todo.Title,
		"description": todo.Description,
		"done":        todo.Done,
		"due_date":    todo.DueDate,
		"updated_at":  time.Now(),
	}

	return r.db.WithContext(ctx).
		Model(&models.Todo{}).
		Where("id = ? AND is_deleted = ?", todo.ID, false).
		Updates(updates).Error
}

func (r *todoRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Todo{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted":  true,
			"delete_time": time.Now(),
		}).Error
}

func (r *todoRepository) ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]domain.Todo, int64, error) {
	var dbTodos []models.Todo
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&models.Todo{}).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&dbTodos).Error
	if err != nil {
		return nil, 0, err
	}

	todos := make([]domain.Todo, len(dbTodos))
	for i, dbTodo := range dbTodos {
		todos[i] = *r.toDomain(&dbTodo)
	}

	return todos, total, nil
}

func (r *todoRepository) ListByStatus(ctx context.Context, userID uint, done bool, offset, limit int) ([]domain.Todo, int64, error) {
	var dbTodos []models.Todo
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&models.Todo{}).
		Where("user_id = ? AND done = ? AND is_deleted = ?", userID, done, false).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND done = ? AND is_deleted = ?", userID, done, false).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&dbTodos).Error
	if err != nil {
		return nil, 0, err
	}

	todos := make([]domain.Todo, len(dbTodos))
	for i, dbTodo := range dbTodos {
		todos[i] = *r.toDomain(&dbTodo)
	}

	return todos, total, nil
}

func (r *todoRepository) ListByDueDate(ctx context.Context, userID uint, startDate, endDate time.Time) ([]domain.Todo, error) {
	var dbTodos []models.Todo
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ? AND due_date BETWEEN ? AND ?", userID, false, startDate, endDate).
		Order("due_date ASC").
		Find(&dbTodos).Error
	if err != nil {
		return nil, err
	}

	todos := make([]domain.Todo, len(dbTodos))
	for i, dbTodo := range dbTodos {
		todos[i] = *r.toDomain(&dbTodo)
	}
	return todos, nil
}

func (r *todoRepository) MarkAsDone(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Todo{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]interface{}{
			"done":       true,
			"updated_at": time.Now(),
		}).Error
}

func (r *todoRepository) MarkAsUndone(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Todo{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Updates(map[string]interface{}{
			"done":       false,
			"updated_at": time.Now(),
		}).Error
}

func (r *todoRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Todo{}).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *todoRepository) CountPending(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Todo{}).
		Where("user_id = ? AND done = ? AND is_deleted = ?", userID, false, false).
		Count(&count).Error
	return count, err
}

func (r *todoRepository) toDomain(dbTodo *models.Todo) *domain.Todo {
	return &domain.Todo{
		ID:          dbTodo.ID,
		UserID:      dbTodo.UserID,
		Title:       dbTodo.Title,
		Description: dbTodo.Description,
		Done:        dbTodo.Done,
		DueDate:     dbTodo.DueDate,
		CreatedAt:   dbTodo.CreatedAt,
		UpdatedAt:   dbTodo.UpdatedAt,
		IsDeleted:   dbTodo.IsDeleted,
		DeleteTime:  dbTodo.DeleteTime,
	}
}
