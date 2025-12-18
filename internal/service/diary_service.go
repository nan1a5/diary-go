package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"diary/config"
	"diary/internal/domain"
	"diary/pkg/utils"
)

var (
	ErrDiaryNotFound = errors.New("日记不存在")
)

type DiaryService interface {
	Create(ctx context.Context, userID uint, title, content, weather, mood, location string, date time.Time, isPublic bool, tagNames []string, imageIDs []uint, properties map[string]interface{}, music string) (*domain.Diary, error)
	GetByID(ctx context.Context, id uint) (*domain.Diary, error)
	Update(ctx context.Context, id uint, title, content, weather, mood, location string, date time.Time, isPublic bool, tagNames []string, properties map[string]interface{}, music string) error
	Delete(ctx context.Context, id uint) error
	ListByUserID(ctx context.Context, userID uint, page, pageSize int) ([]domain.Diary, int64, error)
	ListPublic(ctx context.Context, page, pageSize int) ([]domain.Diary, int64, error)
	Search(ctx context.Context, userID uint, keyword string, page, pageSize int) ([]domain.Diary, int64, error)
	GetByDateRange(ctx context.Context, userID uint, startDate, endDate time.Time) ([]domain.Diary, error)
	GetByIDs(ctx context.Context, userID uint, ids []uint) ([]domain.Diary, error)
	TogglePin(ctx context.Context, userID, diaryID uint) (bool, error)
}

type diaryService struct {
	diaryRepo domain.DiaryRepository
	tagRepo   domain.TagRepository
	imageRepo domain.ImageRepository
	cfg       *config.Config
}

func (s *diaryService) tryEncrypt(text string) (string, error) {
	if text == "" || len(s.cfg.AESKey) != 32 {
		return text, nil
	}
	return utils.EncryptToString(s.cfg.AESKey, text)
}

func (s *diaryService) tryDecrypt(text string) string {
	if text == "" || len(s.cfg.AESKey) != 32 {
		return text
	}
	// 去除可能的空白字符
	text = strings.TrimSpace(text)
	decrypted, err := utils.DecryptFromString(s.cfg.AESKey, text)
	if err != nil {
		// 解密失败，可能是旧数据（明文），直接返回原文本
		return text
	}
	return decrypted
}

func (s *diaryService) decryptDiaries(diaries []domain.Diary) {
	for i := range diaries {
		diaries[i].Title = s.tryDecrypt(diaries[i].Title)
		diaries[i].Weather = s.tryDecrypt(diaries[i].Weather)
		diaries[i].Mood = s.tryDecrypt(diaries[i].Mood)
		diaries[i].Location = s.tryDecrypt(diaries[i].Location)
		diaries[i].Music = s.tryDecrypt(diaries[i].Music)

		// 解密内容并生成动态摘要
		if len(diaries[i].ContentEnc) > 0 && len(diaries[i].IV) > 0 && len(s.cfg.AESKey) == 32 {
			plaintext, err := utils.Decrypt(s.cfg.AESKey, diaries[i].ContentEnc, diaries[i].IV)
			if err == nil {
				content := string(plaintext)
				diaries[i].PlainContent = content
				diaries[i].Summary = makeSummary(content)
			}
		}
	}
}

func NewDiaryService(diaryRepo domain.DiaryRepository, tagRepo domain.TagRepository, imageRepo domain.ImageRepository, cfg *config.Config) DiaryService {
	return &diaryService{
		diaryRepo: diaryRepo,
		tagRepo:   tagRepo,
		imageRepo: imageRepo,
		cfg:       cfg,
	}
}

func (s *diaryService) Create(ctx context.Context, userID uint, title, content, weather, mood, location string, date time.Time, isPublic bool, tagNames []string, imageIDs []uint, properties map[string]interface{}, music string) (*domain.Diary, error) {
	// 处理标签
	var tags []domain.Tag
	for _, name := range tagNames {
		tag, err := s.tagRepo.GetOrCreate(ctx, name)
		if err != nil {
			return nil, err
		}
		tags = append(tags, *tag)
	}

	// 加密敏感字段
	encTitle, _ := s.tryEncrypt(title)
	encWeather, _ := s.tryEncrypt(weather)
	encMood, _ := s.tryEncrypt(mood)
	encLocation, _ := s.tryEncrypt(location)
	encMusic, _ := s.tryEncrypt(music)

	// 创建日记对象
	diary := &domain.Diary{
		UserID:     userID,
		Title:      encTitle,
		Weather:    encWeather,
		Mood:       encMood,
		Location:   encLocation,
		Date:       date,
		IsPublic:   isPublic,
		Tags:       tags,
		Summary:    "", // 暂时不写入摘要，保护隐私
		Properties: properties,
		Music:      encMusic,
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

	// 解密其他字段
	diary.Title = s.tryDecrypt(diary.Title)
	diary.Weather = s.tryDecrypt(diary.Weather)
	diary.Mood = s.tryDecrypt(diary.Mood)
	diary.Location = s.tryDecrypt(diary.Location)
	diary.Music = s.tryDecrypt(diary.Music)

	diary.Summary = makeSummary(diary.PlainContent)

	return diary, nil
}

func (s *diaryService) Update(ctx context.Context, id uint, title, content, weather, mood, location string, date time.Time, isPublic bool, tagNames []string, properties map[string]interface{}, music string) error {
	diary, err := s.diaryRepo.GetByID(ctx, id)
	if err != nil {
		return ErrDiaryNotFound
	}

	encTitle, _ := s.tryEncrypt(title)
	encWeather, _ := s.tryEncrypt(weather)
	encMood, _ := s.tryEncrypt(mood)
	encLocation, _ := s.tryEncrypt(location)
	encMusic, _ := s.tryEncrypt(music)

	diary.Title = encTitle
	diary.Weather = encWeather
	diary.Mood = encMood
	diary.Location = encLocation
	diary.Date = date
	diary.IsPublic = isPublic
	diary.Summary = "" // 暂时不写入摘要，保护隐私
	diary.Properties = properties
	diary.Music = encMusic

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

	var newTagIDs []uint
	for _, name := range tagNames {
		tag, err := s.tagRepo.GetOrCreate(ctx, name)
		if err == nil {
			newTagIDs = append(newTagIDs, tag.ID)
		}
	}

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

	s.decryptDiaries(diaries)

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
	diaries, total, err := s.diaryRepo.ListPublic(ctx, offset, pageSize)
	if err == nil {
		s.decryptDiaries(diaries)
	}
	return diaries, total, err
}

func (s *diaryService) Search(ctx context.Context, userID uint, keyword string, page, pageSize int) ([]domain.Diary, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	diaries, total, err := s.diaryRepo.SearchByUserID(ctx, userID, keyword, offset, pageSize)
	if err == nil {
		s.decryptDiaries(diaries)
	}
	return diaries, total, err
}

func (s *diaryService) GetByDateRange(ctx context.Context, userID uint, startDate, endDate time.Time) ([]domain.Diary, error) {
	diaries, err := s.diaryRepo.GetByDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	s.decryptDiaries(diaries)

	// Decrypt content for export
	for i := range diaries {
		if len(diaries[i].ContentEnc) > 0 && len(diaries[i].IV) > 0 && len(s.cfg.AESKey) == 32 {
			plaintext, err := utils.Decrypt(s.cfg.AESKey, diaries[i].ContentEnc, diaries[i].IV)
			if err == nil {
				diaries[i].PlainContent = string(plaintext)
			}
		}
	}
	return diaries, nil
}

func (s *diaryService) GetByIDs(ctx context.Context, userID uint, ids []uint) ([]domain.Diary, error) {
	diaries, err := s.diaryRepo.GetByIDs(ctx, userID, ids)
	if err != nil {
		return nil, err
	}

	s.decryptDiaries(diaries)

	// Decrypt content for export
	for i := range diaries {
		if len(diaries[i].ContentEnc) > 0 && len(diaries[i].IV) > 0 && len(s.cfg.AESKey) == 32 {
			plaintext, err := utils.Decrypt(s.cfg.AESKey, diaries[i].ContentEnc, diaries[i].IV)
			if err == nil {
				diaries[i].PlainContent = string(plaintext)
			}
		}
	}
	return diaries, nil
}

func (s *diaryService) TogglePin(ctx context.Context, userID, diaryID uint) (bool, error) {
	diary, err := s.diaryRepo.GetByID(ctx, diaryID)
	if err != nil {
		return false, err
	}
	if diary.UserID != userID {
		return false, errors.New("无权操作此日记")
	}

	newStatus := !diary.IsPinned
	if newStatus {
		// Checking limit
		count, err := s.diaryRepo.CountPinned(ctx, userID)
		if err != nil {
			return false, err
		}
		if count >= 3 {
			return false, errors.New("最多只能置顶3篇日记")
		}
	}

	err = s.diaryRepo.UpdatePinStatus(ctx, diaryID, newStatus)
	return newStatus, err
}

func makeSummary(content string) string {
	runes := []rune(content)
	if len(runes) > 200 {
		return string(runes[:200]) + "..."
	}
	return content
}
