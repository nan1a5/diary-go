package service

import (
	"context"
	"errors"
	"time"

	"diary/config"
	"diary/internal/domain"
	"diary/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户已存在")
	ErrInvalidPassword   = errors.New("密码错误")
	ErrInvalidUsername   = errors.New("用户名格式不正确")
	ErrUnableResgister   = errors.New("不允许注册")
)

type UserService interface {
	// Register 用户注册
	Register(ctx context.Context, username, password string) (*domain.User, error)
	// Login 用户登录
	Login(ctx context.Context, username, password string) (*domain.User, string, error)
	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id uint) (*domain.User, error)
	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	// UpdateUsername 更新用户名
	UpdateUsername(ctx context.Context, id uint, newUsername string) error
	// UpdatePassword 更新密码
	UpdatePassword(ctx context.Context, id uint, oldPassword, newPassword string) error
	// Delete 删除用户
	Delete(ctx context.Context, id uint) error
	// List 获取用户列表
	List(ctx context.Context, page, pageSize int) ([]domain.User, int64, error)
}

type userService struct {
	userRepo domain.UserRepository
	cfg      *config.Config
}

func NewUserService(userRepo domain.UserRepository, cfg *config.Config) UserService {
	return &userService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Register 用户注册
func (s *userService) Register(ctx context.Context, username, password string) (*domain.User, error) {
	if !s.cfg.EnableRegistration {
		return nil, ErrUnableResgister
	}
	// 验证用户名格式
	if len(username) < 3 || len(username) > 50 {
		return nil, ErrInvalidUsername
	}

	// 验证密码长度
	if len(password) < 6 {
		return nil, errors.New("密码至少需要6位")
	}

	// 检查用户名是否已存在
	existingUser, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户（注意：这里需要在repository中支持保存密码）
	user := &domain.User{
		Username:  username,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 这里需要通过models创建，因为domain.User不包含Password
	// 我们需要在repository层处理这个
	if err := s.createUserWithPassword(ctx, user, string(hashedPassword)); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *userService) Login(ctx context.Context, username, password string) (*domain.User, string, error) {
	// 获取用户（需要包含密码字段）
	user, hashedPassword, err := s.getUserWithPassword(ctx, username)
	if err != nil {
		return nil, "", ErrUserNotFound
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, "", ErrInvalidPassword
	}

	// 生成JWT Token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// GetByID 根据ID获取用户
func (s *userService) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetByUsername 根据用户名获取用户
func (s *userService) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateUsername 更新用户名
func (s *userService) UpdateUsername(ctx context.Context, id uint, newUsername string) error {
	// 验证新用户名
	if len(newUsername) < 3 || len(newUsername) > 50 {
		return ErrInvalidUsername
	}

	// 检查新用户名是否已被占用
	existingUser, err := s.userRepo.GetByUsername(ctx, newUsername)
	if err == nil && existingUser != nil && existingUser.ID != id {
		return ErrUserAlreadyExists
	}

	// 更新用户名
	user := &domain.User{
		ID:       id,
		Username: newUsername,
	}

	return s.userRepo.Update(ctx, user)
}

// UpdatePassword 更新密码
func (s *userService) UpdatePassword(ctx context.Context, id uint, oldPassword, newPassword string) error {
	// 验证新密码
	if len(newPassword) < 6 {
		return errors.New("密码至少需要6位")
	}

	// 获取用户并验证旧密码
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	// 验证旧密码（需要获取加密后的密码）
	_, hashedPassword, err := s.getUserWithPassword(ctx, user.Username)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(oldPassword)); err != nil {
		return ErrInvalidPassword
	}

	// 加密新密码并更新
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.updateUserPassword(ctx, id, string(newHashedPassword))
}

// Delete 删除用户
func (s *userService) Delete(ctx context.Context, id uint) error {
	return s.userRepo.Delete(ctx, id)
}

// List 获取用户列表
func (s *userService) List(ctx context.Context, page, pageSize int) ([]domain.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.userRepo.List(ctx, offset, pageSize)
}

// 辅助方法：创建带密码的用户
func (s *userService) createUserWithPassword(ctx context.Context, user *domain.User, hashedPassword string) error {
	return s.userRepo.CreateWithPassword(ctx, user, hashedPassword)
}

// 辅助方法：获取用户及其密码
func (s *userService) getUserWithPassword(ctx context.Context, username string) (*domain.User, string, error) {
	return s.userRepo.GetWithPassword(ctx, username)
}

// 辅助方法：更新用户密码
func (s *userService) updateUserPassword(ctx context.Context, id uint, hashedPassword string) error {
	return s.userRepo.UpdatePassword(ctx, id, hashedPassword)
}

// 辅助方法：生成JWT Token
func (s *userService) generateToken(userID uint) (string, error) {
	return utils.CreateJWTToken(userID, s.cfg)
}
