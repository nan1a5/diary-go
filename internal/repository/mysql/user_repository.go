package mysql

import (
	"context"
	"time"

	"diary/internal/domain"
	"diary/internal/models"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// Create 创建新用户
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	dbUser := &models.User{
		Username:   user.Username,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IsDeleted:  false,
	}
	
	if err := r.db.WithContext(ctx).Create(dbUser).Error; err != nil {
		return err
	}
	
	user.ID = dbUser.ID
	user.CreatedAt = dbUser.CreatedAt
	user.UpdatedAt = dbUser.UpdatedAt
	return nil
}

// CreateWithPassword 创建带密码的新用户
func (r *userRepository) CreateWithPassword(ctx context.Context, user *domain.User, hashedPassword string) error {
	dbUser := &models.User{
		Username:   user.Username,
		Password:   hashedPassword,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IsDeleted:  false,
	}
	
	if err := r.db.WithContext(ctx).Create(dbUser).Error; err != nil {
		return err
	}
	
	user.ID = dbUser.ID
	user.CreatedAt = dbUser.CreatedAt
	user.UpdatedAt = dbUser.UpdatedAt
	return nil
}

// GetWithPassword 根据用户名获取用户及密码
func (r *userRepository) GetWithPassword(ctx context.Context, username string) (*domain.User, string, error) {
	var dbUser models.User
	err := r.db.WithContext(ctx).
		Where("username = ? AND is_deleted = ?", username, false).
		First(&dbUser).Error
	if err != nil {
		return nil, "", err
	}
	
	return r.toDomain(&dbUser), dbUser.Password, nil
}

// UpdatePassword 更新用户密码
func (r *userRepository) UpdatePassword(ctx context.Context, id uint, hashedPassword string) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND is_deleted = ?", id, false).
		Update("password", hashedPassword).Error
}

// GetByID 根据ID获取用户
func (r *userRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var dbUser models.User
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&dbUser).Error
	if err != nil {
		return nil, err
	}
	
	return r.toDomain(&dbUser), nil
}

// GetByUsername 根据用户名获取用户（包含密码字段，用于登录验证）
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var dbUser models.User
	err := r.db.WithContext(ctx).
		Where("username = ? AND is_deleted = ?", username, false).
		First(&dbUser).Error
	if err != nil {
		return nil, err
	}
	
	domainUser := r.toDomain(&dbUser)
	return domainUser, nil
}

// Update 更新用户信息
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	updates := map[string]interface{}{
		"username":   user.Username,
		"updated_at": time.Now(),
	}
	
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND is_deleted = ?", user.ID, false).
		Updates(updates).Error
}

// Delete 软删除用户
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted":  true,
			"delete_time": time.Now(),
		}).Error
}

// List 获取用户列表（分页）
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]domain.User, int64, error) {
	var dbUsers []models.User
	var total int64
	
	// 统计总数
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("is_deleted = ?", false).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 查询列表
	err := r.db.WithContext(ctx).
		Where("is_deleted = ?", false).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&dbUsers).Error
	if err != nil {
		return nil, 0, err
	}
	
	users := make([]domain.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = *r.toDomain(&dbUser)
	}
	
	return users, total, nil
}

// toDomain 将数据库模型转换为领域模型
func (r *userRepository) toDomain(dbUser *models.User) *domain.User {
	return &domain.User{
		ID:         dbUser.ID,
		Username:   dbUser.Username,
		CreatedAt:  dbUser.CreatedAt,
		UpdatedAt:  dbUser.UpdatedAt,
		IsDeleted:  dbUser.IsDeleted,
		DeleteTime: dbUser.DeleteTime,
	}
}

