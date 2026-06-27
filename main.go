package main

import (
	"log"
	"net/http"

	"github.com/snakesneaks/webhook-relay/cfg"
	"github.com/snakesneaks/webhook-relay/handler"
	"github.com/snakesneaks/webhook-relay/service"
)

func main() {
	cfg, err := cfg.Load()
	if err != nil {
		log.Fatal(err)
	}

	renderService := service.NewRenderService()

	mux := http.NewServeMux()

	// health endpoints
	mux.Handle("/healthz", handler.NewHealthHandler())

	// relay handler
	mux.Handle("/", handler.NewRelayHandler(cfg, renderService))

	log.Println("Listen and serve on port: ", cfg.Server.Addr)
	log.Println("Routes: ", cfg.App.Routes)
	if err := http.ListenAndServe(cfg.Server.Addr, mux); err != nil {
		log.Fatal(err)
	}
}
