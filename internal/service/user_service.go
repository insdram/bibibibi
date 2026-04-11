package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/bibibibi/bibibibi/internal/model"
	"github.com/bibibibi/bibibibi/internal/store"
)



// UserService 用户服务
type UserService struct{}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{}
}

// Register 用户注册
func (s *UserService) Register(username, password, nickname, email string) (*model.User, error) {
	db := store.GetDB()

	// 检查注册是否开启
	var setting model.SystemSetting
	if err := db.Where("setting_key = ?", "registration_enabled").First(&setting).Error; err == nil {
		if setting.SettingValue == "false" {
			return nil, errors.New("注册已关闭")
		}
	}

	// 检查用户名是否已存在
	var existingUser model.User
	if err := db.Where("username = ?", username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查是否第一个用户
	var userCount int64
	db.Model(&model.User{}).Count(&userCount)
	isFirstUser := userCount == 0

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 生成 Gravatar 头像
	avatar := model.GetGravatarURLWithSource(email, GetGravatarSource())

	user := model.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
		Nickname: nickname,
		Avatar:   avatar,
		IsAdmin:  isFirstUser,
	}

	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// Login 用户登录
func (s *UserService) Login(username, password string) (string, *model.User, error) {
	db := store.GetDB()

	var user model.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("用户不存在")
		}
		return "", nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("密码错误")
	}

	// 生成 JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return tokenString, &user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	db := store.GetDB()
	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}
	// 使用系统设置的 Gravatar 源重新生成头像
	if user.Email != "" {
		user.Avatar = model.GetGravatarURLWithSource(user.Email, GetGravatarSource())
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(id uint, username, nickname, email, website, password string) (*model.User, error) {
	db := store.GetDB()
	var user model.User
	if err := db.First(&user, id).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 检查用户名是否被其他用户使用
	if username != user.Username {
		var existingUser model.User
		if err := db.Where("username = ? AND id != ?", username, id).First(&existingUser).Error; err == nil {
			return nil, errors.New("用户名已被使用")
		}
		user.Username = username
	}

	if nickname != "" {
		user.Nickname = nickname
	}

	if email != "" {
		user.Email = email
		user.Avatar = model.GetGravatarURLWithSource(email, GetGravatarSource())
	}

	if website != "" {
		user.Website = website
	}

	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	if err := db.Save(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUsers 获取所有用户
func (s *UserService) GetUsers() ([]model.User, error) {
	db := store.GetDB()
	var users []model.User
	if err := db.Select("id, username, nickname, email, is_admin, created_at").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
