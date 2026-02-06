package main

import (
	"github.com/cryptoSelect/backendapi/api/auth"
	"github.com/cryptoSelect/backendapi/api/funds"
	"github.com/cryptoSelect/backendapi/api/subscription"
	"github.com/cryptoSelect/backendapi/api/symbol"
	"github.com/cryptoSelect/backendapi/api/user"
	"github.com/cryptoSelect/backendapi/config"
	"github.com/cryptoSelect/backendapi/tgBot"
	"github.com/cryptoSelect/backendapi/utils/logger"

	"github.com/cryptoSelect/public/database"
	"github.com/cryptoSelect/public/models"
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

	// 迁移用户与订阅表（与 public 库同库）
	if err := database.AutoMigrate(&models.UserInfo{}, &models.Subscription{}); err != nil {
		logger.Log.Error("migrate user/subscription failed", map[string]interface{}{"error": err.Error()})
	}

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

	// AuthRoutes（登录/注册）
	AuthRoutes := api.Group("/auth")
	auth.SetupAuthRoutes(AuthRoutes)

	// UserRoutes（需登录）
	UserRoutes := api.Group("/user")
	user.SetupUserRoutes(UserRoutes, auth.RequireAuth)

	// SubscriptionRoutes（需登录）
	SubRoutes := api.Group("/subscription")
	subscription.SetupSubscriptionRoutes(SubRoutes, auth.RequireAuth)

	// Telegram Bot 长轮询，收到 /start <token> 时回调 ConfirmTelegramBind
	go tgBot.Run(config.Cfg.BackendAPIBase)

	logger.Log.Info("Starting API server", map[string]interface{}{"port": ":8080", "mode": config.Cfg.Mode})
	if err := r.Run(":8080"); err != nil {
		logger.Log.Error("API Server failed to start", map[string]interface{}{"error": err.Error()})
		panic("API Server failed to start: " + err.Error())
	}
}
