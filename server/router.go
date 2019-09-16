package server

import (
	"net/http"

	"github.com/coinexchain/trade-server/core"
	"github.com/gorilla/mux"
)

func registerHandler(hub *core.Hub, wsManager *core.WebsocketManager, proxy bool, lcd string) (http.Handler, error) {
	router := mux.NewRouter()

	// REST API Proxy
	if proxy {
		if err := registerProxyHandler(lcd, router); err != nil {
			return nil, err
		}
	}

	// REST
	router.HandleFunc("/misc/block-times", QueryBlockTimesRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/misc/donations", QueryDonationsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/tickers", QueryTickersRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/depths", QueryDepthsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/candle-sticks", QueryCandleSticksRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/user-orders", QueryOrdersRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/deals", QueryDealsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/bancorlite/infos", QueryBancorInfosRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/bancorlite/trades", QueryBancorTradesRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/expiry/redelegations", QueryRedelegationsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/expiry/unbondings", QueryUnbondingsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/expiry/lockeds", QueryLockedRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/expiry/unlocks", QueryUnlocksRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/tx/incomes", QueryIncomesRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/tx/txs", QueryTxsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/comment/comments", QueryCommentsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/slash/slashings", QuerySlashingsRequestHandlerFn(hub)).Methods("GET")

	// websocket
	router.HandleFunc("/ws", ServeWsHandleFn(wsManager, hub))

	return router, nil
}
