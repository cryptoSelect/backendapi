package user

import (
	"net/http"

	"github.com/cryptoSelect/backendapi/api/auth"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Error string      `json:"error"`
	Code  int         `json:"code"`
	Data  interface{} `json:"data"`
}

type MeData struct {
	Email         string `json:"email"`
	TelegramBound bool   `json:"telegram_bound"` // 是否已绑定 Telegram
}

// Me 返回当前登录用户信息（需在 RequireAuth 之后调用）
func Me(c *gin.Context) {
	email, _ := c.Get(auth.ContextUserEmailKey)
	userID, _ := c.Get(auth.ContextUserIDKey)
	tgID := auth.GetUserTelegramID(userID.(uint))
	c.JSON(http.StatusOK, Response{
		Error: "",
		Code:  200,
		Data:  MeData{Email: email.(string), TelegramBound: tgID != ""},
	})
}
