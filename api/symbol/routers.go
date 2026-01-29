package symbol

import (
	"github.com/gin-gonic/gin"
)

// SetupRoleRoutes registerRouter
func SetupSymbolRoutes(router *gin.RouterGroup) {

	// get all roles
	router.GET("/", HandleSymbolQuery)
}
