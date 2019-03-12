package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, reader *http.Request) {
		log.Println(reader.Method, reader.RequestURI)
		next.ServeHTTP(writer, reader)
	})
}

func HealthCheckHandler(writer http.ResponseWriter, reader *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	io.WriteString(writer, "{}")
}

func GatewayTimeoutHandler(writer http.ResponseWriter, reader *http.Request) {
	http.Error(writer, "504 Gateway timeout", http.StatusGatewayTimeout)
}

func main() {
	// Much of this is cribbed from:
	// https://github.com/gorilla/mux#graceful-shutdown
	var wait time.Duration
	var listenAddr string
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.StringVar(&listenAddr, "listen-address", "0.0.0.0:8080", "The address to listen on")
	flag.Parse()

	router := mux.NewRouter()
	// inject some logging middleware so we can log all incoming requets
	router.Use(loggingMiddleware)

	// healthcheck
	router.HandleFunc("/__healthcheck__", HealthCheckHandler)
	// catch all GET method reqeusts and forward to gateway-timeout handler
	router.PathPrefix("/").Methods("GET").HandlerFunc(GatewayTimeoutHandler)

	// set up server
	server := &http.Server{
		Addr:    listenAddr,
		Handler: router,
	}
	// create/start go func for the server so we don't occupy the main thread
	go func() {
		log.Printf("Listening on: %v", listenAddr)
		if err := server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// create a channel for signal notification.
	// shutdown on interrupt, giving some time to allow for closing connections
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	<-signalChannel
	timeoutContext, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	server.Shutdown(timeoutContext)
	log.Println("shutting down")
	os.Exit(0)
}
