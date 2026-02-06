package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/cryptoSelect/backendapi/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	ContextUserIDKey    = "user_id"
	ContextUserEmailKey = "user_email"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthResponse struct {
	Token         string `json:"token"`
	Email         string `json:"email"`
	TelegramBound bool   `json:"telegram_bound"` // 是否已绑定 Telegram，前端据此决定是否展示绑定入口
}

type Response struct {
	Error string      `json:"error"`
	Code  int         `json:"code"`
	Data  interface{} `json:"data"`
}

type claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Login 邮箱+密码登录，返回 JWT
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{Error: "invalid_request", Code: 400, Data: nil})
		return
	}
	secret := config.Cfg.JWTSecret
	if secret == "" {
		c.JSON(http.StatusOK, Response{Error: "server_error", Code: 500, Data: nil})
		return
	}

	// 只按邮箱查用户；库里存的是 bcrypt 哈希，不能把明文密码放进 SQL
	user, err := UserLogin(req.Email)
	if err != nil {
		c.JSON(http.StatusOK, Response{Error: "email_or_password_wrong", Code: 401, Data: nil})
		return
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusOK, Response{Error: "email_or_password_wrong", Code: 401, Data: nil})
		return
	}

	token, err := issueToken(secret, user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusOK, Response{Error: "server_error", Code: 500, Data: nil})
		return
	}
	c.JSON(http.StatusOK, Response{Error: "", Code: 200, Data: AuthResponse{Token: token, Email: user.Email, TelegramBound: strings.TrimSpace(user.TelegramID) != ""}})
}

// Register 注册新用户
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{Error: "invalid_request", Code: 400, Data: nil})
		return
	}
	secret := config.Cfg.JWTSecret
	if secret == "" {
		c.JSON(http.StatusOK, Response{Error: "server_error", Code: 500, Data: nil})
		return
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if EmailExists(email) {
		c.JSON(http.StatusOK, Response{Error: "email_already_registered", Code: 400, Data: nil})
		return
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusOK, Response{Error: "server_error", Code: 500, Data: nil})
		return
	}
	user, err := CreateUserWithEmail(email, string(hashed), "")
	if err != nil {
		c.JSON(http.StatusOK, Response{Error: "server_error", Code: 500, Data: nil})
		return
	}
	token, err := issueToken(secret, user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusOK, Response{Error: "server_error", Code: 500, Data: nil})
		return
	}
	c.JSON(http.StatusOK, Response{Error: "", Code: 200, Data: AuthResponse{Token: token, Email: user.Email, TelegramBound: false}})
}

func issueToken(secret string, userID uint, email string) (string, error) {
	exp := time.Now().Add(7 * 24 * time.Hour)
	claims := claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// RequireAuth 从 Authorization: Bearer <token> 解析 JWT，将 user_id、user_email 写入 context；未登录返回 401
func RequireAuth(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		c.JSON(http.StatusOK, Response{Error: "unauthorized", Code: 401, Data: nil})
		c.Abort()
		return
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.JSON(http.StatusOK, Response{Error: "unauthorized", Code: 401, Data: nil})
		c.Abort()
		return
	}
	tokenString := strings.TrimSpace(parts[1])
	if tokenString == "" {
		c.JSON(http.StatusOK, Response{Error: "unauthorized", Code: 401, Data: nil})
		c.Abort()
		return
	}
	secret := config.Cfg.JWTSecret
	if secret == "" {
		c.JSON(http.StatusOK, Response{Error: "server auth not configured", Code: 500, Data: nil})
		c.Abort()
		return
	}
	var cl claims
	token, err := jwt.ParseWithClaims(tokenString, &cl, func(*jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusOK, Response{Error: "unauthorized", Code: 401, Data: nil})
		c.Abort()
		return
	}
	c.Set(ContextUserIDKey, cl.UserID)
	c.Set(ContextUserEmailKey, cl.Email)
	c.Next()
}
