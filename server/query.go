package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/coinexchain/trade-server/core"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

const (
	queryKeyHeight     = "height"
	queryKeyTimespan   = "timespan"
	queryKeyTime       = "time"
	queryKeySid        = "sid"
	queryKeyCount      = "count"
	queryKeyMarket     = "market"
	queryKeyAccount    = "account"
	queryKeyToken      = "token"
	queryKeyMarketList = "market_list"
)

func QueryBlockTimesRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		height, err := parseQueryHeightParams(r.FormValue(queryKeyHeight))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		count, err := parseQueryCountParams(r.FormValue(queryKeyCount))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		postQueryResponse(w, hub.QueryBlockTime(height, count))
	}
}

func QueryTickersRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		marketStr := vars.Get(queryKeyMarketList)

		Tickers := hub.QueryTickers(strings.Split(marketStr, ","))
		postQueryResponse(w, Tickers)
	}
}

func QueryDepthsRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}
		market := r.FormValue(queryKeyMarket)
		count, err := parseQueryCountParams(r.FormValue(queryKeyCount))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		sell, buy := hub.QueryDepth(market, count)
		postQueryResponse(w, NewDepthResponse(sell, buy))

	}
}

func QueryLockedRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		account := r.FormValue(queryKeyAccount)
		time, sid, count, err := parseQueryKVStoreParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		data, timesid := hub.QueryComment(account, time, sid, count)

		var msg core.LockedSendMsg
		msgs := make([]core.LockedSendMsg, 0)
		for _, v := range data {
			if err = json.Unmarshal(v, &msg); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			msgs = append(msgs, msg)
		}

		postQueryKVStoreResponse(w, msgs, timesid)
	}
}

func QueryCandleSticksRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		market := r.FormValue(queryKeyMarket)
		timespan, err := parseQueryTimespanParams(r.FormValue(queryKeyTimespan))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		time, sid, count, err := parseQueryKVStoreParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		data := hub.QueryCandleStick(market, timespan, time, sid, count)

		var stick core.CandleStick
		sticks := make([]core.CandleStick, 0)
		for _, v := range data {
			if err = json.Unmarshal(v, &stick); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			sticks = append(sticks, stick)
		}

		postQueryResponse(w, sticks)
	}
}

func QueryOrdersRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		account := r.FormValue(queryKeyAccount)
		time, sid, count, err := parseQueryKVStoreParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		data, tags, timesid := hub.QueryOrder(account, time, sid, count)
		createOrders := make([]core.CreateOrderInfo, 0)
		fillOrders := make([]core.FillOrderInfo, 0)
		cancelOrders := make([]core.CancelOrderInfo, 0)
		createTimeSid := make([]int64, 0)
		fillTimeSid := make([]int64, 0)
		cancelTimeSid := make([]int64, 0)
		for i, tag := range tags {
			if tag == core.CreateOrderEndByte {
				var order core.CreateOrderInfo
				if err = json.Unmarshal(data[i], &order); err != nil {
					rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
					return
				}
				createOrders = append(createOrders, order)
				createTimeSid = append(createTimeSid, timesid[i*2])
				createTimeSid = append(createTimeSid, timesid[i*2+1])
			}
			if tag == core.FillOrderEndByte {
				var order core.FillOrderInfo
				if err = json.Unmarshal(data[i], &order); err != nil {
					rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
					return
				}
				fillOrders = append(fillOrders, order)
				fillTimeSid = append(fillTimeSid, timesid[i*2])
				fillTimeSid = append(fillTimeSid, timesid[i*2+1])
			}
			if tag == core.CancelOrderEndByte {
				var order core.CancelOrderInfo
				if err = json.Unmarshal(data[i], &order); err != nil {
					rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
					return
				}
				cancelOrders = append(cancelOrders, order)
				cancelTimeSid = append(cancelTimeSid, timesid[i*2])
				cancelTimeSid = append(cancelTimeSid, timesid[i*2+1])
			}
		}
		orders := core.OrderInfo{
			CreateOrderInfo: core.CreateOrderResponse{Data: createOrders, Timesid: createTimeSid},
			FillOrderInfo:   core.FillOrderResponse{Data: fillOrders, Timesid: fillTimeSid},
			CancelOrderInfo: core.CancelOrderResponse{Data: cancelOrders, Timesid: cancelTimeSid},
		}

		postQueryResponse(w, orders)
	}
}

func QueryDealsRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		market := r.FormValue(queryKeyMarket)
		time, sid, count, err := parseQueryKVStoreParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		data, timesid := hub.QueryDeal(market, time, sid, count)

		var deal core.FillOrderInfo
		deals := make([]core.FillOrderInfo, 0)
		for _, v := range data {
			if err = json.Unmarshal(v, &deal); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			deals = append(deals, deal)
		}

		postQueryKVStoreResponse(w, deals, timesid)
	}
}

func QueryBancorInfosRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		market := r.FormValue(queryKeyMarket)
		time, sid, count, err := parseQueryKVStoreParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		data, timesid := hub.QueryBancorInfo(market, time, sid, count)

		var info core.MsgBancorInfoForKafka
		infos := make([]core.MsgBancorInfoForKafka, 0)
		for _, v := range data {
			if err = json.Unmarshal(v, &info); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			infos = append(infos, info)
		}

		postQueryKVStoreResponse(w, infos, timesid)
	}
}

func QueryBancorTradesRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		account := r.FormValue(queryKeyAccount)
		time, sid, count, err := parseQueryKVStoreParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		data, timesid := hub.QueryBancorTrade(account, time, sid, count)

		var trade core.MsgBancorTradeInfoForKafka
		trades := make([]core.MsgBancorTradeInfoForKafka, 0)
		for _, v := range data {
			if err = json.Unmarshal(v, &trade); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			trades = append(trades, trade)
		}

		postQueryKVStoreResponse(w, trades, timesid)
	}
}

func QueryRedelegationsRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		account := r.FormValue(queryKeyAccount)
		time, sid, count, err := parseQueryKVStoreParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		data, timesid := hub.QueryRedelegation(account, time, sid, count)

		var redelegation core.NotificationBeginRedelegation
		redelegations := make([]core.NotificationBeginRedelegation, 0)
		for _, v := range data {
			if err = json.Unmarshal(v, &redelegation); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			redelegations = append(redelegations, redelegation)
		}

		postQueryKVStoreResponse(w, redelegations, timesid)
	}
}

func QueryUnbondingsRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		account := r.FormValue(queryKeyAccount)
		time, sid, count, err := parseQueryKVStoreParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		data, timesid := hub.QueryUnbonding(account, time, sid, count)

		var unbonding core.NotificationBeginUnbonding
		unbondings := make([]core.NotificationBeginUnbonding, 0)
		for _, v := range data {
			if err = json.Unmarshal(v, &unbonding); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			unbondings = append(unbondings, unbonding)
		}

		postQueryKVStoreResponse(w, unbondings, timesid)
	}
}

func QueryUnlocksRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		account := r.FormValue(queryKeyAccount)
		time, sid, count, err := parseQueryKVStoreParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		data, timesid := hub.QueryUnlock(account, time, sid, count)

		var unLock core.NotificationUnlock
		unLocks := make([]core.NotificationUnlock, 0)
		for _, v := range data {
			if err = json.Unmarshal(v, &unLock); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			unLocks = append(unLocks, unLock)
		}

		postQueryKVStoreResponse(w, unLocks, timesid)
	}
}

func QueryIncomesRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		account := r.FormValue(queryKeyAccount)
		time, sid, count, err := parseQueryKVStoreParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		data, timesid := hub.QueryIncome(account, time, sid, count)

		var tx core.NotificationTx
		txs := make([]core.NotificationTx, 0)
		for _, v := range data {
			if err = json.Unmarshal(v, &tx); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			txs = append(txs, tx)
		}

		postQueryKVStoreResponse(w, txs, timesid)
	}
}

func QueryTxsRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		account := r.FormValue(queryKeyAccount)
		time, sid, count, err := parseQueryKVStoreParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		data, timesid := hub.QueryTx(account, time, sid, count)

		var tx core.NotificationTx
		txs := make([]core.NotificationTx, 0)
		for _, v := range data {
			if err = json.Unmarshal(v, &tx); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			txs = append(txs, tx)
		}

		postQueryKVStoreResponse(w, txs, timesid)
	}
}

func QueryCommentsRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		token := r.FormValue(queryKeyToken)
		time, sid, count, err := parseQueryKVStoreParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		data, timesid := hub.QueryComment(token, time, sid, count)

		var comment core.TokenComment
		comments := make([]core.TokenComment, 0)
		for _, v := range data {
			if err = json.Unmarshal(v, &comment); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			comments = append(comments, comment)
		}

		postQueryKVStoreResponse(w, comments, timesid)
	}
}

func QuerySlashingsRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest,
				sdk.AppendMsgToErr("could not parse query parameters", err.Error()))
			return
		}

		time, sid, count, err := parseQueryKVStoreParams(r)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		data, timesid := hub.QuerySlash(time, sid, count)

		var slashing core.NotificationSlash
		slashings := make([]core.NotificationSlash, 0)
		for _, v := range data {
			if err = json.Unmarshal(v, &slashing); err != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
				return
			}
			slashings = append(slashings, slashing)
		}

		postQueryKVStoreResponse(w, slashings, timesid)
	}
}
func postQueryResponse(w http.ResponseWriter, data interface{}) {
	var (
		baseData []byte
		err      error
	)

	switch data.(type) {
	case []byte:
		baseData = data.([]byte)

	default:
		baseData, err = json.Marshal(data)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(baseData)
}

func postQueryKVStoreResponse(w http.ResponseWriter, data interface{}, timesid []int64) {
	wrappedData := NewDataWrapped(data, timesid)
	output, err := json.Marshal(wrappedData)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(output)
}

func parseQueryHeightParams(str string) (height int64, err error) {

	if str == "" {
		return height, ErrNilParams(queryKeyHeight)
	}

	height, err = strconv.ParseInt(str, 10, 64)
	if err != nil {
		return height, err
	} else if height < 0 {
		return height, ErrNegativeParams(queryKeyHeight)
	}

	return
}

func parseQueryCountParams(str string) (count int, err error) {

	if str == "" {
		return count, ErrNilParams(queryKeyCount)
	}

	count, err = strconv.Atoi(str)
	if err != nil {
		return count, err
	} else if count <= 0 {
		return count, ErrInvalidParams(queryKeyCount)
	}

	return
}

func parseQueryTimespanParams(str string) (byte, error) {
	var (
		timespan int64
		err      error
	)
	if str == "" {
		return 0, ErrNilParams(queryKeyTimespan)
	}

	timespan, err = strconv.ParseInt(str, 10, 8)
	if err != nil {
		return 0, err
	}
	if byte(timespan) != core.Minute && byte(timespan) != core.Hour && byte(timespan) != core.Day {
		return 0, ErrInvalidTimespan()
	}

	return byte(timespan), nil
}

func parseQueryKVStoreParams(r *http.Request) (time int64, sid int64, count int, err error) {

	timeStr := r.FormValue(queryKeyTime)
	if timeStr == "" {
		return time, sid, count, ErrNilParams(queryKeyTime)
	}

	time, err = strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return time, sid, count, err
	} else if time <= 0 {
		return time, sid, count, ErrInvalidParams(queryKeyTime)
	}

	sidStr := r.FormValue(queryKeySid)
	if sidStr == "" {
		return time, sid, count, ErrNilParams(queryKeySid)
	}

	sid, err = strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return time, sid, count, err
	} else if sid < 0 {
		return time, sid, count, ErrNegativeParams(queryKeySid)
	}

	countStr := r.FormValue(queryKeyCount)
	if countStr == "" {
		return time, sid, count, ErrNilParams(queryKeyCount)
	}

	count, err = strconv.Atoi(countStr)
	if err != nil {
		return time, sid, count, err
	} else if count <= 0 {
		return time, sid, count, ErrInvalidParams(queryKeyCount)
	}

	return
}

func ErrNilParams(params string) error {
	return fmt.Errorf("%s can not be nil", params)
}
func ErrNegativeParams(params string) error {
	return fmt.Errorf("%s cannot be negative", params)
}
func ErrInvalidParams(params string) error {
	return fmt.Errorf("%s must greater than 0", params)
}
func ErrInvalidTimespan() error {
	return fmt.Errorf("timespan must be Minute:16/Hour:32/Day:48")
}

type DataWrapped struct {
	Data    interface{} `json:"data"`
	TimeSid []int64     `json:"timesid"`
}

func NewDataWrapped(data interface{}, timesid []int64) DataWrapped {
	return DataWrapped{
		Data:    data,
		TimeSid: timesid,
	}
}

type DepthResponse struct {
	Sell []*core.PricePoint `json:"sell"`
	Buy  []*core.PricePoint `json:"buy"`
}

func NewDepthResponse(sell []*core.PricePoint, buy []*core.PricePoint) DepthResponse {
	return DepthResponse{
		Sell: sell,
		Buy:  buy,
	}
}
