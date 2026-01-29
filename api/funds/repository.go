package funds

import (
	"github.com/cryptoSelect/public/database"
	publicModels "github.com/cryptoSelect/public/models"
	"github.com/gin-gonic/gin"
)

// ListData 分页数据结构
type ListData struct {
	Data  []map[string]interface{} `json:"data"`
	Count int                      `json:"count"`
}

// GetTradeInflowRecord 查询资金流向记录
func GetTradeInflowRecord(c *gin.Context, symbol string) (data ListData, err error) {
	var records []publicModels.CoinTradeInflowDto

	query := database.DB.Model(&publicModels.CoinTradeInflowDto{})

	// 添加过滤条件
	if symbol != "" {
		// 精确查找，不转换大小写
		query = query.Where("symbol = ?", symbol)
	}

	// 查询所有数据
	err = query.Find(&records).Error
	if err != nil {
		return ListData{}, err
	}

	formatted := make([]map[string]interface{}, len(records))
	for i, r := range records {
		formatted[i] = map[string]interface{}{
			"id":                           r.ID,
			"symbol":                       r.Symbol,
			"time_particle_enum":           r.TimeParticleEnum,
			"time":                         r.Time,
			"stop":                         r.Stop,
			"stop_trade_inflow":            r.StopTradeInflow,
			"stop_trade_amount":            r.StopTradeAmount,
			"stop_trade_inflow_change":     r.StopTradeInflowChange,
			"stop_trade_amount_change":     r.StopTradeAmountChange,
			"contract":                     r.Contract,
			"contract_trade_inflow":        r.ContractTradeInflow,
			"contract_trade_amount":        r.ContractTradeAmount,
			"contract_trade_inflow_change": r.ContractTradeInflowChange,
			"contract_trade_amount_change": r.ContractTradeAmountChange,
			"stop_trade_in":                r.StopTradeIn,
			"stop_trade_out":               r.StopTradeOut,
			"contract_trade_in":            r.ContractTradeIn,
			"contract_trade_out":           r.ContractTradeOut,
		}
	}

	return ListData{
		Data:  formatted,
		Count: len(formatted),
	}, nil
}
