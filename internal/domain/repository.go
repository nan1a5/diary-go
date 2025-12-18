package domain

import (
	"context"
	"time"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// Create 创建新用户
	Create(ctx context.Context, user *User) error
	// CreateWithPassword 创建带密码的新用户
	CreateWithPassword(ctx context.Context, user *User, hashedPassword string) error
	// GetByID 根据ID获取用户
	GetByID(ctx context.Context, id uint) (*User, error)
	// GetByUsername 根据用户名获取用户
	GetByUsername(ctx context.Context, username string) (*User, error)
	// GetWithPassword 根据用户名获取用户及密码
	GetWithPassword(ctx context.Context, username string) (*User, string, error)
	// Update 更新用户信息
	Update(ctx context.Context, user *User) error
	// UpdatePassword 更新用户密码
	UpdatePassword(ctx context.Context, id uint, hashedPassword string) error
	// Delete 软删除用户
	Delete(ctx context.Context, id uint) error
	// List 获取用户列表（分页）
	List(ctx context.Context, offset, limit int) ([]User, int64, error)
}

// DiaryRepository 日记仓储接口
type DiaryRepository interface {
	// Create 创建新日记
	Create(ctx context.Context, diary *Diary) error
	// GetByID 根据ID获取日记
	GetByID(ctx context.Context, id uint) (*Diary, error)
	// Update 更新日记
	Update(ctx context.Context, diary *Diary) error
	// Delete 软删除日记
	Delete(ctx context.Context, id uint) error
	// ListByUserID 获取用户的日记列表（分页，按日期降序）
	ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]Diary, int64, error)
	// ListPublic 获取公开日记列表（分页）
	ListPublic(ctx context.Context, offset, limit int) ([]Diary, int64, error)
	// SearchByUserID 搜索用户的日记（通过标题和摘要）
	SearchByUserID(ctx context.Context, userID uint, keyword string, offset, limit int) ([]Diary, int64, error)
	// GetByDateRange 获取指定日期范围的日记
	GetByDateRange(ctx context.Context, userID uint, startDate, endDate time.Time) ([]Diary, error)
	// GetByIDs 根据ID列表批量获取日记
	GetByIDs(ctx context.Context, userID uint, ids []uint) ([]Diary, error)
	// GetByTags 根据标签获取日记
	GetByTags(ctx context.Context, userID uint, tagIDs []uint, offset, limit int) ([]Diary, int64, error)
	// AddTags 为日记添加标签
	AddTags(ctx context.Context, diaryID uint, tagIDs []uint) error
	// RemoveTags 移除日记的标签
	RemoveTags(ctx context.Context, diaryID uint, tagIDs []uint) error
	// GetWithImages 获取带图片的日记
	GetWithImages(ctx context.Context, id uint) (*Diary, error)
	// GetWithTags 获取带标签的日记
	GetWithTags(ctx context.Context, id uint) (*Diary, error)
	// GetWithAll 获取带所有关联数据的日记（图片、标签）
	GetWithAll(ctx context.Context, id uint) (*Diary, error)
	// CountByUserID 统计用户的日记总数
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	// GetMoodStats 统计心情分布
	GetMoodStats(ctx context.Context, userID uint) (map[string]int64, error)
	// GetMonthlyTrend 获取每月日记数量趋势
	GetMonthlyTrend(ctx context.Context, userID uint) ([]MonthlyTrendItem, error)
	// GetTopTags 获取最常用的标签
	GetTopTags(ctx context.Context, userID uint, limit int) ([]TopTagItem, error)
	// UpdatePinStatus 更新置顶状态
	UpdatePinStatus(ctx context.Context, id uint, isPinned bool) error
	// CountPinned 统计用户置顶日记数量
	CountPinned(ctx context.Context, userID uint) (int64, error)
}

// TodoRepository 待办事项仓储接口
type TodoRepository interface {
	// Create 创建待办事项
	Create(ctx context.Context, todo *Todo) error
	// GetByID 根据ID获取待办事项
	GetByID(ctx context.Context, id uint) (*Todo, error)
	// Update 更新待办事项
	Update(ctx context.Context, todo *Todo) error
	// Delete 软删除待办事项
	Delete(ctx context.Context, id uint) error
	// ListByUserID 获取用户的待办事项列表
	ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]Todo, int64, error)
	// ListByStatus 根据完成状态获取待办事项
	ListByStatus(ctx context.Context, userID uint, done bool, offset, limit int) ([]Todo, int64, error)
	// ListByDueDate 获取指定截止日期范围的待办事项
	ListByDueDate(ctx context.Context, userID uint, startDate, endDate time.Time) ([]Todo, error)
	// MarkAsDone 标记为已完成
	MarkAsDone(ctx context.Context, id uint) error
	// MarkAsUndone 标记为未完成
	MarkAsUndone(ctx context.Context, id uint) error
	// CountByUserID 统计用户的待办事项总数
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	// CountPending 统计用户未完成的待办事项数
	CountPending(ctx context.Context, userID uint) (int64, error)
}

// TagRepository 标签仓储接口
type TagRepository interface {
	// Create 创建标签
	Create(ctx context.Context, tag *Tag) error
	// GetByID 根据ID获取标签
	GetByID(ctx context.Context, id uint) (*Tag, error)
	// GetByName 根据名称获取标签
	GetByName(ctx context.Context, name string) (*Tag, error)
	// Update 更新标签
	Update(ctx context.Context, tag *Tag) error
	// Delete 软删除标签
	Delete(ctx context.Context, id uint) error
	// List 获取所有标签
	List(ctx context.Context, offset, limit int) ([]Tag, int64, error)
	// GetByIDs 根据ID列表批量获取标签
	GetByIDs(ctx context.Context, ids []uint) ([]Tag, error)
	// GetOrCreate 获取或创建标签（如果不存在则创建）
	GetOrCreate(ctx context.Context, name string) (*Tag, error)
	// GetByDiaryID 获取日记关联的所有标签
	GetByDiaryID(ctx context.Context, diaryID uint) ([]Tag, error)
	// GetPopularTags 获取热门标签（按使用次数排序）
	GetPopularTags(ctx context.Context, limit int) ([]Tag, error)
}

// ImageRepository 图片仓储接口
type ImageRepository interface {
	// Create 创建图片记录
	Create(ctx context.Context, image *Image) error
	// GetByID 根据ID获取图片
	GetByID(ctx context.Context, id uint) (*Image, error)
	// Update 更新图片信息
	Update(ctx context.Context, image *Image) error
	// Delete 软删除图片
	Delete(ctx context.Context, id uint) error
	// ListByUserID 获取用户的图片列表
	ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]Image, int64, error)
	// ListByDiaryID 获取日记关联的图片列表
	ListByDiaryID(ctx context.Context, diaryID uint) ([]Image, error)
	// ListUnattached 获取未关联日记的图片（用户上传但未使用的图片）
	ListUnattached(ctx context.Context, userID uint, offset, limit int) ([]Image, int64, error)
	// AttachToDiary 将图片关联到日记
	AttachToDiary(ctx context.Context, imageID, diaryID uint) error
	// DetachFromDiary 取消图片与日记的关联
	DetachFromDiary(ctx context.Context, imageID uint) error
	// CountByUserID 统计用户的图片总数
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	// DeleteByPath 根据路径删除图片
	DeleteByPath(ctx context.Context, path string) error
}

// Repository 聚合所有仓储接口
type Repository interface {
	User() UserRepository
	Diary() DiaryRepository
	Todo() TodoRepository
	Tag() TagRepository
	Image() ImageRepository
}
