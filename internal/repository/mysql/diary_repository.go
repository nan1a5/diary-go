package mysql

import (
	"context"
	"encoding/json"
	"time"

	"diary/internal/domain"
	"diary/internal/models"

	"gorm.io/gorm"
)

type diaryRepository struct {
	db *gorm.DB
}

func NewDiaryRepository(db *gorm.DB) domain.DiaryRepository {
	return &diaryRepository{db: db}
}

func (r *diaryRepository) Create(ctx context.Context, diary *domain.Diary) error {
	dbDiary := &models.Diary{
		UserID:     diary.UserID,
		Title:      diary.Title,
		Weather:    diary.Weather,
		Location:   diary.Location,
		Date:       diary.Date,
		IsPublic:   diary.IsPublic,
		Mood:       diary.Mood,
		Music:      diary.Music,
		IsPinned:   diary.IsPinned,
		ContentEnc: diary.ContentEnc,
		IV:         diary.IV,
		Summary:    diary.Summary,
		Properties: diary.Properties,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		IsDeleted:  false,
	}

	// 处理标签关联
	if len(diary.Tags) > 0 {
		var tags []models.Tag
		for _, t := range diary.Tags {
			tags = append(tags, models.Tag{ID: t.ID})
		}
		dbDiary.Tags = tags
	}

	if err := r.db.WithContext(ctx).Create(dbDiary).Error; err != nil {
		return err
	}

	diary.ID = dbDiary.ID
	diary.CreatedAt = dbDiary.CreatedAt
	diary.UpdatedAt = dbDiary.UpdatedAt
	return nil
}

func (r *diaryRepository) GetByID(ctx context.Context, id uint) (*domain.Diary, error) {
	var dbDiary models.Diary
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&dbDiary).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&dbDiary), nil
}

func (r *diaryRepository) Update(ctx context.Context, diary *domain.Diary) error {
	// Manual serialization for properties to avoid driver errors with map interface
	var propBytes []byte
	if diary.Properties != nil {
		var err error
		propBytes, err = json.Marshal(diary.Properties)
		if err != nil {
			return err
		}
	} else {
		propBytes = []byte("{}")
	}

	updates := map[string]interface{}{
		"title":       diary.Title,
		"weather":     diary.Weather,
		"location":    diary.Location,
		"date":        diary.Date,
		"is_public":   diary.IsPublic,
		"mood":        diary.Mood,
		"music":       diary.Music,
		"content_enc": diary.ContentEnc,
		"iv":          diary.IV,
		"summary":     diary.Summary,
		"properties":  propBytes,
		"updated_at":  time.Now(),
	}

	return r.db.WithContext(ctx).
		Model(&models.Diary{}).
		Where("id = ? AND is_deleted = ?", diary.ID, false).
		Updates(updates).Error
}

func (r *diaryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&models.Diary{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted":  true,
			"delete_time": time.Now(),
		}).Error
}

func (r *diaryRepository) ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]domain.Diary, int64, error) {
	var dbDiaries []models.Diary
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&models.Diary{}).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Offset(offset).
		Limit(limit).
		Order("is_pinned DESC, date DESC, created_at DESC").
		Find(&dbDiaries).Error
	if err != nil {
		return nil, 0, err
	}

	diaries := make([]domain.Diary, len(dbDiaries))
	for i, dbDiary := range dbDiaries {
		diaries[i] = *r.toDomain(&dbDiary)
	}

	return diaries, total, nil
}

func (r *diaryRepository) ListPublic(ctx context.Context, offset, limit int) ([]domain.Diary, int64, error) {
	var dbDiaries []models.Diary
	var total int64

	if err := r.db.WithContext(ctx).
		Model(&models.Diary{}).
		Where("is_public = ? AND is_deleted = ?", true, false).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Where("is_public = ? AND is_deleted = ?", true, false).
		Offset(offset).
		Limit(limit).
		Order("date DESC, created_at DESC").
		Find(&dbDiaries).Error
	if err != nil {
		return nil, 0, err
	}

	diaries := make([]domain.Diary, len(dbDiaries))
	for i, dbDiary := range dbDiaries {
		diaries[i] = *r.toDomain(&dbDiary)
	}

	return diaries, total, nil
}

func (r *diaryRepository) SearchByUserID(ctx context.Context, userID uint, keyword string, offset, limit int) ([]domain.Diary, int64, error) {
	var dbDiaries []models.Diary
	var total int64

	query := "%" + keyword + "%"

	if err := r.db.WithContext(ctx).
		Model(&models.Diary{}).
		Where("user_id = ? AND is_deleted = ? AND (title LIKE ? OR summary LIKE ?)", userID, false, query, query).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ? AND (title LIKE ? OR summary LIKE ?)", userID, false, query, query).
		Offset(offset).
		Limit(limit).
		Order("date DESC").
		Find(&dbDiaries).Error
	if err != nil {
		return nil, 0, err
	}

	diaries := make([]domain.Diary, len(dbDiaries))
	for i, dbDiary := range dbDiaries {
		diaries[i] = *r.toDomain(&dbDiary)
	}

	return diaries, total, nil
}

func (r *diaryRepository) GetByDateRange(ctx context.Context, userID uint, startDate, endDate time.Time) ([]domain.Diary, error) {
	var dbDiaries []models.Diary
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ? AND date BETWEEN ? AND ?", userID, false, startDate, endDate).
		Order("date ASC").
		Find(&dbDiaries).Error
	if err != nil {
		return nil, err
	}

	diaries := make([]domain.Diary, len(dbDiaries))
	for i, dbDiary := range dbDiaries {
		diaries[i] = *r.toDomain(&dbDiary)
	}
	return diaries, nil
}

func (r *diaryRepository) GetByIDs(ctx context.Context, userID uint, ids []uint) ([]domain.Diary, error) {
	var dbDiaries []models.Diary
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_deleted = ? AND id IN ?", userID, false, ids).
		Order("date ASC").
		Find(&dbDiaries).Error
	if err != nil {
		return nil, err
	}

	diaries := make([]domain.Diary, len(dbDiaries))
	for i, dbDiary := range dbDiaries {
		diaries[i] = *r.toDomain(&dbDiary)
	}
	return diaries, nil
}

func (r *diaryRepository) GetByTags(ctx context.Context, userID uint, tagIDs []uint, offset, limit int) ([]domain.Diary, int64, error) {
	// 这是一个复杂查询，需要关联 diaries_tags 表
	// SELECT d.* FROM diaries d
	// JOIN diaries_tags dt ON d.id = dt.diary_id
	// WHERE d.user_id = ? AND d.is_deleted = false AND dt.tag_id IN ?
	// GROUP BY d.id
	// HAVING COUNT(DISTINCT dt.tag_id) = ? (如果需要包含所有标签) 或者 >= 1 (包含任意标签)
	// 这里假设包含任意标签即可

	var dbDiaries []models.Diary
	var total int64

	// 先统计总数
	err := r.db.WithContext(ctx).
		Model(&models.Diary{}).
		Joins("JOIN diaries_tags dt ON diaries.id = dt.diary_id").
		Where("diaries.user_id = ? AND diaries.is_deleted = ? AND dt.tag_id IN ?", userID, false, tagIDs).
		Group("diaries.id").
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	err = r.db.WithContext(ctx).
		Joins("JOIN diaries_tags dt ON diaries.id = dt.diary_id").
		Where("diaries.user_id = ? AND diaries.is_deleted = ? AND dt.tag_id IN ?", userID, false, tagIDs).
		Group("diaries.id").
		Offset(offset).
		Limit(limit).
		Order("diaries.date DESC").
		Find(&dbDiaries).Error
	if err != nil {
		return nil, 0, err
	}

	diaries := make([]domain.Diary, len(dbDiaries))
	for i, dbDiary := range dbDiaries {
		diaries[i] = *r.toDomain(&dbDiary)
	}

	return diaries, total, nil
}

func (r *diaryRepository) AddTags(ctx context.Context, diaryID uint, tagIDs []uint) error {
	var tags []models.Tag
	for _, id := range tagIDs {
		tags = append(tags, models.Tag{ID: id})
	}

	return r.db.WithContext(ctx).
		Model(&models.Diary{ID: diaryID}).
		Association("Tags").
		Append(tags)
}

func (r *diaryRepository) RemoveTags(ctx context.Context, diaryID uint, tagIDs []uint) error {
	var tags []models.Tag
	for _, id := range tagIDs {
		tags = append(tags, models.Tag{ID: id})
	}

	return r.db.WithContext(ctx).
		Model(&models.Diary{ID: diaryID}).
		Association("Tags").
		Delete(tags)
}

func (r *diaryRepository) GetWithImages(ctx context.Context, id uint) (*domain.Diary, error) {
	var dbDiary models.Diary
	err := r.db.WithContext(ctx).
		Preload("Images", "is_deleted = ?", false).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&dbDiary).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&dbDiary), nil
}

func (r *diaryRepository) GetWithTags(ctx context.Context, id uint) (*domain.Diary, error) {
	var dbDiary models.Diary
	err := r.db.WithContext(ctx).
		Preload("Tags", "is_deleted = ?", false).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&dbDiary).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&dbDiary), nil
}

func (r *diaryRepository) GetWithAll(ctx context.Context, id uint) (*domain.Diary, error) {
	var dbDiary models.Diary
	err := r.db.WithContext(ctx).
		Preload("Images", "is_deleted = ?", false).
		Preload("Tags", "is_deleted = ?", false).
		Where("id = ? AND is_deleted = ?", id, false).
		First(&dbDiary).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(&dbDiary), nil
}

func (r *diaryRepository) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Diary{}).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (r *diaryRepository) GetMonthlyTrend(ctx context.Context, userID uint) ([]domain.MonthlyTrendItem, error) {
	type Result struct {
		Month string
		Count int64
	}
	var results []Result
	// Assuming MySQL
	err := r.db.WithContext(ctx).
		Model(&models.Diary{}).
		Select("DATE_FORMAT(date, '%Y-%m') as month, count(*) as count").
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Group("month").
		Order("month ASC").
		Limit(12).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	items := make([]domain.MonthlyTrendItem, len(results))
	for i, res := range results {
		items[i] = domain.MonthlyTrendItem{
			Month: res.Month,
			Count: res.Count,
		}
	}
	return items, nil
}

func (r *diaryRepository) GetTopTags(ctx context.Context, userID uint, limit int) ([]domain.TopTagItem, error) {
	type Result struct {
		TagName string
		Count   int64
	}
	var results []Result

	err := r.db.WithContext(ctx).
		Table("tags").
		Select("tags.name as tag_name, count(dt.diary_id) as count").
		Joins("JOIN diaries_tags dt ON tags.id = dt.tag_id").
		Joins("JOIN diaries d ON dt.diary_id = d.id").
		Where("d.user_id = ? AND d.is_deleted = ?", userID, false).
		Group("tags.name").
		Order("count DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	items := make([]domain.TopTagItem, len(results))
	for i, res := range results {
		items[i] = domain.TopTagItem{
			Tag:   res.TagName,
			Count: res.Count,
		}
	}
	return items, nil
}

func (r *diaryRepository) UpdatePinStatus(ctx context.Context, id uint, isPinned bool) error {
	return r.db.WithContext(ctx).
		Model(&models.Diary{}).
		Where("id = ?", id).
		Update("is_pinned", isPinned).Error
}

func (r *diaryRepository) CountPinned(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Diary{}).
		Where("user_id = ? AND is_deleted = ? AND is_pinned = ?", userID, false, true).
		Count(&count).Error
	return count, err
}

func (r *diaryRepository) GetMoodStats(ctx context.Context, userID uint) (map[string]int64, error) {
	type Result struct {
		Mood  string
		Count int64
	}
	var results []Result
	err := r.db.WithContext(ctx).
		Model(&models.Diary{}).
		Select("mood, count(*) as count").
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Group("mood").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, res := range results {
		if res.Mood == "" {
			res.Mood = "unknown"
		}
		stats[res.Mood] = res.Count
	}
	return stats, nil
}

func (r *diaryRepository) toDomain(dbDiary *models.Diary) *domain.Diary {
	diary := &domain.Diary{
		ID:         dbDiary.ID,
		UserID:     dbDiary.UserID,
		Title:      dbDiary.Title,
		Weather:    dbDiary.Weather,
		Location:   dbDiary.Location,
		Date:       dbDiary.Date,
		IsPublic:   dbDiary.IsPublic,
		Mood:       dbDiary.Mood,
		Music:      dbDiary.Music,
		IsPinned:   dbDiary.IsPinned,
		ContentEnc: dbDiary.ContentEnc,
		IV:         dbDiary.IV,
		Summary:    dbDiary.Summary,
		Properties: dbDiary.Properties,
		CreatedAt:  dbDiary.CreatedAt,
		UpdatedAt:  dbDiary.UpdatedAt,
		IsDeleted:  dbDiary.IsDeleted,
		DeleteTime: dbDiary.DeleteTime,
	}

	if len(dbDiary.Images) > 0 {
		diary.Images = make([]domain.Image, len(dbDiary.Images))
		for i, img := range dbDiary.Images {
			diary.Images[i] = domain.Image{
				ID:        img.ID,
				UserID:    img.UserID,
				DiaryID:   img.DiaryID,
				Path:      img.Path,
				CreatedAt: img.CreatedAt,
				IsDeleted: img.IsDeleted,
			}
		}
	}

	if len(dbDiary.Tags) > 0 {
		diary.Tags = make([]domain.Tag, len(dbDiary.Tags))
		for i, t := range dbDiary.Tags {
			diary.Tags[i] = domain.Tag{
				ID:        t.ID,
				Name:      t.Name,
				CreatedAt: t.CreatedAt,
				IsDeleted: t.IsDeleted,
			}
		}
	}

	return diary
}
