package service

import (
	"gorm.io/gorm"

	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
)

// TagService 标签服务
type TagService struct{}

// NewTagService 创建标签服务
func NewTagService() *TagService {
	return &TagService{}
}

// CreateTag 创建标签
func (s *TagService) CreateTag(name string, creatorID uint) (*model.Tag, error) {
	db := store.GetDB()

	tag := model.Tag{
		Name:      name,
		CreatorID: creatorID,
	}

	if err := db.Create(&tag).Error; err != nil {
		return nil, err
	}

	return &tag, nil
}

// GetTagByID 根据ID获取标签
func (s *TagService) GetTagByID(id uint) (*model.Tag, error) {
	db := store.GetDB()
	var tag model.Tag
	if err := db.First(&tag, id).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetTags 获取标签列表
func (s *TagService) GetTags(creatorID uint) ([]model.Tag, error) {
	db := store.GetDB()
	var tags []model.Tag

	query := db.Model(&model.Tag{})
	if creatorID > 0 {
		query = query.Where("creator_id = ?", creatorID)
	}

	if err := query.Order("name ASC").Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

// UpdateTag 更新标签
func (s *TagService) UpdateTag(id uint, name string) (*model.Tag, error) {
	db := store.GetDB()

	var tag model.Tag
	if err := db.First(&tag, id).Error; err != nil {
		return nil, err
	}

	tag.Name = name
	if err := db.Save(&tag).Error; err != nil {
		return nil, err
	}

	return &tag, nil
}

// DeleteTag 删除标签
func (s *TagService) DeleteTag(id uint) error {
	db := store.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		// 删除标签关联
		if err := tx.Where("tag_id = ?", id).Delete(&model.BibiTag{}).Error; err != nil {
			return err
		}

		// 删除标签
		if err := tx.Delete(&model.Tag{}, id).Error; err != nil {
			return err
		}

		return nil
	})
}
