package core

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	log "github.com/sirupsen/logrus"
)

// -----------
// convert string and byte in timespan

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

// ----------------
// price and amount

// merge some candle sticks with smaller time span, into one candle stick with larger time span
func merge(subList []baseCandleStick) (cs baseCandleStick) {
	cs = newBaseCandleStick()
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

// Merge the PricePoints at the granularity defined by 'level'
// Returns a depth map
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

// 'm' keeps a depth map, use 'point' to update it at the granularity defined by 'mulDec'
func updateAmount(m map[string]*PricePoint, point *PricePoint, mulDec sdk.Dec, isBuy bool) {
	price := point.Price.Mul(mulDec).Ceil()
	if isBuy {
		price = point.Price.Mul(mulDec).TruncateDec()
	}
	price = price.Quo(mulDec)
	s := string(decToBigEndianBytes(price))
	if val, ok := m[s]; ok {
		val.Amount = val.Amount.Add(point.Amount)
	} else {
		m[s] = &PricePoint{
			Price:  price,
			Amount: point.Amount,
		}
	}
}

func decToBigEndianBytes(d sdk.Dec) []byte {
	var result [DecByteCount]byte
	bytes := d.Int.Bytes() //  returns the absolute value of d as a big-endian byte slice.
	for i := 1; i <= len(bytes); i++ {
		result[DecByteCount-i] = bytes[len(bytes)-i]
	}
	return result[:]
}

// -------------
// websocket topic

func GetTopicAndParams(subscriptionTopic string) (topic string, params []string, err error) {
	values := strings.Split(subscriptionTopic, SeparateArgu)
	if len(values) < 1 || len(values) > MaxArguNum+1 {
		return "", nil, fmt.Errorf("Invalid params count : [%s] ", subscriptionTopic)
	}
	if !checkTopicValid(values[0], values[1:]) {
		return "", nil, fmt.Errorf("The subscribed topic [%s] is illegal ", subscriptionTopic)
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

// -----------
// encode data
type LevelsPricePoint map[string]map[string]*PricePoint // [level][priceBigEndian][*PricePoint]

func encodeDepthLevels(market string, buyDepths, sellDepths LevelsPricePoint) map[string][]byte {
	levelsData := make(map[string][]byte)
	if len(buyDepths) == 0 && len(sellDepths) == 0 {
		return levelsData
	}
	for level, depth := range buyDepths {
		if sell, ok := sellDepths[level]; ok {
			if bz, err := encodeDepthLevel(market, depth, sell); err == nil {
				levelsData[level] = bz
			}
		} else {
			if bz, err := encodeDepthLevel(market, depth, nil); err == nil {
				levelsData[level] = bz
			}
		}
	}
	return levelsData
}

func encodeDepthLevel(market string, buyDepth map[string]*PricePoint, sellDepth map[string]*PricePoint) ([]byte, error) {
	buyValues := make([]*PricePoint, 0, len(buyDepth))
	sellValues := make([]*PricePoint, 0, len(sellDepth))
	for _, p := range buyDepth {
		buyValues = append(buyValues, p)
	}
	for _, p := range sellDepth {
		sellValues = append(sellValues, p)
	}
	if len(buyValues) > 1 {
		sort.Slice(buyValues, func(i, j int) bool {
			return buyValues[i].Price.GT(buyValues[j].Price)
		})
	}
	if len(sellValues) > 1 {
		sort.Slice(sellValues, func(i, j int) bool {
			return sellValues[i].Price.LT(sellValues[j].Price)
		})
	}
	bz, err := encodeDepthData(market, buyValues, sellValues)
	return bz, err
}

func encodeDepthData(market string, buy, sell []*PricePoint) ([]byte, error) {
	depRes := DepthDetails{
		TradingPair: market,
		Bids:        buy,
		Asks:        sell,
	}
	bz, err := json.Marshal(depRes)
	return bz, err
}

func groupOfDataPacket(topic string, data []json.RawMessage) []byte {
	bz := make([]byte, 0, len(data))
	bz = append(bz, []byte(fmt.Sprintf("{\"type\":\"%s\", \"payload\":[", topic))...)
	for _, v := range data {
		bz = append(bz, []byte(v)...)
		bz = append(bz, []byte(",")...)
	}
	if len(data) > 0 {
		bz[len(bz)-1] = byte(']')
		bz = append(bz, []byte("}")...)
	} else {
		bz = append(bz, []byte("]}")...)
	}
	return bz
}

func Int64ToBigEndianBytes(n int64) []byte {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(n))
	return b[:]
}

func BigEndianBytesToInt64(bz []byte) int64 {
	val := binary.BigEndian.Uint64(bz)
	return int64(val)
}

// ---------------
// key in Hub

func getCandleStickStartKey(market string, timespan byte) []byte {
	bz := append([]byte(market), []byte{0, timespan}...)
	return getStartKeyFromBytes(CandleStickByte, bz)
}

func getCandleStickEndKey(market string, timespan byte, endTime int64, sid int64) []byte {
	bz := append([]byte(market), []byte{0, timespan}...)
	return getEndKeyFromBytes(CandleStickByte, bz, endTime, sid)
}

func getStartKeyFromBytes(firstByte byte, bz []byte) []byte {
	res := make([]byte, 0, 1+1+len(bz)+1)
	res = append(res, firstByte)
	res = append(res, byte(len(bz)))
	res = append(res, bz...)
	res = append(res, byte(0))
	return res
}

func getEndKeyFromBytes(firstByte byte, bz []byte, time int64, sid int64) []byte {
	res := make([]byte, 0, 1+1+len(bz)+1+16)
	res = append(res, firstByte)
	res = append(res, byte(len(bz)))
	res = append(res, bz...)
	res = append(res, byte(0))
	res = append(res, Int64ToBigEndianBytes(time)...)
	res = append(res, Int64ToBigEndianBytes(sid)...)
	return res
}

func GetOffsetKey(partition int32) []byte {
	key := make([]byte, 5)
	key[0] = OffsetByte
	binary.BigEndian.PutUint32(key[1:], uint32(partition))
	return key
}

func GetDumpKey() []byte {
	key := make([]byte, 2)
	key[0] = DumpByte
	key[1] = DumpVersion
	return key
}

// --------------
// other function in Hub

func limitCount(count int) int {
	if count > MaxCount {
		return MaxCount
	}
	return count
}

func getTokenNameFromAmount(amount string) string {
	tokenName := ""
	for i, c := range amount {
		if c < '0' || c > '9' {
			tokenName = amount[i:]
			break
		}
	}
	return tokenName
}

func getMarketName(info MarketInfo) string {
	return info.Stock + "/" + info.Money
}

func unmarshalAndLogErr(bz []byte, v interface{}) error {
	err := json.Unmarshal(bz, v)
	if err != nil {
		log.Error("Err in Unmarshal ", reflect.TypeOf(v).String())
	}
	return err
}

// Append tx_hash into a json string
func appendHashID(bz []byte, hashID string) []byte {
	if len(hashID) == 0 {
		return bz
	}
	bz = bz[0 : len(bz)-1] // Delete the last character: '}'
	return append(bz, []byte(fmt.Sprintf(`,"tx_hash":"%s"}`, hashID))...)
}

func formatCandleStick(info *CandleStick) []byte {
	bz, err := json.Marshal(info)
	if err != nil {
		log.Errorf(err.Error())
		return nil
	}
	return bz
}

func getChainID(bz []byte) string {
	type ChainID struct {
		ChainID string `json:"chain_id"`
	}
	var id ChainID
	if err := json.Unmarshal(bz, &id); err != nil {
		log.Error("unmarshal chaid failed")
		return ""
	}
	return id.ChainID
}
