package mysql

import (
	"context"
	"time"

	"diary/internal/domain"
	"diary/internal/models"
	"gorm.io/gorm"
)

type imageRepository struct {
	db *gorm.DB
}

func NewImageRepository(db *gorm.DB) domain.ImageRepository {
	return &imageRepository{db: db}
}

func (r *imageRepository) Create(ctx context.Context, image *domain.Image) error {
	dbImage := &models.Image{
		UserID:    image.UserID,
		DiaryID:   image.DiaryID,
		Path:      image.Path,
		CreatedAt: time.Now(),
		IsDeleted: false,
	}

	if err := r.db.WithContext(ctx).Create(dbImage).Error; err != nil {
		return err
	}

	image.ID = dbImage.ID
	image.CreatedAt = dbImage.CreatedAt
	return nil
}

func (r *imageRepository) GetByID(ctx context.Context, id uint) (*domain.Image, error) {
	var dbImage models.Image
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&dbImage).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&dbImage), nil
}

func (r *imageRepository) Update(ctx context.Context, image *domain.Image) error {
	return r.db.WithContext(ctx).
		Model(&models.Image{}).
		Where("id = ? AND is_deleted = ?", image.ID, false).
		Updates(map[string]interface{}{
			"diary_id": image.DiaryID,
			"path":     image.Path,
		}).Error
}

func (r *imageRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Image{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted":  true,
			"delete_time": time.Now(),
		}).Error
}

func (r *imageRepository) ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]domain.Image, int64, error) {
	var dbImages []models.Image
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&models.Image{}).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&dbImages).Error
	if err != nil {
		return nil, 0, err
	}

	images := make([]domain.Image, len(dbImages))
	for i, dbImage := range dbImages {
		images[i] = *r.toDomain(&dbImage)
	}

	return images, total, nil
}

func (r *imageRepository) ListByDiaryID(ctx context.Context, diaryID uint) ([]domain.Image, error) {
	var dbImages []models.Image
	err := r.db.WithContext(ctx).
		Where("diary_id = ? AND is_deleted = ?", diaryID, false).
		Find(&dbImages).Error
	if err != nil {
		return nil, err
	}

	images := make([]domain.Image, len(dbImages))
	for i, dbImage := range dbImages {
		images[i] = *r.toDomain(&dbImage)
	}
	return images, nil
}

func (r *imageRepository) ListUnattached(ctx context.Context, userID uint, offset, limit int) ([]domain.Image, int64, error) {
	var dbImages []models.Image
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&models.Image{}).
		Where("user_id = ? AND diary_id IS NULL AND is_deleted = ?", userID, false).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND diary_id IS NULL AND is_deleted = ?", userID, false).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&dbImages).Error
	if err != nil {
		return nil, 0, err
	}

	images := make([]domain.Image, len(dbImages))
	for i, dbImage := range dbImages {
		images[i] = *r.toDomain(&dbImage)
	}

	return images, total, nil
}

func (r *imageRepository) AttachToDiary(ctx context.Context, imageID, diaryID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Image{}).
		Where("id = ? AND is_deleted = ?", imageID, false).
		Update("diary_id", diaryID).Error
}

func (r *imageRepository) DetachFromDiary(ctx context.Context, imageID uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Image{}).
		Where("id = ? AND is_deleted = ?", imageID, false).
		Update("diary_id", nil).Error
}

func (r *imageRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Image{}).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *imageRepository) DeleteByPath(ctx context.Context, path string) error {
	return r.db.WithContext(ctx).
		Model(&models.Image{}).
		Where("path = ?", path).
		Updates(map[string]interface{}{
			"is_deleted":  true,
			"delete_time": time.Now(),
		}).Error
}

func (r *imageRepository) toDomain(dbImage *models.Image) *domain.Image {
	return &domain.Image{
		ID:         dbImage.ID,
		UserID:     dbImage.UserID,
		DiaryID:    dbImage.DiaryID,
		Path:       dbImage.Path,
		CreatedAt:  dbImage.CreatedAt,
		IsDeleted:  dbImage.IsDeleted,
		DeleteTime: dbImage.DeleteTime,
	}
}
