package store

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/bibibibi/bibibibi/internal/model"
)

var DB *gorm.DB

// InitDB 初始化数据库
func InitDB(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	// 自动迁移数据库
	err = DB.AutoMigrate(
		&model.User{},
		&model.Bibi{},
		&model.Tag{},
		&model.BibiTag{},
		&model.Comment{},
		&model.Like{},
		&model.SystemSetting{},
		&model.Token{},
		&model.FeedSource{},
	)
	if err != nil {
		return err
	}

	// 初始化系统设置
	var setting model.SystemSetting
	if err := DB.Where("setting_key = ?", "registration_enabled").First(&setting).Error; err != nil {
		DB.Create(&model.SystemSetting{SettingKey: "registration_enabled", SettingValue: "true"})
	}

	// 初始化 Gravatar 源（默认使用 weavatar 镜像）
	var gravatarSource model.SystemSetting
	if err := DB.Where("setting_key = ?", "gravatar_source").First(&gravatarSource).Error; err != nil {
		DB.Create(&model.SystemSetting{SettingKey: "gravatar_source", SettingValue: "https://weavatar.com/avatar/"})
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
