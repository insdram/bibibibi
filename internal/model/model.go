package model

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"
)

// User 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:64;not null"`
	Password  string    `json:"-" gorm:"size:255;not null"`
	Email     string    `json:"email" gorm:"size:255"`
	Nickname  string    `json:"nickname" gorm:"size:64"`
	Avatar    string    `json:"avatar" gorm:"size:255"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AfterFind 查询后计算 Gravatar 头像
func (u *User) AfterFind() error {
	if u.Email != "" && u.Avatar == "" {
		u.Avatar = GetGravatarURL(u.Email)
	}
	return nil
}

// Bibi 笔记模型
type Bibi struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	CreatorID  uint      `json:"creator_id" gorm:"index;not null"`
	Content    string    `json:"content" gorm:"type:text;not null"`
	Visibility string    `json:"visibility" gorm:"size:20;default:PUBLIC"`
	Pinned     bool      `json:"pinned" gorm:"default:false"`
	LikeCount  int       `json:"like_count" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Creator    User      `json:"creator" gorm:"foreignKey:CreatorID"`
	Tags       []Tag     `json:"tags" gorm:"many2many:bibi_tags;"`
	Comments   []Comment `json:"comments" gorm:"foreignKey:BibiID"`
	Likes      []Like    `json:"likes" gorm:"foreignKey:BibiID"`
}

// Tag 标签模型
type Tag struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:64;not null"`
	CreatorID uint      `json:"creator_id" gorm:"index;not null"`
	CreatedAt time.Time `json:"created_at"`
}

// BibiTag 笔记-标签关联
type BibiTag struct {
	BibiID uint `json:"bibi_id" gorm:"primaryKey"`
	TagID  uint `json:"tag_id" gorm:"primaryKey"`
}

// Comment 评论模型
type Comment struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	BibiID    uint      `json:"bibi_id" gorm:"index;not null"`
	ParentID  uint      `json:"parent_id" gorm:"index;default:0"`
	Name      string    `json:"name" gorm:"size:64;not null"`
	Email     string    `json:"email" gorm:"size:255;not null"`
	Website   string    `json:"website" gorm:"size:255"`
	Content   string    `json:"content" gorm:"type:text;not null"`
	Avatar    string    `json:"avatar" gorm:"-"`
	CreatedAt time.Time `json:"created_at"`
}

// AfterFind 查询后计算 Gravatar 头像
func (c *Comment) AfterFind() error {
	if c.Email != "" {
		c.Avatar = GetGravatarURL(c.Email)
	}
	return nil
}

// GetGravatarURL 根据邮箱生成 Gravatar URL
func GetGravatarURL(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	hash := fmt.Sprintf("%x", md5.Sum([]byte(email)))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?s=80&d=identicon", hash)
}

// Like 点赞模型
type Like struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	BibiID    uint      `json:"bibi_id" gorm:"index;not null"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	CreatedAt time.Time `json:"created_at"`
}
