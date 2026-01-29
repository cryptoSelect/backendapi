package symbol

import (
	"fmt"
	"strings"

	"github.com/cryptoSelect/public/database"
	publicModels "github.com/cryptoSelect/public/models"
	"github.com/gin-gonic/gin"
)

// 周期排序顺序：币种不为空且未选周期时，按此顺序排序
const cycleOrderSQL = "CASE cycle WHEN '5m' THEN 1 WHEN '15m' THEN 2 WHEN '30m' THEN 3 WHEN '1h' THEN 4 WHEN '4h' THEN 5 WHEN '1d' THEN 6 WHEN '1w' THEN 7 WHEN '1M' THEN 8 ELSE 99 END"

// 查询多条记录并分页
func GetSymbolRecord(c *gin.Context, symbol string, cycle string, shape int, rsi float64, rsiProvided bool, crossType int, page int, pageSize int, orderBy, orderType string) (data ListData, err error) {
	var records []publicModels.SymbolRecord
	var total int64

	query := database.DB.Model(&publicModels.SymbolRecord{})

	// 添加过滤条件
	if symbol != "" {
		// 检查symbol后4位转大写是否为USDT
		if len(symbol) >= 4 && strings.ToUpper(symbol[len(symbol)-4:]) == "USDT" {
			// 如果是USDT结尾，按精确名称查询，获取所有周期
			query = query.Where("symbol = ?", strings.ToUpper(symbol))
		} else {
			// 否则按模糊查询
			query = query.Where("symbol ILIKE ?", "%"+symbol+"%")
		}
	}
	if cycle != "" {
		// 如果指定了周期，仍然按周期过滤
		query = query.Where("cycle = ?", cycle)
	}
	if shape > 0 {
		query = query.Where("shape = ?", shape)
	}
	// RSI 规则：
	// - 仅当用户传了 rsi 参数时才过滤
	// - rsi >= 50：查询 rsi >= 参数
	// - rsi < 50：查询 rsi < 参数
	if rsiProvided {
		if rsi >= 50 {
			query = query.Where("rsi >= ?", rsi)
		} else {
			query = query.Where("rsi < ?", rsi)
		}
	}
	if crossType > 0 {
		query = query.Where("cross_type = ?", crossType)
	}

	// 获取总数
	err = query.Count(&total).Error
	if err != nil {
		return ListData{}, err
	}

	// 添加排序
	if symbol != "" && cycle == "" {
		// 币种不为空且未选周期：先按周期顺序(5m,15m,30m,1h,...)，再按更新时间倒序
		query = query.Order(cycleOrderSQL + ", updated_at DESC")
	} else if symbol != "" && len(symbol) >= 4 && strings.ToUpper(symbol[len(symbol)-4:]) == "USDT" {
		// 指定了周期时的 USDT 查询：按更新时间倒序
		query = query.Order("updated_at DESC")
	} else if orderBy != "" {
		// 否则按用户指定的排序
		orderDirection := "ASC"
		if orderType == "desc" {
			orderDirection = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", orderBy, orderDirection))
	} else {
		// 默认按更新时间降序排序
		query = query.Order("updated_at DESC")
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Find(&records).Error
	if err != nil {
		return ListData{}, err
	}

	formatted := make([]map[string]interface{}, len(records))
	for i, r := range records {
		fmt.Println("rate_cycle:", r.RateCycle)
		formatted[i] = map[string]interface{}{
			"id":               r.ID,
			"symbol":           r.Symbol,
			"cycle":            r.Cycle,
			"shape":            r.Shape,
			"rsi":              r.Rsi,
			"cross_type":       r.CrossType,
			"cross_time":       r.CrossTime,
			"price":            r.Price,
			"volume":           r.Volume,
			"taker_buy_volume": r.TakerBuyVolume,
			"taker_buy_ratio":  r.TakerBuyRatio,
			"rate":             r.Rate,
			"rate_cycle":       r.RateCycle,
			"change":           r.Change,
			"updated_at":       r.UpdatedAt,
			"vp_signal":        r.VpSignal,
			"smc_signal":       r.SMCSignal,
			"fvg":              r.Fvg,
			"ob":               r.Ob,
		}
	}

	return ListData{
		Data:  formatted,
		Count: int(total),
	}, nil
}
