package server

import (
	"net/http"

	"github.com/coinexchain/trade-server/core"
	"github.com/gorilla/mux"
)

func registerHandler(hub *core.Hub, wsManager *core.WebsocketManager, proxy bool, lcdv0, lcd string, register RegisterRouter) (http.Handler, error) {
	router := mux.NewRouter()

	// REST API Proxy
	if isEnableProxy(register) {
		if proxy {
			if err := registerProxyHandler(lcd, router); err != nil {
				return nil, err
			}
		}
		if proxy && len(lcdv0) != 0 {
			if err := registerProxyHandlerLegacy("/v0", lcdv0, router); err != nil {
				return nil, err
			}
		}
	} else {
		register(router)
	}

	// REST
	router.HandleFunc("/misc/height", QueryLatestHeight(hub)).Methods("GET")
	router.HandleFunc("/misc/block-times", QueryBlockTimesRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/misc/donations", QueryDonationsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/tickers", QueryTickersRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/depths", QueryDepthsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/candle-sticks", QueryCandleSticksRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/user-orders", QueryOrdersRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/deals", QueryDealsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/delist", QueryDelistRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/market/delists", QueryDelistsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/bancorlite/infos", QueryBancorInfosRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/bancorlite/trades", QueryBancorTradesRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/bancorlite/deals", QueryBancorDealsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/expiry/redelegations", QueryRedelegationsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/expiry/unbondings", QueryUnbondingsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/expiry/lockeds", QueryLockedRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/expiry/unlocks", QueryUnlocksRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/tx/incomes", QueryIncomesRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/tx/txs", QueryTxsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/tx/txs/{hash}", QueryTxsByHashRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/comment/comments", QueryCommentsRequestHandlerFn(hub)).Methods("GET")
	router.HandleFunc("/slash/slashings", QuerySlashingsRequestHandlerFn(hub)).Methods("GET")

	// websocket
	router.HandleFunc("/ws", ServeWsHandleFn(wsManager, hub))

	return router, nil
}

func isEnableProxy(register RegisterRouter) bool {
	return register == nil
}
