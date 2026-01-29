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
	Mode     string     `json:"Mode"`
	Database DBConfig   `json:"Database"`
	Page     PageConfig `json:"Page"`
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
