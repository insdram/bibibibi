package service

import (
	"fmt"

	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
)

// CommentService 评论服务
type CommentService struct{}

// NewCommentService 创建评论服务
func NewCommentService() *CommentService {
	return &CommentService{}
}

// CreateComment 创建评论
func (s *CommentService) CreateComment(bibiID uint, name, email, website, content string, parentID uint) (*model.Comment, error) {
	db := store.GetDB()

	// 检查笔记是否存在
	var bibi model.Bibi
	if err := db.First(&bibi, bibiID).Error; err != nil {
		return nil, fmt.Errorf("笔记不存在")
	}

	// 计算 Gravatar 头像
	avatar := getGravatarURL(email)

	comment := model.Comment{
		BibiID:   bibiID,
		ParentID: parentID,
		Name:     name,
		Email:    email,
		Website:  website,
		Content:  content,
		Avatar:   avatar,
	}

	if err := db.Create(&comment).Error; err != nil {
		return nil, err
	}

	return &comment, nil
}

// GetCommentsByBibiID 获取笔记的评论列表
func (s *CommentService) GetCommentsByBibiID(bibiID uint, page, pageSize int) ([]model.Comment, int64, error) {
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
		comments[i].Avatar = getGravatarURL(comments[i].Email)
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
	comment.Avatar = getGravatarURL(email)

	if err := db.Save(&comment).Error; err != nil {
		return nil, err
	}

	return &comment, nil
}

// DeleteComment 删除评论
func (s *CommentService) DeleteComment(id uint) error {
	db := store.GetDB()
	return db.Delete(&model.Comment{}, id).Error
}

// getGravatarURL 根据邮箱生成 Gravatar URL
func getGravatarURL(email string) string {
	if email == "" {
		return ""
	}
	return model.GetGravatarURLWithSource(email, GetGravatarSource())
}
