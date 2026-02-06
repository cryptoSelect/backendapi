package user

import (
	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(router *gin.RouterGroup, requireAuth gin.HandlerFunc) {
	router.GET("/me", requireAuth, Me)
}
