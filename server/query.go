package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

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
	queryKeyOrderTag   = "tag"
)

func QueryLatestHeight(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postQueryResponse(w, hub.QueryBlockInfo())
	}
}

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

		token := r.FormValue(queryKeyToken)
		data, timesid := hub.QueryLockedAboutToken(strings.ToLower(token), account, time, sid, count)

		postQueryKVStoreResponse(w, data, timesid)
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

		postQueryResponse(w, data)
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

		token := r.FormValue(queryKeyToken)
		tag := r.FormValue(queryKeyOrderTag)
		if err := parseQueryTagParams(tag); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		data, tags, timesid := hub.QueryOrderAboutToken(tag, strings.ToLower(token), account, time, sid, count)

		createOrders := make([]json.RawMessage, 0)
		fillOrders := make([]json.RawMessage, 0)
		cancelOrders := make([]json.RawMessage, 0)
		createTimeSid := make([]int64, 0)
		fillTimeSid := make([]int64, 0)
		cancelTimeSid := make([]int64, 0)
		for i, tag := range tags {
			if tag == core.CreateOrderEndByte {
				createOrders = append(createOrders, data[i])
				createTimeSid = append(createTimeSid, timesid[i*2])
				createTimeSid = append(createTimeSid, timesid[i*2+1])
			}
			if tag == core.FillOrderEndByte {
				fillOrders = append(fillOrders, data[i])
				fillTimeSid = append(fillTimeSid, timesid[i*2])
				fillTimeSid = append(fillTimeSid, timesid[i*2+1])
			}
			if tag == core.CancelOrderEndByte {
				cancelOrders = append(cancelOrders, data[i])
				cancelTimeSid = append(cancelTimeSid, timesid[i*2])
				cancelTimeSid = append(cancelTimeSid, timesid[i*2+1])
			}
		}

		var orders interface{}
		switch tag {
		case core.CreateOrderStr:
			orders = core.OrderResponse{Data: createOrders, Timesid: createTimeSid}
		case core.FillOrderStr:
			orders = core.OrderResponse{Data: fillOrders, Timesid: fillTimeSid}
		case core.CancelOrderStr:
			orders = core.OrderResponse{Data: cancelOrders, Timesid: cancelTimeSid}
		case "":
			orders = core.OrderInfo{
				CreateOrderInfo: core.OrderResponse{Data: createOrders, Timesid: createTimeSid},
				FillOrderInfo:   core.OrderResponse{Data: fillOrders, Timesid: fillTimeSid},
				CancelOrderInfo: core.OrderResponse{Data: cancelOrders, Timesid: cancelTimeSid},
			}
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

		postQueryKVStoreResponse(w, data, timesid)
	}
}

func QueryBancorDealsRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
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

		data, timesid := hub.QueryBancorDeal(market, time, sid, count)

		postQueryKVStoreResponse(w, data, timesid)
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

		postQueryKVStoreResponse(w, data, timesid)
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

		token := r.FormValue(queryKeyToken)
		data, timesid := hub.QueryBancorTradeAboutToken(strings.ToLower(token), account, time, sid, count)

		postQueryKVStoreResponse(w, data, timesid)
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

		postQueryKVStoreResponse(w, data, timesid)
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

		postQueryKVStoreResponse(w, data, timesid)
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

		token := r.FormValue(queryKeyToken)
		data, timesid := hub.QueryUnlockAboutToken(strings.ToLower(token), account, time, sid, count)

		postQueryKVStoreResponse(w, data, timesid)
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

		token := r.FormValue(queryKeyToken)
		data, timesid := hub.QueryIncomeAboutToken(strings.ToLower(token), account, time, sid, count)

		postQueryKVStoreResponse(w, data, timesid)
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

		token := r.FormValue(queryKeyToken)
		data, timesid := hub.QueryTxAboutToken(strings.ToLower(token), account, time, sid, count)

		postQueryKVStoreResponse(w, data, timesid)
	}
}
func QueryTxsByHashRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		hashHexStr := vars["hash"]
		data := hub.QueryTxByHashID(hashHexStr)
		postQueryResponse(w, data)
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

		postQueryKVStoreResponse(w, data, timesid)
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

		postQueryKVStoreResponse(w, data, timesid)
	}
}

func QueryDonationsRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
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

		data, timesid := hub.QueryDonation(time, sid, count)

		postQueryKVStoreResponse(w, data, timesid)
	}
}

func QueryDelistRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
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

		data, timesid := hub.QueryDelist(market, time, sid, count)
		var cancelTime int64
		if len(data) > 0 {
			cancelTime = core.BigEndianBytesToInt64(data[0])
			postQueryKVStoreResponse(w, []int64{cancelTime}, timesid)
		} else {
			postQueryKVStoreResponse(w, []struct{}{}, timesid)
		}
	}
}
func QueryDelistsRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
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

		data, timesid := hub.QueryDelists(time, sid, count)
		type CancelTradingPair struct {
			TradingPair string `json:"trading_pair"`
			CancelTime  int64  `json:"cancel_time"`
		}
		vals := make([]CancelTradingPair, len(data)/2)
		for i := 0; i < len(data); i += 2 {
			vals[i/2].TradingPair = string(data[i])
			vals[i/2].CancelTime = core.BigEndianBytesToInt64(data[i+1])
		}
		postQueryKVStoreResponse(w, vals, timesid)
	}
}

func postQueryResponse(w http.ResponseWriter, data interface{}) {
	var (
		err      error
		baseData []byte
	)
	switch data := data.(type) {
	case []byte:
		baseData = data
	default:
		if baseData, err = json.Marshal(data); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	writeData(w, baseData)
}

func writeData(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(data); err != nil {
		log.WithError(err).Error("write response failed")
		return
	}
}

func postQueryKVStoreResponse(w http.ResponseWriter, data interface{}, timesid []int64) {
	wrappedData := NewDataWrapped(data, timesid)
	output, err := json.Marshal(wrappedData)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeData(w, output)
}

func parseQueryHeightParams(str string) (height int64, err error) {
	if str == "" {
		return height, ErrNilParams(queryKeyHeight)
	}
	if height, err = strconv.ParseInt(str, 10, 64); err != nil {
		return
	}
	if height < 0 {
		return height, ErrNegativeParams(queryKeyHeight)
	}
	return
}

func parseQueryCountParams(str string) (count int, err error) {
	if str == "" {
		return count, ErrNilParams(queryKeyCount)
	}
	if count, err = strconv.Atoi(str); err != nil {
		return count, err
	}
	if count <= 0 {
		return count, ErrInvalidParams(queryKeyCount)
	}
	return
}

func parseQueryTimespanParams(str string) (timespan byte, err error) {
	if str == "" {
		return timespan, ErrNilParams(queryKeyTimespan)
	}
	if timespan = core.GetSpanFromSpanStr(str); timespan == 0 {
		return 0, ErrInvalidTimespan()
	}
	return timespan, nil
}

func parseQueryTagParams(str string) error {
	if str != "" && str != core.CreateOrderStr && str != core.FillOrderStr && str != core.CancelOrderStr {
		return ErrInvalidTag()
	}
	return nil
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

	sid, err = strconv.ParseInt(sidStr, 10, 64)
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
