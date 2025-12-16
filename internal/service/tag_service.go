package service

import (
	"context"
	"errors"

	"diary/internal/domain"
)

var (
	ErrTagNotFound      = errors.New("标签不存在")
	ErrTagAlreadyExists = errors.New("标签已存在")
)

type TagService interface {
	Create(ctx context.Context, name string) (*domain.Tag, error)
	GetByID(ctx context.Context, id uint) (*domain.Tag, error)
	GetByName(ctx context.Context, name string) (*domain.Tag, error)
	Update(ctx context.Context, id uint, name string) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, page, pageSize int) ([]domain.Tag, int64, error)
	GetPopularTags(ctx context.Context, limit int) ([]domain.Tag, error)
}

type tagService struct {
	tagRepo domain.TagRepository
}

func NewTagService(tagRepo domain.TagRepository) TagService {
	return &tagService{
		tagRepo: tagRepo,
	}
}

func (s *tagService) Create(ctx context.Context, name string) (*domain.Tag, error) {
	existingTag, err := s.tagRepo.GetByName(ctx, name)
	if err == nil && existingTag != nil {
		return nil, ErrTagAlreadyExists
	}

	tag := &domain.Tag{
		Name: name,
	}

	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *tagService) GetByID(ctx context.Context, id uint) (*domain.Tag, error) {
	tag, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTagNotFound
	}
	return tag, nil
}

func (s *tagService) GetByName(ctx context.Context, name string) (*domain.Tag, error) {
	tag, err := s.tagRepo.GetByName(ctx, name)
	if err != nil {
		return nil, ErrTagNotFound
	}
	return tag, nil
}

func (s *tagService) Update(ctx context.Context, id uint, name string) error {
	tag, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		return ErrTagNotFound
	}

	// 检查名称是否重复
	if tag.Name != name {
		existingTag, err := s.tagRepo.GetByName(ctx, name)
		if err == nil && existingTag != nil {
			return ErrTagAlreadyExists
		}
	}

	tag.Name = name
	return s.tagRepo.Update(ctx, tag)
}

func (s *tagService) Delete(ctx context.Context, id uint) error {
	return s.tagRepo.Delete(ctx, id)
}

func (s *tagService) List(ctx context.Context, page, pageSize int) ([]domain.Tag, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	return s.tagRepo.List(ctx, offset, pageSize)
}

func (s *tagService) GetPopularTags(ctx context.Context, limit int) ([]domain.Tag, error) {
	if limit < 1 {
		limit = 10
	}
	return s.tagRepo.GetPopularTags(ctx, limit)
}
