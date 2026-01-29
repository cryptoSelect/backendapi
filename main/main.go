package main

import (
	"github.com/cryptoSelect/backendapi/api/funds"
	"github.com/cryptoSelect/backendapi/api/symbol"
	"github.com/cryptoSelect/backendapi/config"
	"github.com/cryptoSelect/backendapi/utils/logger"

	"github.com/cryptoSelect/public/database"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	config.Init()
	logger.Init(config.Cfg.Mode) // 使用配置文件中的模式
	database.InitDB(
		config.Cfg.Database.Host,
		config.Cfg.Database.User,
		config.Cfg.Database.Password,
		config.Cfg.Database.DBName,
		config.Cfg.Database.Port,
	)

	// 根据配置设置 GIN 模式
	if config.Cfg.Mode == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	// init-gin
	r := gin.Default()
	r.Use(cors.Default())

	// api
	api := r.Group("/api")

	// SymbolRoutes
	SymbolRoutes := api.Group("/symbol")
	symbol.SetupSymbolRoutes(SymbolRoutes)

	// FundsRoutes
	FundsRoutes := api.Group("/funds")
	funds.SetupFundsRoutes(FundsRoutes)

	logger.Log.Info("Starting API server", map[string]interface{}{"port": ":8080", "mode": config.Cfg.Mode})
	if err := r.Run(":8080"); err != nil {
		logger.Log.Error("API Server failed to start", map[string]interface{}{"error": err.Error()})
		panic("API Server failed to start: " + err.Error())
	}
}
