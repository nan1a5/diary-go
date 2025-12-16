package mysql

import (
	"context"
	"time"

	"diary/internal/domain"
	"diary/internal/models"
	"gorm.io/gorm"
)

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) domain.TagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, tag *domain.Tag) error {
	dbTag := &models.Tag{
		Name:      tag.Name,
		CreatedAt: time.Now(),
		IsDeleted: false,
	}

	if err := r.db.WithContext(ctx).Create(dbTag).Error; err != nil {
		return err
	}

	tag.ID = dbTag.ID
	tag.CreatedAt = dbTag.CreatedAt
	return nil
}

func (r *tagRepository) GetByID(ctx context.Context, id uint) (*domain.Tag, error) {
	var dbTag models.Tag
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&dbTag).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&dbTag), nil
}

func (r *tagRepository) GetByName(ctx context.Context, name string) (*domain.Tag, error) {
	var dbTag models.Tag
	err := r.db.WithContext(ctx).
		Where("name = ? AND is_deleted = ?", name, false).
		First(&dbTag).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&dbTag), nil
}

func (r *tagRepository) Update(ctx context.Context, tag *domain.Tag) error {
	return r.db.WithContext(ctx).
		Model(&models.Tag{}).
		Where("id = ? AND is_deleted = ?", tag.ID, false).
		Update("name", tag.Name).Error
}

func (r *tagRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Tag{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted":  true,
			"delete_time": time.Now(),
		}).Error
}

func (r *tagRepository) List(ctx context.Context, offset, limit int) ([]domain.Tag, int64, error) {
	var dbTags []models.Tag
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&models.Tag{}).
		Where("is_deleted = ?", false).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Where("is_deleted = ?", false).
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&dbTags).Error
	if err != nil {
		return nil, 0, err
	}

	tags := make([]domain.Tag, len(dbTags))
	for i, dbTag := range dbTags {
		tags[i] = *r.toDomain(&dbTag)
	}

	return tags, total, nil
}

func (r *tagRepository) GetByIDs(ctx context.Context, ids []uint) ([]domain.Tag, error) {
	var dbTags []models.Tag
	err := r.db.WithContext(ctx).
		Where("id IN ? AND is_deleted = ?", ids, false).
		Find(&dbTags).Error
	if err != nil {
		return nil, err
	}

	tags := make([]domain.Tag, len(dbTags))
	for i, dbTag := range dbTags {
		tags[i] = *r.toDomain(&dbTag)
	}
	return tags, nil
}

func (r *tagRepository) GetOrCreate(ctx context.Context, name string) (*domain.Tag, error) {
	tag, err := r.GetByName(ctx, name)
	if err == nil {
		return tag, nil
	}

	newTag := &domain.Tag{Name: name}
	if err := r.Create(ctx, newTag); err != nil {
		return nil, err
	}
	return newTag, nil
}

func (r *tagRepository) GetByDiaryID(ctx context.Context, diaryID uint) ([]domain.Tag, error) {
	var dbTags []models.Tag
	// 需要通过关联表查询，这里假设 GORM 关联已设置
	err := r.db.WithContext(ctx).
		Model(&models.Diary{ID: diaryID}).
		Association("Tags").
		Find(&dbTags)
	if err != nil {
		return nil, err
	}

	tags := make([]domain.Tag, len(dbTags))
	for i, dbTag := range dbTags {
		tags[i] = *r.toDomain(&dbTag)
	}
	return tags, nil
}

func (r *tagRepository) GetPopularTags(ctx context.Context, limit int) ([]domain.Tag, error) {
	// 这个比较复杂，通常需要聚合查询
	// SELECT t.*, COUNT(dt.diary_id) as count FROM tags t 
	// JOIN diaries_tags dt ON t.id = dt.tag_id 
	// WHERE t.is_deleted = false 
	// GROUP BY t.id 
	// ORDER BY count DESC 
	// LIMIT ?
	
	// 简化实现，或者使用原生 SQL
	var dbTags []models.Tag
	err := r.db.WithContext(ctx).
		Raw("SELECT t.* FROM tags t JOIN diaries_tags dt ON t.id = dt.tag_id WHERE t.is_deleted = ? GROUP BY t.id ORDER BY COUNT(dt.diary_id) DESC LIMIT ?", false, limit).
		Scan(&dbTags).Error
	if err != nil {
		return nil, err
	}
	
	tags := make([]domain.Tag, len(dbTags))
	for i, dbTag := range dbTags {
		tags[i] = *r.toDomain(&dbTag)
	}
	return tags, nil
}

func (r *tagRepository) toDomain(dbTag *models.Tag) *domain.Tag {
	return &domain.Tag{
		ID:         dbTag.ID,
		Name:       dbTag.Name,
		CreatedAt:  dbTag.CreatedAt,
		IsDeleted:  dbTag.IsDeleted,
		DeleteTime: dbTag.DeleteTime,
	}
}
