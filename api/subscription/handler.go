package subscription

import (
	"net/http"
	"strings"

	"github.com/cryptoSelect/backendapi/api/auth"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Error string      `json:"error"`
	Code  int         `json:"code"`
	Data  interface{} `json:"data"`
}

type CreateRequest struct {
	Symbol string   `json:"symbol" binding:"required"`
	Cycles []string `json:"cycles" binding:"required"` // 支持一次选择多个周期
}

// Create 创建订阅（需登录），symbol+cycles 与当前 user_id 写入 subscription 表，支持多周期
func Create(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	uid := userID.(uint)

	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, Response{Error: "symbol and cycles required", Code: 400, Data: nil})
		return
	}
	symbol := strings.TrimSpace(strings.ToUpper(req.Symbol))
	if symbol == "" || len(req.Cycles) == 0 {
		c.JSON(http.StatusOK, Response{Error: "symbol and at least one cycle required", Code: 400, Data: nil})
		return
	}
	seen := make(map[string]bool)
	for _, cy := range req.Cycles {
		cycle := strings.TrimSpace(cy)
		if cycle == "" || seen[cycle] {
			continue
		}
		seen[cycle] = true
		if err := CreateSubscription(uid, symbol, cycle); err != nil {
			c.JSON(http.StatusOK, Response{Error: err.Error(), Code: 500, Data: nil})
			return
		}
	}
	c.JSON(http.StatusOK, Response{Error: "", Code: 200, Data: gin.H{"ok": true}})
}

// List 获取当前用户订阅（需登录），返回 [{symbol, cycle}, ...]。可选 query symbol 筛选指定代币
func List(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	uid := userID.(uint)
	symbol := strings.TrimSpace(strings.ToUpper(c.Query("symbol")))
	var items []SubItem
	var err error
	if symbol != "" {
		items, err = ListByUserIDSymbol(uid, symbol)
	} else {
		items, err = ListByUserID(uid)
	}
	if err != nil {
		c.JSON(http.StatusOK, Response{Error: err.Error(), Code: 500, Data: nil})
		return
	}
	c.JSON(http.StatusOK, Response{Error: "", Code: 200, Data: gin.H{"data": items}})
}

// Delete 删除当前用户的某条订阅（需登录）
func Delete(c *gin.Context) {
	userID, _ := c.Get(auth.ContextUserIDKey)
	uid := userID.(uint)
	symbol := strings.TrimSpace(strings.ToUpper(c.Query("symbol")))
	cycle := strings.TrimSpace(c.Query("cycle"))
	if symbol == "" || cycle == "" {
		c.JSON(http.StatusOK, Response{Error: "symbol and cycle required", Code: 400, Data: nil})
		return
	}
	if err := DeleteByUserIDSymbolCycle(uid, symbol, cycle); err != nil {
		c.JSON(http.StatusOK, Response{Error: err.Error(), Code: 500, Data: nil})
		return
	}
	c.JSON(http.StatusOK, Response{Error: "", Code: 200, Data: gin.H{"ok": true}})
}
