package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func registerHandler() http.Handler {
	router := mux.NewRouter()
	subRouter := router.PathPrefix("/v1").Subrouter()
	subRouter.HandleFunc("/test", testHandler).Queries("filter", "{filter}").Queries("limit", "{limit}")
	subRouter.HandleFunc("/test/{key:[0-9]+}", testKeyHandler)
	return router
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "TestFilter: %v %v\n", vars["filter"], vars["limit"])
}

func testKeyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "TestKey: %v\n", vars["key"])
}