package service

import (
	"errors"

	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
	"gorm.io/gorm"
)

// LikeService 点赞服务
type LikeService struct{}

// NewLikeService 创建点赞服务
func NewLikeService() *LikeService {
	return &LikeService{}
}

// ToggleLike 点赞（不可取消）
func (s *LikeService) ToggleLike(bibiID, userID uint) (bool, error) {
	db := store.GetDB()

	// 检查笔记是否存在
	var bibi model.Bibi
	if err := db.First(&bibi, bibiID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("笔记不存在")
		}
		return false, err
	}

	// 添加点赞
	like := model.Like{
		BibiID: bibiID,
		UserID:  userID,
	}
	if err := db.Create(&like).Error; err != nil {
		return false, err
	}
	// 增加点赞数
	db.Exec("UPDATE bibis SET like_count = like_count + 1 WHERE id = ?", bibiID)
	return true, nil
}

// GetLikeStatus 获取用户对笔记的点赞状态
func (s *LikeService) GetLikeStatus(bibiID, userID uint) (bool, error) {
	db := store.GetDB()
	var like model.Like
	err := db.Where("bibi_id = ? AND user_id = ?", bibiID, userID).First(&like).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetLikesByBibiID 获取笔记的所有点赞
func (s *LikeService) GetLikesByBibiID(bibiID uint) ([]model.Like, error) {
	db := store.GetDB()
	var likes []model.Like
	if err := db.Where("bibi_id = ?", bibiID).Find(&likes).Error; err != nil {
		return nil, err
	}
	return likes, nil
}
