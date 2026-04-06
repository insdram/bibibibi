package service

import (
	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
	"gorm.io/gorm"
)

// BibiService 笔记服务
type BibiService struct{}

// NewBibiService 创建笔记服务
func NewBibiService() *BibiService {
	return &BibiService{}
}

// CreateBibi 创建笔记
func (s *BibiService) CreateBibi(creatorID uint, content, visibility string, tagIDs []uint) (*model.Bibi, error) {
	db := store.GetDB()

	bibi := model.Bibi{
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
	if err := db.Preload("Creator").Preload("Tags").Preload("Comments").First(&bibi, bibi.ID).Error; err != nil {
		return nil, err
	}

	// 为评论设置 Gravatar 头像
	for j := range bibi.Comments {
		bibi.Comments[j].Avatar = model.GetGravatarURL(bibi.Comments[j].Email)
	}

	return &bibi, nil
}

// GetBibiByID 根据ID获取笔记
func (s *BibiService) GetBibiByID(id uint) (*model.Bibi, error) {
	db := store.GetDB()
	var bibi model.Bibi
	if err := db.Preload("Creator").Preload("Tags").Preload("Comments").First(&bibi, id).Error; err != nil {
		return nil, err
	}
	// 为评论设置 Gravatar 头像
	for j := range bibi.Comments {
		bibi.Comments[j].Avatar = model.GetGravatarURL(bibi.Comments[j].Email)
	}
	return &bibi, nil
}

// GetBibis 获取笔记列表
func (s *BibiService) GetBibis(page, pageSize int, visibility string) ([]model.Bibi, int64, error) {
	db := store.GetDB()
	var bibis []model.Bibi
	var total int64

	query := db.Model(&model.Bibi{})

	// 可见性筛选
	if visibility != "" {
		query = query.Where("visibility = ?", visibility)
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

	// 为评论设置 Gravatar 头像
	for i := range bibis {
		for j := range bibis[i].Comments {
			bibis[i].Comments[j].Avatar = model.GetGravatarURL(bibis[i].Comments[j].Email)
		}
	}

	return bibis, total, nil
}

// UpdateBibi 更新笔记
func (s *BibiService) UpdateBibi(id uint, content, visibility string, tagIDs []uint) (*model.Bibi, error) {
	db := store.GetDB()

	var bibi model.Bibi
	if err := db.First(&bibi, id).Error; err != nil {
		return nil, err
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
	if err := db.Preload("Creator").Preload("Tags").Preload("Comments").First(&bibi, bibi.ID).Error; err != nil {
		return nil, err
	}

	// 为评论设置 Gravatar 头像
	for j := range bibi.Comments {
		bibi.Comments[j].Avatar = model.GetGravatarURL(bibi.Comments[j].Email)
	}

	return &bibi, nil
}

// DeleteBibi 删除笔记
func (s *BibiService) DeleteBibi(id uint) error {
	db := store.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		// 删除标签关联
		if err := tx.Where("bibi_id = ?", id).Delete(&model.BibiTag{}).Error; err != nil {
			return err
		}

		// 删除评论
		if err := tx.Where("bibi_id = ?", id).Delete(&model.Comment{}).Error; err != nil {
			return err
		}

		// 删除笔记
		if err := tx.Delete(&model.Bibi{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}

// TogglePin 切换置顶状态
func (s *BibiService) TogglePin(id uint) (*model.Bibi, error) {
	db := store.GetDB()

	var bibi model.Bibi
	if err := db.First(&bibi, id).Error; err != nil {
		return nil, err
	}

	bibi.Pinned = !bibi.Pinned
	if err := db.Save(&bibi).Error; err != nil {
		return nil, err
	}

	// 重新查询以加载关联数据
	if err := db.Preload("Creator").Preload("Tags").Preload("Comments").First(&bibi, bibi.ID).Error; err != nil {
		return nil, err
	}

	// 为评论设置 Gravatar 头像
	for j := range bibi.Comments {
		bibi.Comments[j].Avatar = model.GetGravatarURL(bibi.Comments[j].Email)
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

	// 为评论设置 Gravatar 头像
	for i := range bibis {
		for j := range bibis[i].Comments {
			bibis[i].Comments[j].Avatar = model.GetGravatarURL(bibis[i].Comments[j].Email)
		}
	}

	return bibis, total, nil
}
