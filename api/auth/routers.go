package auth

import (
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(router *gin.RouterGroup) {
	router.POST("/login", Login)
	router.POST("/register", Register)

	// Telegram 绑定：需先登录，弹窗获取链接 -> 打开 Telegram 发送 /start -> Bot 回调 confirm，成功后发「绑定成功」消息
	router.POST("/tg/bind/start", RequireAuth, StartTelegramBind)
	router.GET("/tg/bind/status", TelegramBindStatus)
	router.POST("/tg/bind/confirm", ConfirmTelegramBind)
}
