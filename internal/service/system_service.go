package service

import (
	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
)

// SystemService 系统服务
type SystemService struct{}

// NewSystemService 创建系统服务
func NewSystemService() *SystemService {
	return &SystemService{}
}

// GetSetting 获取设置
func (s *SystemService) GetSetting(key string) (string, error) {
	db := store.GetDB()
	var setting model.SystemSetting
	if err := db.Where("setting_key = ?", key).First(&setting).Error; err != nil {
		return "", err
	}
	return setting.SettingValue, nil
}

// UpdateSetting 更新设置
func (s *SystemService) UpdateSetting(key, value string) error {
	db := store.GetDB()
	var setting model.SystemSetting
	if err := db.Where("setting_key = ?", key).First(&setting).Error; err != nil {
		// 不存在则创建
		setting = model.SystemSetting{SettingKey: key, SettingValue: value}
		return db.Create(&setting).Error
	}
	setting.SettingValue = value
	return db.Save(&setting).Error
}

// GetGravatarSource 获取 Gravatar 源地址
func GetGravatarSource() string {
	source, err := NewSystemService().GetSetting("gravatar_source")
	if err != nil || source == "" {
		return "https://weavatar.com/avatar/"
	}
	return source
}
