package service

import (
	"errors"
	"fmt"

	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
	"gorm.io/gorm"
)

// CommentService 评论服务
type CommentService struct{}

// NewCommentService 创建评论服务
func NewCommentService() *CommentService {
	return &CommentService{}
}

// CreateComment 创建评论
func (s *CommentService) CreateComment(bibiID string, name, email, website, content string, parentID uint) (*model.Comment, error) {
	db := store.GetDB()

	var comment *model.Comment
	err := db.Transaction(func(tx *gorm.DB) error {
		// 检查笔记是否存在
		var bibi model.Bibi
		if err := tx.First(&bibi, "id = ?", bibiID).Error; err != nil {
			return fmt.Errorf("笔记不存在")
		}

		// 计算 Gravatar 头像
		avatar := model.GetGravatarURLWithSource(email, GetGravatarSource())

		comment = &model.Comment{
			BibiID:   bibiID,
			ParentID: parentID,
			Name:     name,
			Email:    email,
			Website:  website,
			Content:  content,
			Avatar:   avatar,
		}

		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		// 增加评论数
		if err := tx.Exec("UPDATE bibis SET comment_count = comment_count + 1 WHERE id = ?", bibiID).Error; err != nil {
			return err
		}

		return nil
	})

	return comment, err
}

// GetCommentsByBibiID 获取笔记的评论列表
func (s *CommentService) GetCommentsByBibiID(bibiID string, page, pageSize int) ([]model.Comment, int64, error) {
	db := store.GetDB()
	var comments []model.Comment
	var total int64

	query := db.Model(&model.Comment{}).Where("bibi_id = ?", bibiID)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at ASC").
		Offset(offset).Limit(pageSize).
		Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	// 手动设置 Gravatar 头像
	for i := range comments {
		comments[i].Avatar = model.GetGravatarURLWithSource(comments[i].Email, GetGravatarSource())
	}

	return comments, total, nil
}

// UpdateComment 更新评论
func (s *CommentService) UpdateComment(id uint, name, email, website, content string) (*model.Comment, error) {
	db := store.GetDB()

	var comment model.Comment
	if err := db.First(&comment, id).Error; err != nil {
		return nil, err
	}

	// 更新字段
	comment.Name = name
	comment.Email = email
	comment.Website = website
	comment.Content = content
	comment.Avatar = model.GetGravatarURLWithSource(email, GetGravatarSource())

	if err := db.Save(&comment).Error; err != nil {
		return nil, err
	}

	return &comment, nil
}

// DeleteComment 删除评论
func (s *CommentService) DeleteComment(id uint, userID uint) error {
	db := store.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		var comment model.Comment
		if err := tx.First(&comment, id).Error; err != nil {
			return err
		}

		var bibi model.Bibi
		if err := tx.First(&bibi, "id = ?", comment.BibiID).Error; err != nil {
			return err
		}

		if bibi.CreatorID != userID {
			return errors.New("无权删除此评论")
		}

		// 删除评论
		if err := tx.Delete(&model.Comment{}, id).Error; err != nil {
			return err
		}

		// 减少评论数
		return tx.Exec("UPDATE bibis SET comment_count = comment_count - 1 WHERE id = ?", comment.BibiID).Error
	})
}


