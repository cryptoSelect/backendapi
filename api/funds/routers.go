package funds

import (
	"github.com/gin-gonic/gin"
)

// SetupFundsRoutes 设置资金流向相关路由
func SetupFundsRoutes(group *gin.RouterGroup) {
	// 直接在传入的group上设置路由，而不是创建子组
	group.GET("/", HandleTradeInflowQuery) // 查询资金流向数据
}
