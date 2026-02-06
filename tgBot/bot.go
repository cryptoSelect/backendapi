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

// Run 启动 Bot 长轮询，收到 /start <token> 时调用后端确认绑定
func Run(backendBase string) {
	token := strings.TrimSpace(config.Cfg.TelegramBotToken)
	if token == "" {
		return
	}
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
			confirmBind(confirmURL, bindToken, telegramID)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func confirmBind(confirmURL, token, telegramID string) {
	payload, _ := json.Marshal(map[string]string{"token": token, "telegram_id": telegramID})
	req, err := http.NewRequest("POST", confirmURL, bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// 忽略响应，仅触发绑定
}
