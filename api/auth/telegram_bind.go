package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/cryptoSelect/backendapi/config"
	"github.com/gin-gonic/gin"
)

// 说明：
// - 用户先注册，登录后再绑定 Telegram。StartTelegramBind 需登录，token 与 user_id 关联。
// - Bot 收到 /start 后调用 ConfirmTelegramBind，后端根据 token 找到 user_id，将 telegram_id 写入该用户。
// - 多人同时绑定时，每人有独立 token，Bot 收到的每条 /start 对应唯一 user_id。
// - 当前实现使用内存保存绑定状态：适合单实例/本地；多实例部署请替换为 Redis/DB。

type tgBindRecord struct {
	UserID     uint   // 发起绑定的用户 ID，用于区分不同用户
	TelegramID string // Bot 确认后写入
	ExpiresAt  time.Time
}

var (
	tgBindMu         sync.Mutex
	tgBindByToken    = map[string]tgBindRecord{}
	tgBotNameCache   string
	tgBotNameCacheMu sync.Mutex
)

type TgBindStartResp struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"` // 秒
	BotName   string `json:"bot_name,omitempty"`
	StartURL  string `json:"start_url,omitempty"`
}

// StartTelegramBind 生成一次性绑定 token（需登录，token 与当前 user_id 关联）
func StartTelegramBind(c *gin.Context) {
	userIDVal, ok := c.Get(ContextUserIDKey)
	if !ok {
		c.JSON(http.StatusOK, Response{Error: "unauthorized", Code: 401, Data: nil})
		return
	}
	userID := userIDVal.(uint)

	ttl := 5 * time.Minute
	token, err := newBindToken()
	if err != nil {
		c.JSON(http.StatusOK, Response{Error: "failed to generate token", Code: 500, Data: nil})
		return
	}

	now := time.Now()
	tgBindMu.Lock()
	cleanupExpiredLocked(now)
	tgBindByToken[token] = tgBindRecord{UserID: userID, TelegramID: "", ExpiresAt: now.Add(ttl)}
	tgBindMu.Unlock()

	bot := resolveBotName()
	startURL := ""
	if bot != "" {
		startURL = "https://t.me/" + bot + "?start=" + token
	}
	c.JSON(http.StatusOK, Response{Error: "", Code: 200, Data: TgBindStartResp{
		Token:     token,
		ExpiresIn: int64(ttl.Seconds()),
		BotName:   bot,
		StartURL:  startURL,
	}})
}

type TgBindStatusResp struct {
	Bound      bool   `json:"bound"`
	TelegramID string `json:"telegram_id,omitempty"`
}

// TelegramBindStatus 查询 token 是否已被 Bot 绑定 telegram_id
func TelegramBindStatus(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusOK, Response{Error: "token required", Code: 400, Data: nil})
		return
	}

	now := time.Now()
	tgBindMu.Lock()
	cleanupExpiredLocked(now)
	rec, ok := tgBindByToken[token]
	tgBindMu.Unlock()

	if !ok {
		c.JSON(http.StatusOK, Response{Error: "token not found or expired", Code: 404, Data: TgBindStatusResp{Bound: false}})
		return
	}
	if rec.TelegramID == "" {
		c.JSON(http.StatusOK, Response{Error: "", Code: 200, Data: TgBindStatusResp{Bound: false}})
		return
	}
	c.JSON(http.StatusOK, Response{Error: "", Code: 200, Data: TgBindStatusResp{Bound: true, TelegramID: rec.TelegramID}})
}

type TgBindConfirmReq struct {
	Token      string `json:"token" binding:"required"`
	TelegramID string `json:"telegram_id" binding:"required"`
}

// ConfirmTelegramBind Bot 回调：根据 token 找到 user_id，将 telegram_id 写入该用户
func ConfirmTelegramBind(c *gin.Context) {
	var req TgBindConfirmReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{Error: "invalid request", Code: 400, Data: nil})
		return
	}

	now := time.Now()
	tgBindMu.Lock()
	cleanupExpiredLocked(now)
	rec, ok := tgBindByToken[req.Token]
	if !ok {
		tgBindMu.Unlock()
		c.JSON(http.StatusOK, Response{Error: "token not found or expired", Code: 404, Data: nil})
		return
	}
	userID := rec.UserID
	rec.TelegramID = req.TelegramID
	tgBindByToken[req.Token] = rec
	tgBindMu.Unlock()

	if err := UpdateUserTelegramID(userID, req.TelegramID); err != nil {
		c.JSON(http.StatusOK, Response{Error: "failed to update user", Code: 500, Data: nil})
		return
	}

	c.JSON(http.StatusOK, Response{Error: "", Code: 200, Data: gin.H{"ok": true}})
}

func cleanupExpiredLocked(now time.Time) {
	for k, v := range tgBindByToken {
		if now.After(v.ExpiresAt) {
			delete(tgBindByToken, k)
		}
	}
}

func newBindToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// resolveBotName 返回 Bot 用户名：优先 config.TelegramBotName，否则用 TelegramBotToken 调用 getMe 获取并缓存
func resolveBotName() string {
	name := strings.TrimSpace(config.Cfg.TelegramBotName)
	if name != "" {
		return name
	}
	token := strings.TrimSpace(config.Cfg.TelegramBotToken)
	if token == "" {
		return ""
	}
	tgBotNameCacheMu.Lock()
	defer tgBotNameCacheMu.Unlock()
	if tgBotNameCache != "" {
		return tgBotNameCache
	}
	// 调用 Telegram getMe 获取 bot 用户名
	url := "https://api.telegram.org/bot" + token + "/getMe"
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}
	if json.Unmarshal(body, &result) != nil || !result.OK || result.Result.Username == "" {
		return ""
	}
	tgBotNameCache = result.Result.Username
	return tgBotNameCache
}
