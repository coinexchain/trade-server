package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

const (
	ReadTimeout  = 10
	WriteTimeout = 10
	WaitTimeout  = 10
)

var (
	help bool
	port int
)

func init() {
	flag.BoolVar(&help, "h", false, "this help")
	flag.IntVar(&port, "p", 8000, "HTTP listen port")
	flag.Usage = usage
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	router := registerHandler()

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: WriteTimeout * time.Second,
		ReadTimeout:  ReadTimeout * time.Second,
	}

	log.Printf("Server start... (port: %v)\n", port)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	waitForSignal()

	ctx, cancel := context.WithTimeout(context.Background(), WaitTimeout * time.Second)
	defer cancel()

	_ = srv.Shutdown(ctx)

	log.Println("Server end...")
}

func waitForSignal()  {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func usage() {
	_, _ = fmt.Println(`Options:`)
	flag.PrintDefaults()
}

func registerHandler() http.Handler {
	router := mux.NewRouter()
	subRouter := router.PathPrefix("/v1").Subrouter()
	subRouter.HandleFunc("/test", TestHandler).Queries("filter", "{filter}").Queries("limit", "{limit}")
	subRouter.HandleFunc("/test/{key:[0-9]+}", TestKeyHandler)
	return router
}

func TestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "TestFilter: %v %v\n", vars["filter"], vars["limit"])
}

func TestKeyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "TestKey: %v\n", vars["key"])
}
