package config

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
)

var Cfg *ServerConfig

// DBConfig 数据库配置结构 (临时定义，等待 public 模块更新)
type DBConfig struct {
	Host     string `json:"Host"`
	Port     int    `json:"Port"`
	User     string `json:"User"`
	Password string `json:"Password"`
	DBName   string `json:"DBName"`
	SSLMode  string `json:"SSLMode"`
}

type ServerConfig struct {
	Mode      string     `json:"Mode"`
	Database  DBConfig   `json:"Database"`
	Page      PageConfig `json:"Page"`
	JWTSecret string     `json:"JWTSecret"` // JWT 签发与校验密钥，生产环境务必修改
	// TelegramBotName 用于生成绑定链接：https://t.me/<bot>?start=<token>，填 Bot 用户名（不含 @）
	TelegramBotName string `json:"TelegramBotName"`
	// TelegramBotToken Bot API token，用于 getMe 与 tgBot 接收 /start
	TelegramBotToken string `json:"TelegramBotToken"`
	// BackendAPIBase tgBot 回调 ConfirmTelegramBind 的地址，如 http://localhost:8080
	BackendAPIBase string `json:"BackendAPIBase"`
}

type PageConfig struct {
	PageSize int `json:"PageSize"`
}

func LoadConfig(configNmae string) {
	wd, _ := os.Getwd()
	configPath := filepath.Join(wd, "config", configNmae)
	file, err := os.ReadFile(configPath)
	if err != nil {
		panic("config  file error:" + err.Error())
	}

	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))
	var tmp ServerConfig
	if err := json.Unmarshal(file, &tmp); err != nil {
		panic("unmarshal json config err:" + err.Error())
	}
	Cfg = &tmp
}

func Init() {
	LoadConfig("config.json")
}
