package core

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// merge some candle sticks with smaller time span, into one candle stick with larger time span
func merge(subList []baseCandleStick) (cs baseCandleStick) {
	cs = NewBaseCandleStick()
	for _, sub := range subList {
		if !sub.hasDeal() {
			continue
		}
		if !cs.hasDeal() {
			cs.OpenPrice = sub.OpenPrice
		}
		cs.ClosePrice = sub.ClosePrice
		if !cs.hasDeal() {
			cs.HighPrice = sub.HighPrice
			cs.LowPrice = sub.LowPrice
		} else {
			if cs.HighPrice.LT(sub.HighPrice) {
				cs.HighPrice = sub.HighPrice
			}
			if cs.LowPrice.GT(sub.LowPrice) {
				cs.LowPrice = sub.LowPrice
			}
		}
		cs.TotalDeal = cs.TotalDeal.Add(sub.TotalDeal)
	}
	return
}

func getSpanStrFromSpan(span byte) string {
	switch span {
	case Minute:
		return MinuteStr
	case Hour:
		return HourStr
	case Day:
		return DayStr
	default:
		return ""
	}
}

func GetSpanFromSpanStr(spanStr string) byte {
	switch spanStr {
	case MinuteStr:
		return Minute
	case HourStr:
		return Hour
	case DayStr:
		return Day
	default:
		return 0
	}
}

func mergePrice(updated []*PricePoint, level string, isBuy bool) map[string]*PricePoint {
	p, err := sdk.NewDecFromStr(level)
	if err != nil {
		return nil
	}
	mulDec := sdk.OneDec().QuoTruncate(p)
	depth := make(map[string]*PricePoint)
	for _, point := range updated {
		updateAmount(depth, point, mulDec, isBuy)
	}
	return depth
}

func updateAmount(m map[string]*PricePoint, point *PricePoint, mulDec sdk.Dec, isBuy bool) {
	price := point.Price.Mul(mulDec).Ceil()
	if isBuy {
		price = point.Price.Mul(mulDec).TruncateDec()
	}
	price = price.Quo(mulDec)
	s := string(DecToBigEndianBytes(price))
	if val, ok := m[s]; ok {
		val.Amount = val.Amount.Add(point.Amount)
	} else {
		m[s] = &PricePoint{
			Price:  price,
			Amount: point.Amount,
		}
	}
}

func GetTopicAndParams(subscriptionTopic string) (topic string, params []string, err error) {
	values := strings.Split(subscriptionTopic, SeparateArgu)
	if len(values) < 1 || len(values) > MaxArguNum+1 {
		return "", nil, fmt.Errorf("Invalid params count : [%s] ", subscriptionTopic)
	}
	if !checkTopicValid(topic, params) {
		return "", nil, fmt.Errorf("The subscribed topic [%s] is illegal ", topic)
	}
	return values[0], values[1:], nil
}

func checkTopicValid(topic string, params []string) bool {
	switch topic {
	case BlockInfoKey, SlashKey:
		return len(params) == 0
	case TickerKey: // ticker:abc/cet; ticker:B:abc/cet
		if len(params) == 1 {
			return true
		}
		if len(params) == 2 {
			return params[0] == "B"
		}
	case UnbondingKey, RedelegationKey, LockedKey,
		UnlockKey, TxKey, IncomeKey, OrderKey, CommentKey,
		BancorTradeKey, BancorKey, DealKey, BancorDealKey:
		return len(params) == 1
	case KlineKey: // kline:abc/cet:1min; kline:B:abc/cet:1min
		if len(params) != 2 && len(params) != 3 {
			return false
		}
		timeSpan := params[1]
		if len(params) == 3 {
			if params[0] != "B" {
				return false
			}
			timeSpan = params[2]
		}
		switch timeSpan {
		case MinuteStr, HourStr, DayStr:
			return true
		}
	case DepthKey: //depth:<trading-pair>:<level>
		if len(params) == 1 {
			return true
		}
		if len(params) == 2 {
			switch params[1] {
			case "100", "10", "1", "0.1", "0.01", "0.001", "0.0001", "0.00001", "0.000001",
				"0.0000001", "0.00000001", "0.000000001", "0.0000000001",
				"0.00000000001", "0.000000000001", "0.0000000000001",
				"0.00000000000001", "0.000000000000001", "0.0000000000000001",
				"0.00000000000000001", "0.000000000000000001", "all":
				return true
			}
		}
	}
	return false
}

func getCount(count int) int {
	count = limitCount(count)
	if count == 0 {
		count = 10
	}
	return count
}
