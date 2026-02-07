package auth

import (
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(router *gin.RouterGroup) {
	router.POST("/login", Login)
	router.POST("/register", Register)

	// Telegram 绑定：需先登录。Widget 方式点击即完成；start/status/confirm 为 Bot /start 兼容流程
	router.POST("/tg/bind/widget", RequireAuth, TelegramBindWidget)
	router.GET("/tg/bind/bot-name", RequireAuth, GetTelegramBotName)
	router.POST("/tg/bind/start", RequireAuth, StartTelegramBind)
	router.GET("/tg/bind/status", TelegramBindStatus)
	router.POST("/tg/bind/confirm", ConfirmTelegramBind)
}
