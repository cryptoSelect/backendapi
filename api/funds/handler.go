package funds

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Error string      `json:"error"`
	Code  int         `json:"code"`
	Data  interface{} `json:"data"`
}

// HandleTradeInflowQuery 查询资金流向数据
func HandleTradeInflowQuery(c *gin.Context) {
	// 获取 Query 参数
	symbol := c.Query("symbol")

	// symbol参数为必传
	if symbol == "" {
		c.JSON(400, Response{
			Error: "symbol parameter is required",
			Code:  400,
			Data:  ListData{Data: []map[string]interface{}{}, Count: 0},
		})
		return
	}

	// 不转换大小写，保持原始输入
	symbol = strings.TrimSpace(symbol)

	// 如果后4位为USDT则去掉，因为数据库中记录的不带USDT
	if len(symbol) >= 4 && strings.ToUpper(symbol[len(symbol)-4:]) == "USDT" {
		symbol = symbol[:len(symbol)-4]
	}

	dataList, err := GetTradeInflowRecord(c, symbol)
	if err != nil {
		c.JSON(200, Response{
			Error: err.Error(),
			Code:  500,
			Data:  ListData{Data: []map[string]interface{}{}, Count: 0}, // 返回空数组而不是nil
		})
		return
	}

	// 确保即使没有数据也返回空数组
	if dataList.Data == nil {
		dataList.Data = []map[string]interface{}{}
	}

	c.JSON(200, Response{
		Error: "",
		Code:  200,
		Data:  dataList,
	})
}
