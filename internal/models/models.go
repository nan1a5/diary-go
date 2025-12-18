package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Username   string    `gorm:"uniqueIndex;size:100" json:"username"`
	Password   string    `gorm:"size:255"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Diaries    []Diary   `json:"-"`
	Todos      []Todo    `json:"-"`
	IsDeleted  bool      `gorm:"default:false" json:"is_deleted"`
	DeleteTime time.Time `json:"delete_time,omitempty"`
}

type Diary struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	UserID     uint   `gorm:"index" json:"user_id"`
	Title      string `gorm:"size:255;index" json:"title"`
	Weather    string `gorm:"size:255" json:"weather"`
	Location   string `gorm:"size:255" json:"location"`
	Date       time.Time
	IsPublic   bool                   `gorm:"default:false" json:"is_public"`
	IsDeleted  bool                   `gorm:"default:false" json:"is_deleted"`
	DeleteTime time.Time              `json:"delete_time,omitempty"`
	Mood       string                 `gorm:"size:255" json:"mood"`
	Music      string                 `gorm:"type:text" json:"music"`
	IsPinned   bool                   `gorm:"default:false" json:"is_pinned"`
	ContentEnc []byte                 `gorm:"type:blob" json:"-"`
	IV         []byte                 `gorm:"type:blob" json:"-"`
	Summary    string                 `gorm:"size:512;index" json:"summary,omitempty"`     // 明文短摘用于搜索/列表（可为空）
	Properties map[string]interface{} `gorm:"serializer:json" json:"properties,omitempty"` // 扩展字段 (JSON)
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Images     []Image                `json:"images,omitempty"`
	Tags       []Tag                  `gorm:"many2many:diaries_tags" json:"tags,omitempty"`
	// Not stored:
	PlainContent string `gorm:"-" json:"content,omitempty"`      // 解密后的 Markdown 文本（仅 API 输出）
	ContentHTML  string `gorm:"-" json:"content_html,omitempty"` // 服务端渲染的 HTML（仅 API 输出，可选）
}

type Tag struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"uniqueIndex;size:100" json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	IsDeleted  bool      `gorm:"default:false" json:"is_deleted"`
	DeleteTime time.Time `json:"delete_time,omitempty"`
}

type Todo struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"index" json:"user_id"`
	Title       string     `gorm:"size:255" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	Done        bool       `json:"done"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	IsDeleted   bool       `gorm:"default:false" json:"is_deleted"`
	DeleteTime  time.Time  `json:"delete_time,omitempty"`
}

type Image struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     uint      `gorm:"index" json:"user_id"`
	DiaryID    *uint     `gorm:"index;null" json:"diary_id,omitempty"`
	Path       string    `gorm:"size:1024" json:"path"`
	CreatedAt  time.Time `json:"created_at"`
	IsDeleted  bool      `gorm:"default:false" json:"is_deleted"`
	DeleteTime time.Time `json:"delete_time,omitempty"`
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&User{}, &Diary{}, &Tag{}, &Todo{}, &Image{})
}
