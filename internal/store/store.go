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
	)
	if err != nil {
		return err
	}

	// 初始化系统设置
	var setting model.SystemSetting
	if err := DB.Where("setting_key = ?", "registration_enabled").First(&setting).Error; err != nil {
		// 默认开启注册
		DB.Create(&model.SystemSetting{SettingKey: "registration_enabled", SettingValue: "true"})
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
