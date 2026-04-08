package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
)

// TokenService Token 服务
type TokenService struct{}

// NewTokenService 创建 Token 服务
func NewTokenService() *TokenService {
	return &TokenService{}
}

// generateRandomToken 生成长度为 32 的随机字符串
func generateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CreateToken 创建 Token
func (s *TokenService) CreateToken(userID uint, description string, expiresInHours *int) (*model.Token, error) {
	db := store.GetDB()

	// 生成随机字符串作为 token 标识
	randomToken, err := generateRandomToken()
	if err != nil {
		return nil, err
	}

	token := &model.Token{
		UserID:      userID,
		Token:       randomToken,
		Description: description,
	}

	if expiresInHours != nil && *expiresInHours > 0 {
		expiresAt := time.Now().Add(time.Duration(*expiresInHours) * time.Hour)
		token.ExpiresAt = &expiresAt
	}

	if err := db.Create(token).Error; err != nil {
		return nil, err
	}

	return token, nil
}

// GetTokensByUserID 获取用户的所有 Token
func (s *TokenService) GetTokensByUserID(userID uint) ([]model.Token, error) {
	db := store.GetDB()
	var tokens []model.Token
	if err := db.Where("user_id = ?", userID).Order("created_at DESC").Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

// DeleteToken 删除 Token
func (s *TokenService) DeleteToken(tokenID, userID uint) error {
	db := store.GetDB()
	result := db.Where("id = ? AND user_id = ?", tokenID, userID).Delete(&model.Token{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("token 不存在或无权删除")
	}
	return nil
}

// ValidateToken 验证 Token 是否有效
func (s *TokenService) ValidateToken(tokenString string) (uint, error) {
	db := store.GetDB()

	// 先查找 token 记录
	var token model.Token
	if err := db.Where("token = ?", tokenString).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("token 不存在")
		}
		return 0, err
	}

	// 检查是否过期
	if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
		return 0, errors.New("token 已过期")
	}

	return token.UserID, nil
}
