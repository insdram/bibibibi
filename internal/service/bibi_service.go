package service

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
	"gorm.io/gorm"
)

// generateBibiID 生成笔记全局唯一 ID
// 算法: username + timestamp + random -> SHA1
func generateBibiID(username string) string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 16)
	for i := range randomBytes {
		randomBytes[i] = byte(timestamp >> (i * 8) & 0xFF)
	}
	input := fmt.Sprintf("%s:%d:%s", username, timestamp, string(randomBytes))
	hash := sha1.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

// regenerateCreatorAvatar 重新生成创建者的头像
func regenerateCreatorAvatar(bibi *model.Bibi) {
	if bibi.Creator.Email != "" {
		bibi.Creator.Avatar = model.GetGravatarURLWithSource(bibi.Creator.Email, GetGravatarSource())
	}
}

// regenerateCreatorAvatars 批量重新生成创建者的头像
func regenerateCreatorAvatars(bibis []model.Bibi) {
	for i := range bibis {
		regenerateCreatorAvatar(&bibis[i])
	}
}

// BibiService 笔记服务
type BibiService struct{}

// NewBibiService 创建笔记服务
func NewBibiService() *BibiService {
	return &BibiService{}
}

// CreateBibi 创建笔记
func (s *BibiService) CreateBibi(creatorID uint, content, visibility string, tagIDs []uint) (*model.Bibi, error) {
	db := store.GetDB()

	// 获取创建者用户名
	var creator model.User
	if err := db.First(&creator, creatorID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 生成全局唯一 ID
	bibiID := generateBibiID(creator.Username)

	bibi := model.Bibi{
		ID:         bibiID,
		CreatorID:  creatorID,
		Content:    content,
		Visibility: visibility,
	}

	// 开启事务
	err := db.Transaction(func(tx *gorm.DB) error {
		// 创建笔记
		if err := tx.Create(&bibi).Error; err != nil {
			return err
		}

		// 添加标签关联
		if len(tagIDs) > 0 {
			for _, tagID := range tagIDs {
				bibiTag := model.BibiTag{
					BibiID: bibi.ID,
					TagID:  tagID,
				}
				if err := tx.Create(&bibiTag).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 重新查询以加载关联数据
	if err := db.Preload("Creator").Preload("Tags").Preload("Comments").First(&bibi, "id = ?", bibi.ID).Error; err != nil {
		return nil, err
	}

	// 重新生成创建者头像
	regenerateCreatorAvatar(&bibi)

	// 为评论设置 Gravatar 头像
	for j := range bibi.Comments {
		bibi.Comments[j].Avatar = model.GetGravatarURLWithSource(bibi.Comments[j].Email, GetGravatarSource())
	}

	return &bibi, nil
}

// GetBibiByID 根据ID获取笔记
func (s *BibiService) GetBibiByID(id string) (*model.Bibi, error) {
	db := store.GetDB()
	var bibi model.Bibi
	if err := db.Preload("Creator").Preload("Tags").Preload("Comments").First(&bibi, "id = ?", id).Error; err != nil {
		return nil, err
	}
	// 重新生成创建者头像
	regenerateCreatorAvatar(&bibi)

	// 为评论设置 Gravatar 头像
	for j := range bibi.Comments {
		bibi.Comments[j].Avatar = model.GetGravatarURLWithSource(bibi.Comments[j].Email, GetGravatarSource())
	}

	return &bibi, nil
}
// GetBibis 获取笔记列表
func (s *BibiService) GetBibis(page, pageSize int, visibility string, creatorID *uint) ([]model.Bibi, int64, error) {
	db := store.GetDB()
	var bibis []model.Bibi
	var total int64

	query := db.Model(&model.Bibi{})

	// 可见性筛选
	if visibility != "" {
		query = query.Where("visibility = ?", visibility)
	} else if creatorID == nil {
		// 广场默认只显示公开笔记
		query = query.Where("visibility = 'PUBLIC'")
	}

	// 创建者筛选
	if creatorID != nil {
		query = query.Where("creator_id = ?", *creatorID)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Preload("Creator").Preload("Tags").Preload("Comments").
		Order("pinned DESC, created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&bibis).Error; err != nil {
		return nil, 0, err
	}

	// 重新生成创建者头像
	regenerateCreatorAvatars(bibis)

	// 为评论设置 Gravatar 头像
	for i := range bibis {
		for j := range bibis[i].Comments {
			bibis[i].Comments[j].Avatar = model.GetGravatarURLWithSource(bibis[i].Comments[j].Email, GetGravatarSource())
		}
	}

	return bibis, total, nil
}

// GetAllPublicBibis 获取所有公开笔记（未登录时显示）
func (s *BibiService) GetAllPublicBibis(page, pageSize int) ([]model.Bibi, int64, error) {
	db := store.GetDB()
	var bibis []model.Bibi
	var total int64

	query := db.Model(&model.Bibi{}).Where("visibility = 'PUBLIC'")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Creator").Preload("Tags").Preload("Comments").
		Order("pinned DESC, created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&bibis).Error; err != nil {
		return nil, 0, err
	}

	regenerateCreatorAvatars(bibis)
	for i := range bibis {
		for j := range bibis[i].Comments {
			bibis[i].Comments[j].Avatar = model.GetGravatarURLWithSource(bibis[i].Comments[j].Email, GetGravatarSource())
		}
	}

	return bibis, total, nil
}

// UpdateBibi 更新笔记
func (s *BibiService) UpdateBibi(id string, content, visibility string, tagIDs []uint, creatorID uint) (*model.Bibi, error) {
	db := store.GetDB()

	var bibi model.Bibi
	if err := db.First(&bibi, "id = ?", id).Error; err != nil {
		return nil, err
	}

	if bibi.CreatorID != creatorID {
		return nil, errors.New("无权操作此笔记")
	}

	// 更新字段
	bibi.Content = content
	bibi.Visibility = visibility

	// 开启事务
	err := db.Transaction(func(tx *gorm.DB) error {
		// 更新笔记
		if err := tx.Save(&bibi).Error; err != nil {
			return err
		}

		// 删除旧的标签关联
		if err := tx.Where("bibi_id = ?", id).Delete(&model.BibiTag{}).Error; err != nil {
			return err
		}

		// 添加新的标签关联
		if len(tagIDs) > 0 {
			for _, tagID := range tagIDs {
				bibiTag := model.BibiTag{
					BibiID: bibi.ID,
					TagID:  tagID,
				}
				if err := tx.Create(&bibiTag).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 重新查询以加载关联数据
	if err := db.Preload("Creator").Preload("Tags").Preload("Comments").First(&bibi, "id = ?", bibi.ID).Error; err != nil {
		return nil, err
	}

	// 重新生成创建者头像
	regenerateCreatorAvatar(&bibi)

	// 为评论设置 Gravatar 头像
	for j := range bibi.Comments {
		bibi.Comments[j].Avatar = model.GetGravatarURLWithSource(bibi.Comments[j].Email, GetGravatarSource())
	}

	return &bibi, nil
}

// DeleteBibi 删除笔记
func (s *BibiService) DeleteBibi(id string, creatorID uint) error {
	db := store.GetDB()

	var bibi model.Bibi
	if err := db.First(&bibi, "id = ?", id).Error; err != nil {
		return err
	}

	if bibi.CreatorID != creatorID {
		return errors.New("无权操作此笔记")
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// 删除标签关联
		if err := tx.Where("bibi_id = ?", id).Delete(&model.BibiTag{}).Error; err != nil {
			return err
		}

		// 删除评论
		if err := tx.Where("bibi_id = ?", id).Delete(&model.Comment{}).Error; err != nil {
			return err
		}

		// 删除点赞
		if err := tx.Where("bibi_id = ?", id).Delete(&model.Like{}).Error; err != nil {
			return err
		}

		// 删除笔记
		if err := tx.Delete(&model.Bibi{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}

// TogglePin 切换置顶状态
func (s *BibiService) TogglePin(id string, creatorID uint) (*model.Bibi, error) {
	db := store.GetDB()

	var bibi model.Bibi
	if err := db.First(&bibi, "id = ?", id).Error; err != nil {
		return nil, err
	}

	if bibi.CreatorID != creatorID {
		return nil, errors.New("无权操作此笔记")
	}

	bibi.Pinned = !bibi.Pinned
	if err := db.Save(&bibi).Error; err != nil {
		return nil, err
	}

	// 重新查询以加载关联数据
	if err := db.Preload("Creator").Preload("Tags").Preload("Comments").First(&bibi, "id = ?", bibi.ID).Error; err != nil {
		return nil, err
	}

	// 重新生成创建者头像
	regenerateCreatorAvatar(&bibi)

	// 为评论设置 Gravatar 头像
	for j := range bibi.Comments {
		bibi.Comments[j].Avatar = model.GetGravatarURLWithSource(bibi.Comments[j].Email, GetGravatarSource())
	}

	return &bibi, nil
}

// SearchBibis 搜索笔记
func (s *BibiService) SearchBibis(keyword string, page, pageSize int) ([]model.Bibi, int64, error) {
	db := store.GetDB()
	var bibis []model.Bibi
	var total int64

	query := db.Model(&model.Bibi{}).Where("content LIKE ?", "%"+keyword+"%")

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Preload("Creator").Preload("Tags").Preload("Comments").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&bibis).Error; err != nil {
		return nil, 0, err
	}

	// 重新生成创建者头像和评论头像
	regenerateCreatorAvatars(bibis)
	for i := range bibis {
		for j := range bibis[i].Comments {
			bibis[i].Comments[j].Avatar = model.GetGravatarURLWithSource(bibis[i].Comments[j].Email, GetGravatarSource())
		}
	}

	return bibis, total, nil
}
