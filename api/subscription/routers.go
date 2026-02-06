package subscription

import (
	"github.com/gin-gonic/gin"
)

func SetupSubscriptionRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	router.GET("/", requireAuth, List)
	router.POST("/", requireAuth, Create)
	router.DELETE("/", requireAuth, Delete)
}
