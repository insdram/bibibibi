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

	// 创建复合索引（SQLite 支持在 AutoMigrate 后手动创建）
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_bibi_visibility_created ON bibis(visibility, created_at)")
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_comment_bibi_created ON comments(bibi_id, created_at)")

	// 初始化系统设置（使用 FirstOrCreate 避免重复检查）
	DB.Where(model.SystemSetting{SettingKey: "registration_enabled"}).Assign(model.SystemSetting{SettingValue: "true"}).FirstOrCreate(&model.SystemSetting{})
	DB.Where(model.SystemSetting{SettingKey: "gravatar_source"}).Assign(model.SystemSetting{SettingValue: "https://weavatar.com/avatar/"}).FirstOrCreate(&model.SystemSetting{})

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
