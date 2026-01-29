package symbol

import (
	"strconv"
	"strings"

	"github.com/cryptoSelect/backendapi/config"
	"github.com/gin-gonic/gin"
)

// 通用响应结构
type Response struct {
	Error string      `json:"error"`
	Code  int         `json:"code"`
	Data  interface{} `json:"data"`
}

// 列表数据负载
type ListData struct {
	Data  []map[string]interface{} `json:"data"`
	Count int                      `json:"count"`
}

func HandleSymbolQuery(c *gin.Context) {
	// 获取 Query 参数
	symbol := c.Query("symbol")
	cycle := c.Query("cycle")
	shapeStr := c.Query("shape")
	rsiStr := c.Query("rsi")
	crossTypeStr := c.Query("cross_type")

	// 转换参数
	var shape int
	if shapeStr != "" {
		shape, _ = strconv.Atoi(shapeStr)
	}

	var rsi float64
	rsiProvided := false
	if rsiStr != "" {
		rsi, _ = strconv.ParseFloat(rsiStr, 64)
		rsiProvided = true
	}

	var crossType int
	if crossTypeStr != "" {
		crossType, _ = strconv.Atoi(crossTypeStr)
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}
	pageSize := config.Cfg.Page.PageSize // 用户要求每页10条

	// 解析排序参数
	orderBy := strings.ToLower(c.DefaultQuery("order_by", "updated_at"))
	orderType := strings.ToLower(c.DefaultQuery("order_type", "desc"))

	// 校验和映射排序字段，防止 SQL 注入
	allowedOrders := map[string]string{
		"updated_at": "updated_at",
		"rsi":        "rsi",
		"change":     "change",
		"symbol":     "symbol",
	}

	if field, ok := allowedOrders[orderBy]; ok {
		orderBy = field
	} else {
		orderBy = "updated_at"
	}

	if orderType != "asc" && orderType != "desc" {
		orderType = "desc"
	}

	if symbol != "" {
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
	}

	dataList, err := GetSymbolRecord(c, symbol, cycle, shape, rsi, rsiProvided, crossType, page, pageSize, orderBy, orderType)
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
