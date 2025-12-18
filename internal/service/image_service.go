package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"diary/config"
	"diary/internal/domain"

	"github.com/google/uuid"
)

var (
	ErrImageNotFound = errors.New("图片不存在")
	ErrImageUpload   = errors.New("图片上传失败")
)

type ImageService interface {
	Upload(ctx context.Context, userID uint, file io.Reader, filename string) (*domain.Image, error)
	GetByID(ctx context.Context, id uint) (*domain.Image, error)
	Delete(ctx context.Context, id uint) error
	ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]domain.Image, int64, error)
	ListUnattached(ctx context.Context, userID uint, page, pageSize int) ([]domain.Image, int64, error)
	AttachToDiary(ctx context.Context, imageID, diaryID uint) error
	DetachFromDiary(ctx context.Context, imageID uint) error
}

type imageService struct {
	imageRepo domain.ImageRepository
	cfg       *config.Config
}

func NewImageService(imageRepo domain.ImageRepository, cfg *config.Config) ImageService {
	// 确保上传目录存在
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		// 在这里 panic 也可以，因为服务启动时应该检查
		fmt.Printf("Error creating upload dir: %v\n", err)
	}

	return &imageService{
		imageRepo: imageRepo,
		cfg:       cfg,
	}
}

func (s *imageService) Upload(ctx context.Context, userID uint, file io.Reader, filename string) (*domain.Image, error) {
	// 生成唯一文件名
	ext := filepath.Ext(filename)
	newFilename := uuid.New().String() + ext
	// 按日期分目录，避免单目录文件过多
	dateDir := time.Now().Format("2006/01/02")
	saveDir := filepath.Join(s.cfg.UploadDir, dateDir)

	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return nil, ErrImageUpload
	}

	savePath := filepath.Join(saveDir, newFilename)
	// 相对路径用于存储和访问
	relPath := filepath.Join(dateDir, newFilename)

	// 创建文件
	dst, err := os.Create(savePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	// 写入内容
	if _, err := io.Copy(dst, file); err != nil {
		return nil, err
	}

	// 保存记录
	image := &domain.Image{
		UserID: userID,
		Path:   relPath, // 存储相对路径
	}

	if err := s.imageRepo.Create(ctx, image); err != nil {
		// 如果数据库保存失败，尝试删除文件
		os.Remove(savePath)
		return nil, err
	}

	return image, nil
}

func (s *imageService) GetByID(ctx context.Context, id uint) (*domain.Image, error) {
	image, err := s.imageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrImageNotFound
	}
	return image, nil
}

func (s *imageService) Delete(ctx context.Context, id uint) error {
	_, err := s.imageRepo.GetByID(ctx, id)
	if err != nil {
		return ErrImageNotFound
	}

	// 软删除数据库记录
	if err := s.imageRepo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

func (s *imageService) ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]domain.Image, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.imageRepo.ListByUserID(ctx, userID, offset, pageSize)
}

func (s *imageService) ListUnattached(ctx context.Context, userID uint, page, pageSize int) ([]domain.Image, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.imageRepo.ListUnattached(ctx, userID, offset, pageSize)
}

func (s *imageService) AttachToDiary(ctx context.Context, imageID, diaryID uint) error {
	return s.imageRepo.AttachToDiary(ctx, imageID, diaryID)
}

func (s *imageService) DetachFromDiary(ctx context.Context, imageID uint) error {
	return s.imageRepo.DetachFromDiary(ctx, imageID)
}
