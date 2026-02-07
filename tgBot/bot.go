// Package tgBot 接收用户 /start 消息，解析绑定 token，调用后端 ConfirmTelegramBind
package tgBot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cryptoSelect/backendapi/config"
	"github.com/cryptoSelect/backendapi/utils/logger"
)

// Telegram Update 结构（简化）
type tgUpdate struct {
	UpdateID int `json:"update_id"`
	Message  *struct {
		MessageID int `json:"message_id"`
		From      *struct {
			ID        int64  `json:"id"`
			Username  string `json:"username"`
			FirstName string `json:"first_name"`
		} `json:"from"`
		Text string `json:"text"`
	} `json:"message"`
}

type tgUpdatesResp struct {
	OK     bool       `json:"ok"`
	Result []tgUpdate `json:"result"`
}

// deleteWebhook 删除 webhook，否则 getUpdates 收不到消息
func deleteWebhook(token string) {
	url := "https://api.telegram.org/bot" + token + "/deleteWebhook"
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	resp.Body.Close()
}

// Run 启动 Bot 长轮询，收到 /start <token> 时调用后端确认绑定
func Run(backendBase string) {
	token := strings.TrimSpace(config.Cfg.TelegramBotToken)
	if token == "" {
		logger.Log.Info("tgBot skip: TelegramBotToken not configured", nil)
		return
	}
	logger.Log.Info("tgBot starting", map[string]interface{}{"backend": backendBase})
	// 删除 webhook，否则 getUpdates 无法接收消息
	deleteWebhook(token)

	base := strings.TrimRight(backendBase, "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	apiURL := "https://api.telegram.org/bot" + token + "/getUpdates"
	confirmURL := base + "/api/auth/tg/bind/confirm"

	var offset int
	for {
		req, _ := http.NewRequest("GET", apiURL+"?offset="+fmt.Sprint(offset)+"&timeout=30", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var data tgUpdatesResp
		if json.Unmarshal(body, &data) != nil || !data.OK {
			time.Sleep(5 * time.Second)
			continue
		}

		for _, u := range data.Result {
			offset = u.UpdateID + 1
			if u.Message == nil || u.Message.From == nil {
				continue
			}
			text := strings.TrimSpace(u.Message.Text)
			if !strings.HasPrefix(strings.ToLower(text), "/start") {
				continue
			}
			// /start 或 /start <bindToken>
			parts := strings.Fields(text)
			bindToken := ""
			if len(parts) >= 2 {
				bindToken = strings.TrimSpace(parts[1])
			}
			if bindToken == "" {
				continue
			}
			telegramID := fmt.Sprintf("%d", u.Message.From.ID)
			tokenPreview := bindToken
			if len(tokenPreview) > 8 {
				tokenPreview = tokenPreview[:8] + "..."
			}
			logger.Log.Info("tgBot received /start", map[string]interface{}{"telegram_id": telegramID, "token": tokenPreview})
			confirmBind(confirmURL, bindToken, telegramID)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func confirmBind(confirmURL, token, telegramID string) {
	payload, _ := json.Marshal(map[string]string{"token": token, "telegram_id": telegramID})
	req, err := http.NewRequest("POST", confirmURL, bytes.NewReader(payload))
	if err != nil {
		logger.Log.Error("tgBot confirmBind newRequest failed", map[string]interface{}{"error": err.Error()})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log.Error("tgBot confirmBind request failed", map[string]interface{}{"url": confirmURL, "error": err.Error()})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Log.Error("tgBot confirmBind response non-200", map[string]interface{}{"status": resp.StatusCode, "body": string(body)})
		return
	}
	logger.Log.Info("tgBot confirmBind success", map[string]interface{}{"telegram_id": telegramID})
}
