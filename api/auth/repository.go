package auth

import (
	"strings"

	"github.com/cryptoSelect/public/database"
	publicModels "github.com/cryptoSelect/public/models"
)

func CreateUser(user *publicModels.UserInfo) error {
	if err := database.DB.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// CreateUserWithEmail 创建用户（邮箱、已哈希密码、TelegramID），返回创建后的用户
func CreateUserWithEmail(email, hashedPassword, telegramID string) (*publicModels.UserInfo, error) {
	user := publicModels.UserInfo{
		Email:      email,
		Password:   hashedPassword,
		TelegramID: strings.TrimSpace(telegramID),
	}
	if err := database.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// EmailExists 检查邮箱是否已注册
func EmailExists(email string) bool {
	var user publicModels.UserInfo
	return database.DB.Where("email = ?", email).First(&user).Error == nil
}

// UserLogin 按邮箱查询用户，仅用于登录时取回用户（含密码哈希）；密码校验在 handler 里用 bcrypt.CompareHashAndPassword
func UserLogin(email string) (*publicModels.UserInfo, error) {
	var user publicModels.UserInfo
	if err := database.DB.Where("email = ?", strings.TrimSpace(strings.ToLower(email))).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserTelegramID 更新用户的 telegram_id
func UpdateUserTelegramID(userID uint, telegramID string) error {
	return database.DB.Model(&publicModels.UserInfo{}).Where("id = ?", userID).Update("telegram_id", strings.TrimSpace(telegramID)).Error
}

// GetUserTelegramID 获取用户的 telegram_id
func GetUserTelegramID(userID uint) string {
	var user publicModels.UserInfo
	if database.DB.Select("telegram_id").Where("id = ?", userID).First(&user).Error != nil {
		return ""
	}
	return strings.TrimSpace(user.TelegramID)
}
