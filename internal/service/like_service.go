package service

import (
	"github.com/bibibibi/bibibibi/internal/store"
)

// LikeService 点赞服务
type LikeService struct{}

// NewLikeService 创建点赞服务
func NewLikeService() *LikeService {
	return &LikeService{}
}

// ToggleLike 点赞（前端已做10秒冷却，后端只管+1）
func (s *LikeService) ToggleLike(bibiID string) error {
	db := store.GetDB()
	return db.Exec("UPDATE bibis SET like_count = like_count + 1 WHERE id = ?", bibiID).Error
}
