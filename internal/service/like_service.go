package service

import "github.com/bibibibi/bibibibi/internal/store"

type LikeService struct{}

func NewLikeService() *LikeService {
	return &LikeService{}
}

func (s *LikeService) ToggleLike(bibiID string) error {
	db := store.GetDB()
	return db.Exec("UPDATE bibis SET like_count = like_count + 1 WHERE id = ?", bibiID).Error
}
