package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// When a client subscribes something, we get the full information from rocksdb and push them to client.
// After that, we begin to push delta/incremental information to clients

func pushFullInformation(topic string, params []string, count int, c Subscriber, hub *Hub) error {
	var (
		err        error
		depthLevel = "all"
	)
	count = getCount(count)
	if len(params) == 2 && topic == DepthKey {
		depthLevel = params[1]
	}
	switch topic {
	case SlashKey:
		err = querySlashAndPush(hub, c, count)
	case KlineKey:
		err = queryKlineAndpush(hub, c, params, count)
	case DepthKey:
		err = queryDepthAndPush(hub, c, params[0], depthLevel, count)
	case OrderKey:
		err = queryOrderAndPush(hub, c, params[0], count)
	case TickerKey:
		market := params[0]
		if len(params) == 2 {
			market = strings.Join(params, SeparateArgu)
		}
		err = queryTickerAndPush(hub, c, market)
	case TxKey:
		err = queryAndPushFunc(hub, c, TxKey, params[0], count, hub.QueryTx)
	case LockedKey:
		err = queryAndPushFunc(hub, c, LockedKey, params[0], count, hub.QueryLocked)
	case UnlockKey:
		err = queryAndPushFunc(hub, c, UnlockKey, params[0], count, hub.QueryUnlock)
	case IncomeKey:
		err = queryAndPushFunc(hub, c, IncomeKey, params[0], count, hub.QueryIncome)
	case DealKey:
		err = queryAndPushFunc(hub, c, DealKey, params[0], count, hub.QueryDeal)
	case BancorKey:
		err = queryAndPushFunc(hub, c, BancorKey, params[0], count, hub.QueryBancorInfo)
	case BancorTradeKey:
		err = queryAndPushFunc(hub, c, BancorTradeKey, params[0], count, hub.QueryBancorTrade)
	case BancorDealKey:
		err = queryAndPushFunc(hub, c, BancorDealKey, "B:"+params[0], count, hub.QueryBancorDeal)
	case RedelegationKey:
		err = queryAndPushFunc(hub, c, RedelegationKey, params[0], count, hub.QueryRedelegation)
	case UnbondingKey:
		err = queryAndPushFunc(hub, c, UnbondingKey, params[0], count, hub.QueryUnbonding)
	case CommentKey:
		err = queryAndPushFunc(hub, c, CommentKey, params[0], count, hub.QueryComment)
	}
	return err
}

func getCount(count int) int {
	count = limitCount(count)
	if count == 0 {
		count = 10
	}
	return count
}

func querySlashAndPush(hub *Hub, c Subscriber, count int) error {
	data, _ := hub.QuerySlash(hub.currBlockTime.Unix(), hub.sid, count)
	bz := groupOfDataPacket(SlashKey, data)
	err := c.WriteMsg(bz)
	if err != nil {
		return err
	}
	return nil
}

func queryKlineAndpush(hub *Hub, c Subscriber, params []string, count int) error {
	tradingPair := params[0]
	if len(params) == 3 {
		tradingPair = strings.Join(params[:2], SeparateArgu)
	}
	candleBz := hub.QueryCandleStick(tradingPair, GetSpanFromSpanStr(params[1]), hub.currBlockTime.Unix(), hub.sid, count)
	bz := groupOfDataPacket(KlineKey, candleBz)
	err := c.WriteMsg(bz)
	if err != nil {
		return err
	}
	return nil
}

func queryDepthAndPush(hub *Hub, c Subscriber, market string, level string, count int) error {
	bz, err := hub.getDepthFullData(market, level, count)
	if err != nil {
		return err
	}
	msg := []byte(fmt.Sprintf("{\"type\":\"%s\", \"payload\":%s}", DepthFull, string(bz)))
	if err := c.WriteMsg(msg); err != nil {
		return err
	}
	return hub.AddLevel(market, level)
}

func queryOrderAndPush(hub *Hub, c Subscriber, account string, count int) error {
	data, tags, _ := hub.QueryOrder(account, hub.currBlockTime.Unix(), hub.sid, count)
	if len(data) != len(tags) {
		return errors.Errorf("The number of orders and tags is not equal")
	}
	createData := make([]json.RawMessage, 0, len(data)/2)
	fillData := make([]json.RawMessage, 0, len(data)/2)
	cancelData := make([]json.RawMessage, 0, len(data)/2)
	for i := len(data) - 1; i >= 0; i-- {
		if tags[i] == CreateOrderEndByte {
			createData = append(createData, data[i])
		} else if tags[i] == FillOrderEndByte {
			fillData = append(fillData, data[i])
		} else if tags[i] == CancelOrderEndByte {
			cancelData = append(cancelData, data[i])
		}
	}
	bz := groupOfDataPacket(CreateOrderKey, createData)
	if err := c.WriteMsg(bz); err != nil {
		return err
	}
	bz = groupOfDataPacket(FillOrderKey, fillData)
	if err := c.WriteMsg(bz); err != nil {
		return err
	}
	bz = groupOfDataPacket(CancelOrderKey, cancelData)
	err := c.WriteMsg(bz)
	return err
}

func queryTickerAndPush(hub *Hub, c Subscriber, market string) error {
	tickers := hub.QueryTickers([]string{market})
	baseData, err := json.Marshal(tickers)
	if err != nil {
		return err
	}
	err = c.WriteMsg([]byte(fmt.Sprintf("{\"type\":\"%s\","+
		" \"payload\":%s}", TickerKey, string(baseData))))
	return err
}

type queryFunc func(string, int64, int64, int) ([]json.RawMessage, []int64)

func queryAndPushFunc(hub *Hub, c Subscriber, typeKey string, param string, count int, qf queryFunc) error {
	data, _ := qf(param, hub.currBlockTime.Unix(), hub.sid, count)
	bz := groupOfDataPacket(typeKey, data)
	err := c.WriteMsg(bz)
	if err != nil {
		return err
	}
	return nil
}
