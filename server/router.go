package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/coinexchain/trade-server/core"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	queryKeyHeight   = "height"
	queryKeyTimespan = "timespan"
	queryKeyTime     = "time"
	queryKeySid      = "sid"
	queryKeyCount    = "count"
	queryKeyMarket   = "market"
	queryKeyAccount  = "account"
	queryKeyToken    = "token"
)

const (
	Subscribe   = "subscribe"
	Unsubscribe = "unsubscribe"
	Ping        = "ping"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type OpCommand struct {
	Op   string   `json:"op"`
	Args []string `json:"args"`
}

func registerHandler(hub *core.Hub, wsManager *core.WebsocketManager) http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/misc/block-times", QueryBlockTimesRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/tickers", QueryTickersRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/depths", QueryDepthsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/candle-sticks", QueryCandleSticksRequestHandlerFn(hub)).Methods("GET")

	router.HandleFunc("/market/orders", QueryOrdersRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/deals", QueryDealsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/bancorlite/infos", QueryBancorInfosRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/bancorlite/trades", QueryBancorTradesRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/expiry/redelegations", QueryRedelegationsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/expiry/unbondings", QueryUnbondingsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/expiry/unlocks", QueryUnlocksRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/tx/incomes", QueryIncomesRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/tx/txs", QueryTxsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/comment/comments", QueryCommentsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/slash/slashings", QuerySlashingsRequestHandlerFn(hub)).Methods("GET")

	router.HandleFunc("/ws", ServeWsHandleFn(wsManager))

	return router
}

func ServeWsHandleFn(wsManager *core.WebsocketManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logrus.Error(err)
			return
		}
		wsConn := core.NewConn(c)
		wsManager.AddConn(wsConn)

		go func() {
			for {
				_, message, err := wsConn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {

					}
					// TODO. will handle the error
					_ = wsManager.CloseConn(wsConn)
					break
				}

				// TODO. will handler the error
				var command OpCommand
				if err := json.Unmarshal(message, &command); err != nil {
					logrus.Error(err)
					continue
				}

				// TODO. will handler the error
				switch command.Op {
				case Subscribe:
					for _, subTopic := range command.Args {
						_ = wsManager.AddSubscribeConn(subTopic, wsConn)
					}
				case Unsubscribe:
					for _, subTopic := range command.Args {
						_ = wsManager.RemoveSubscribeConn(subTopic, wsConn)
					}
				case Ping:
					_ = wsConn.PongHandler()(`{"type": "pong"}`)
				}

			}
		}()

	}
}

// QueryBlockTimesRequestHandlerFn - func
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

// QueryTickersRequestHandlerFn - func
func QueryTickersRequestHandlerFn(hub *core.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// todo
	}
}

// QueryDepthsRequestHandlerFn - func
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

// QueryCandleSticksRequestHandlerFn - func
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

		postQueryResponse(w, hub.QueryCandleStick(market, timespan, time, sid, count))
	}
}

// QueryOrdersRequestHandlerFn - func
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

		data, timesid := hub.QueryDeal(account, time, sid, count)

		postQueryKVStoreResponse(w, data, timesid)
	}
}

// QueryDealsRequestHandlerFn - func
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

// QueryBancorInfosRequestHandlerFn - func
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

// QueryBancorTradesRequestHandlerFn - func
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

		postQueryKVStoreResponse(w, data, timesid)
	}
}

// QueryRedelegationsRequestHandlerFn - func
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

// QueryUnbondingsRequestHandlerFn - func
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

// QueryUnlocksRequestHandlerFn - func
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

		postQueryKVStoreResponse(w, data, timesid)
	}
}

// QueryIncomesRequestHandlerFn - func
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

		postQueryKVStoreResponse(w, data, timesid)
	}
}

// QueryTxsRequestHandlerFn - func
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

		postQueryKVStoreResponse(w, data, timesid)
	}
}

// QueryCommentsRequestHandlerFn - func
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

// QuerySlashingsRequestHandlerFn - func
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

func postQueryKVStoreResponse(w http.ResponseWriter, data [][]byte, timesid []int64) {

	baseData, err := json.Marshal(data)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	wrappedData := NewDataWrapped(baseData, timesid)
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
	Data    json.RawMessage `json:"data"`
	TimeSid []int64         `json:"timesid"`
}

func NewDataWrapped(data json.RawMessage, timesid []int64) DataWrapped {
	return DataWrapped{
		Data:    data,
		TimeSid: timesid,
	}
}

type DepthResponse struct {
	sell []*core.PricePoint
	buy  []*core.PricePoint
}

func NewDepthResponse(sell []*core.PricePoint, buy []*core.PricePoint) DepthResponse {
	return DepthResponse{
		sell: sell,
		buy:  buy,
	}
}
