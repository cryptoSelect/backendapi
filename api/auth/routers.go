package auth

import (
	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(router *gin.RouterGroup) {
	router.POST("/login", Login)
	router.POST("/register", Register)

	// Telegram 绑定：需先登录，start 需鉴权；status/confirm 无需鉴权（token 为密钥 / Bot 回调）
	router.POST("/tg/bind/start", RequireAuth, StartTelegramBind)
	router.GET("/tg/bind/status", TelegramBindStatus)
	router.POST("/tg/bind/confirm", ConfirmTelegramBind)
}
