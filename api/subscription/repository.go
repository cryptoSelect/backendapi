package subscription

import (
	"github.com/cryptoSelect/public/database"
	publicModels "github.com/cryptoSelect/public/models"
)

func CreateSubscription(userID uint, symbol, cycle string) error {
	sub := publicModels.Subscription{
		Symbol: symbol,
		Cycle:  cycle,
		UserID: userID,
	}
	// FirstOrCreate 避免重复订阅
	return database.DB.Where(publicModels.Subscription{UserID: userID, Symbol: symbol, Cycle: cycle}).FirstOrCreate(&sub).Error
}

// SubItem 订阅项
type SubItem struct {
	Symbol string `json:"symbol"`
	Cycle  string `json:"cycle"`
}

// ListByUserID 获取用户所有订阅（symbol+cycle）
func ListByUserID(userID uint) ([]SubItem, error) {
	return listByUserID(userID, "")
}

// ListByUserIDSymbol 获取用户指定 symbol 的订阅（返回 cycle 列表）
func ListByUserIDSymbol(userID uint, symbol string) ([]SubItem, error) {
	return listByUserID(userID, symbol)
}

func listByUserID(userID uint, symbol string) ([]SubItem, error) {
	query := database.DB.Where("user_id = ?", userID)
	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}
	var subs []publicModels.Subscription
	if err := query.Find(&subs).Error; err != nil {
		return nil, err
	}
	items := make([]SubItem, 0, len(subs))
	for _, s := range subs {
		if s.Symbol != "" && s.Cycle != "" {
			items = append(items, SubItem{Symbol: s.Symbol, Cycle: s.Cycle})
		}
	}
	return items, nil
}

// DeleteByUserIDSymbolCycle 删除用户的某条订阅
func DeleteByUserIDSymbolCycle(userID uint, symbol, cycle string) error {
	return database.DB.Where("user_id = ? AND symbol = ? AND cycle = ?", userID, symbol, cycle).Delete(&publicModels.Subscription{}).Error
}
