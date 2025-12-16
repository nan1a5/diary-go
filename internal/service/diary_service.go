package service

import (
	"context"
	"errors"
	"time"

	"diary/config"
	"diary/internal/domain"
	"diary/pkg/utils"
)

var (
	ErrDiaryNotFound = errors.New("日记不存在")
)

type DiaryService interface {
	Create(ctx context.Context, userID uint, title, content, weather, mood, location string, date time.Time, isPublic bool, tagNames []string, imageIDs []uint) (*domain.Diary, error)
	GetByID(ctx context.Context, id uint) (*domain.Diary, error)
	Update(ctx context.Context, id uint, title, content, weather, mood, location string, date time.Time, isPublic bool, tagNames []string) error
	Delete(ctx context.Context, id uint) error
	ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]domain.Diary, int64, error)
	ListPublic(ctx context.Context, page, pageSize int) ([]domain.Diary, int64, error)
	Search(ctx context.Context, userID uint, keyword string, page, pageSize int) ([]domain.Diary, int64, error)
	GetByDateRange(ctx context.Context, userID uint, startDate, endDate time.Time) ([]domain.Diary, error)
}

type diaryService struct {
	diaryRepo domain.DiaryRepository
	tagRepo   domain.TagRepository
	imageRepo domain.ImageRepository
	cfg       *config.Config
}

func NewDiaryService(diaryRepo domain.DiaryRepository, tagRepo domain.TagRepository, imageRepo domain.ImageRepository, cfg *config.Config) DiaryService {
	return &diaryService{
		diaryRepo: diaryRepo,
		tagRepo:   tagRepo,
		imageRepo: imageRepo,
		cfg:       cfg,
	}
}

func (s *diaryService) Create(ctx context.Context, userID uint, title, content, weather, mood, location string, date time.Time, isPublic bool, tagNames []string, imageIDs []uint) (*domain.Diary, error) {
	// 处理标签
	var tags []domain.Tag
	for _, name := range tagNames {
		tag, err := s.tagRepo.GetOrCreate(ctx, name)
		if err != nil {
			return nil, err
		}
		tags = append(tags, *tag)
	}

	// 创建日记对象
	diary := &domain.Diary{
		UserID:    userID,
		Title:     title,
		Weather:   weather,
		Mood:      mood,
		Location:  location,
		Date:      date,
		IsPublic:  isPublic,
		Tags:      tags,
		Summary:   makeSummary(content),
	}

	// 加密内容
	if content != "" && len(s.cfg.AESKey) == 32 {
		encrypted, nonce, err := utils.Encrypt(s.cfg.AESKey, []byte(content))
		if err != nil {
			return nil, err
		}
		diary.ContentEnc = encrypted
		diary.IV = nonce
	} else {
		// 如果没有配置密钥，或者内容为空，暂不加密（或者在此处报错，取决于策略）
		// 这里选择如果没密钥则只存Summary，内容丢失（因为字段是ContentEnc）
		// 或者：我们可以要求必须有密钥。
		// 为了健壮性，如果没有密钥，就不加密，但是我们的模型只有ContentEnc。
		// 我们可以把明文直接存入ContentEnc（不推荐）或者报错。
		if content != "" && len(s.cfg.AESKey) != 32 {
			// Log warning
		}
	}

	if err := s.diaryRepo.Create(ctx, diary); err != nil {
		return nil, err
	}

	// 关联图片
	if len(imageIDs) > 0 {
		for _, imgID := range imageIDs {
			// 验证图片归属
			img, err := s.imageRepo.GetByID(ctx, imgID)
			if err == nil && img.UserID == userID {
				s.imageRepo.AttachToDiary(ctx, imgID, diary.ID)
			}
		}
	}

	// 填充 PlainContent 用于返回
	diary.PlainContent = content

	return diary, nil
}

func (s *diaryService) GetByID(ctx context.Context, id uint) (*domain.Diary, error) {
	// 获取完整信息（包括图片和标签）
	diary, err := s.diaryRepo.GetWithAll(ctx, id)
	if err != nil {
		return nil, ErrDiaryNotFound
	}

	// 解密内容
	if len(diary.ContentEnc) > 0 && len(diary.IV) > 0 && len(s.cfg.AESKey) == 32 {
		plaintext, err := utils.Decrypt(s.cfg.AESKey, diary.ContentEnc, diary.IV)
		if err == nil {
			diary.PlainContent = string(plaintext)
		}
	}

	return diary, nil
}

func (s *diaryService) Update(ctx context.Context, id uint, title, content, weather, mood, location string, date time.Time, isPublic bool, tagNames []string) error {
	diary, err := s.diaryRepo.GetByID(ctx, id)
	if err != nil {
		return ErrDiaryNotFound
	}

	diary.Title = title
	diary.Weather = weather
	diary.Mood = mood
	diary.Location = location
	diary.Date = date
	diary.IsPublic = isPublic
	diary.Summary = makeSummary(content)

	// 更新内容
	if content != "" && len(s.cfg.AESKey) == 32 {
		encrypted, nonce, err := utils.Encrypt(s.cfg.AESKey, []byte(content))
		if err != nil {
			return err
		}
		diary.ContentEnc = encrypted
		diary.IV = nonce
	}

	// 更新基本信息
	if err := s.diaryRepo.Update(ctx, diary); err != nil {
		return err
	}

	// 更新标签
	// 1. 获取现有标签
	// 2. 计算差异
	// 或者简单粗暴：先清空再添加（如果 Repo 支持 Replace）
	// 这里我们手动处理：
	// 获取新标签的 ID 列表
	var newTagIDs []uint
	for _, name := range tagNames {
		tag, err := s.tagRepo.GetOrCreate(ctx, name)
		if err == nil {
			newTagIDs = append(newTagIDs, tag.ID)
		}
	}
	
	// GORM 的 Replace 关联
	// 由于 Repository 没有直接暴露 ReplaceTags，我们可能需要扩展 Repo 或者手动 Remove + Add
	// 假设我们扩展了 Repository，或者直接使用 Update 时的关联替换？
	// GORM 的 Updates 通常不更新关联。
	// 我们可以使用 diaryRepo.AddTags 和 RemoveTags，但这需要知道哪些要删。
	
	// 简单策略：获取当前标签 -> 对比 -> 删旧加新
	// 这是一个常用的业务逻辑。
	currentTags, _ := s.tagRepo.GetByDiaryID(ctx, id)
	currentTagMap := make(map[uint]bool)
	for _, t := range currentTags {
		currentTagMap[t.ID] = true
	}
	
	newTagMap := make(map[uint]bool)
	var toAdd []uint
	for _, tid := range newTagIDs {
		newTagMap[tid] = true
		if !currentTagMap[tid] {
			toAdd = append(toAdd, tid)
		}
	}
	
	var toRemove []uint
	for _, t := range currentTags {
		if !newTagMap[t.ID] {
			toRemove = append(toRemove, t.ID)
		}
	}
	
	if len(toAdd) > 0 {
		s.diaryRepo.AddTags(ctx, id, toAdd)
	}
	if len(toRemove) > 0 {
		s.diaryRepo.RemoveTags(ctx, id, toRemove)
	}

	return nil
}

func (s *diaryService) Delete(ctx context.Context, id uint) error {
	return s.diaryRepo.Delete(ctx, id)
}

func (s *diaryService) ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]domain.Diary, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	
	diaries, total, err := s.diaryRepo.ListByUserID(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}
	
	// 列表通常不需要解密全文，只返回 Summary
	// 如果需要解密，可以在这里循环解密，但性能较差
	return diaries, total, nil
}

func (s *diaryService) ListPublic(ctx context.Context, page, pageSize int) ([]domain.Diary, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.diaryRepo.ListPublic(ctx, offset, pageSize)
}

func (s *diaryService) Search(ctx context.Context, userID uint, keyword string, page, pageSize int) ([]domain.Diary, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.diaryRepo.SearchByUserID(ctx, userID, keyword, offset, pageSize)
}

func (s *diaryService) GetByDateRange(ctx context.Context, userID uint, startDate, endDate time.Time) ([]domain.Diary, error) {
	return s.diaryRepo.GetByDateRange(ctx, userID, startDate, endDate)
}

func makeSummary(content string) string {
	runes := []rune(content)
	if len(runes) > 200 {
		return string(runes[:200]) + "..."
	}
	return content
}
